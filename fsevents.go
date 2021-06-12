// +build darwin

package rnotify

import (
	"log"
	"math"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsevents"
	"github.com/fsnotify/fsnotify"
)

const unsupportedOp = math.MaxUint32

// Watcher watches files and directories, delivering events to a channel.
type Watcher struct {
	Events  chan fsnotify.Event
	Errors  chan error
	mu      sync.Mutex
	watches map[string]*fsevents.EventStream
}

// NewWatcher builds a new watcher.
func NewWatcher() (*Watcher, error) {
	w := &Watcher{
		Events:  make(chan fsnotify.Event),
		Errors:  make(chan error),
		watches: make(map[string]*fsevents.EventStream),
	}
	return w, nil
}

// Add starts watching the directory (recursively).
func (w *Watcher) Add(name string) error {
	absPath, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}

	dev, err := fsevents.DeviceForPath(absPath)
	if err != nil {
		return err
	}

	es := &fsevents.EventStream{
		Paths:   []string{absPath},
		Latency: 500 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents}

	w.mu.Lock()
	w.watches[name] = es
	w.mu.Unlock()

	es.Start()
	ec := es.Events

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

// Close stops watching.
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.watches) == 0 {
		return nil
	}

	for path, es := range w.watches {
		es.Stop()
		delete(w.watches, path)
	}

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

//nolint:unused
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

//nolint:unused
func (w *Watcher) logEvent(event fsevents.Event) {
	note := ""
	for bit, description := range noteDescription {
		if event.Flags&bit == bit {
			note += description + " "
		}
	}
	log.Printf("EventID: %d Path: %s Flags: %s", event.ID, event.Path, note)
}
