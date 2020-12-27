// +build linux

package rnotify

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Events    chan fsnotify.Event
	Errors    chan error
	fswatcher *fsnotify.Watcher
}

func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fswatcher: watcher,
		Events:    make(chan fsnotify.Event),
		Errors:    make(chan error),
	}

	go w.readEvents()
	return w, nil
}

func (w *Watcher) Close() error {
	return w.fswatcher.Close()
}

func (w *Watcher) Add(name string) error {
	err := filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err = w.fswatcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})

	return err
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
						if err = w.fswatcher.Add(event.Name); err != nil {
							w.Errors <- err
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
