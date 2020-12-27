// +build windows

package rnotify

import (
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

	watcher.SetRecursive()

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
	return w.fswatcher.Add(name)
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
