// +build darwin

package rnotify

import (
	"log"
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
					// FIXME(y-yagi) Is '/' really need?
					e := fsnotify.Event{Name: "/" + event.Path, Op: op}
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
	if event&fsevents.ItemRemoved == fsevents.ItemRemoved {
		return fsnotify.Remove
	}

	if event&fsevents.ItemRenamed == fsevents.ItemRenamed {
		return fsnotify.Rename
	}

	if event&fsevents.ItemChangeOwner == fsevents.ItemChangeOwner {
		return fsnotify.Chmod
	}

	if event&fsevents.ItemModified == fsevents.ItemModified {
		if event&fsevents.ItemCreated == fsevents.ItemCreated {
			return fsnotify.Create
		}
		return fsnotify.Write
	}

	if event&fsevents.ItemCreated == fsevents.ItemCreated {
		return fsnotify.Create
	}

	return unsupportedOp
}

var noteDescription = map[fsevents.EventFlags]string{
	fsevents.MustScanSubDirs: "MustScanSubdirs",
	fsevents.UserDropped:     "UserDropped",
	fsevents.KernelDropped:   "KernelDropped",
	fsevents.EventIDsWrapped: "EventIDsWrapped",
	fsevents.HistoryDone:     "HistoryDone",
	fsevents.RootChanged:     "RootChanged",
	fsevents.Mount:           "Mount",
	fsevents.Unmount:         "Unmount",

	fsevents.ItemCreated:       "Created",
	fsevents.ItemRemoved:       "Removed",
	fsevents.ItemInodeMetaMod:  "InodeMetaMod",
	fsevents.ItemRenamed:       "Renamed",
	fsevents.ItemModified:      "Modified",
	fsevents.ItemFinderInfoMod: "FinderInfoMod",
	fsevents.ItemChangeOwner:   "ChangeOwner",
	fsevents.ItemXattrMod:      "XAttrMod",
	fsevents.ItemIsFile:        "IsFile",
	fsevents.ItemIsDir:         "IsDir",
	fsevents.ItemIsSymlink:     "IsSymLink",
}

func (w *Watcher) logEvent(event fsevents.Event) {
	note := ""
	for bit, description := range noteDescription {
		if event.Flags&bit == bit {
			note += description + " "
		}
	}
	log.Printf("EventID: %d Path: %s Flags: %s", event.ID, event.Path, note)
}
