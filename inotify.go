// +build linux

package rnotify

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches files and directories, delivering events to a channel.
type Watcher struct {
	Events    chan fsnotify.Event
	Errors    chan error
	fswatcher *fsnotify.Watcher
	ignore    map[string]struct{}
}

// NewWatcher builds a new watcher.
func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fswatcher: watcher,
		Events:    make(chan fsnotify.Event),
		Errors:    make(chan error),
		ignore:    map[string]struct{}{},
	}

	go w.readEvents()
	return w, nil
}

// Add starts watching the directory (recursively).
func (w *Watcher) Add(name string) error {
	err := filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for ignorePath := range w.ignore {
				if strings.Contains(path, ignorePath) {
					return nil
				}
			}

			if err = w.fswatcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// Close stops watching.
func (w *Watcher) Close() error {
	return w.fswatcher.Close()
}

// Ignore specifies directories to ignore.
func (w *Watcher) Ignore(paths []string) {
	for _, path := range paths {
		w.ignore[path] = struct{}{}
	}
}

func (w *Watcher) readEvents() {
	defer close(w.Errors)
	defer close(w.Events)

	for {
		select {
		case event, ok := <-w.fswatcher.Events:
			if ok {
				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err != nil {
						w.Errors <- err
					} else if info.IsDir() {
						skip := false
						for ignorePath := range w.ignore {
							if strings.Contains(event.Name, ignorePath) {
								skip = true
								break
							}
						}
						if !skip {
							if err = w.fswatcher.Add(event.Name); err != nil {
								w.Errors <- err
							}
						}
					}
				}
				w.Events <- event
			}
		case err, ok := <-w.fswatcher.Errors:
			if ok {
				w.Errors <- err
			}
		}
	}
}
