// +build windows

package rnotify

import (
	"github.com/fsnotify/fsnotify"
)

// Watcher watches files and directories, delivering events to a channel.
type Watcher struct {
	Events    chan fsnotify.Event
	Errors    chan error
	fswatcher *fsnotify.Watcher
}

// NewWatcher builds a new watcher.
func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watcher.SetRecursive()

	w := &Watcher{
		fswatcher: watcher,
		Events:    make(chan fsnotify.Event),
		Errors:    make(chan error),
	}

	go w.readEvents()
	return w, nil
}

// Add starts watching the directory (recursively).
func (w *Watcher) Add(name string) error {
	return w.fswatcher.Add(name)
}

// Close stops watching.
func (w *Watcher) Close() error {
	return w.fswatcher.Close()
}

func (w *Watcher) readEvents() {
	for {
		select {
		case event, ok := <-w.fswatcher.Events:
			if ok {
				w.Events <- event
			}
		case err, ok := <-w.fswatcher.Errors:
			if ok {
				w.Errors <- err
			}
		}
	}
}
