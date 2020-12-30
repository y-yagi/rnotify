// +build darwin

package rnotify

import (
	"math"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsevents"
	"github.com/fsnotify/fsnotify"
)

const unsupportedOp = math.MaxUint32

type Watcher struct {
	es     *fsevents.EventStream
	Events chan fsnotify.Event
	Errors chan error
}

func NewWatcher() (*Watcher, error) {
	w := &Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}
	return w, nil
}

func (w *Watcher) Add(name string) error {
	absPath, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}

	dev, err := fsevents.DeviceForPath(absPath)
	if err != nil {
		return err
	}

	w.es = &fsevents.EventStream{
		Paths:   []string{absPath},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents}

	w.es.Start()
	ec := w.es.Events

	go func() {
		for msg := range ec {
			for _, event := range msg {
				op := w.translateEvent(event.Flags)
				if op != unsupportedOp {
					e := fsnotify.Event{Name: event.Path, Op: op}
					w.Events <- e
				}
			}
		}
	}()

	return nil
}

func (w *Watcher) Close() error {
	if w.es == nil {
		return nil
	}

	w.es.Stop()
	return nil
}

func (w *Watcher) translateEvent(event fsevents.EventFlags) fsnotify.Op {
	if event&fsevents.ItemCreated == fsevents.ItemCreated {
		return fsnotify.Create
	} else if event&fsevents.ItemModified == fsevents.ItemModified {
		return fsnotify.Write
	} else if event&fsevents.ItemRemoved == fsevents.ItemRemoved {
		return fsnotify.Remove
	} else if event&fsevents.ItemRemoved == fsevents.ItemRemoved {
		return fsnotify.Rename
	} else if event&fsevents.ItemChangeOwner == fsevents.ItemChangeOwner {
		return fsnotify.Chmod
	}
	return unsupportedOp
}
