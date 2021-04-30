package rnotify_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/y-yagi/rnotify"
)

func TestWatchDirRecursive(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "rnotify_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Mkdir(filepath.Join(tempDir, "foo"), 0777); err != nil {
		t.Fatal(err)
	}

	if tempDir, err = filepath.EvalSymlinks(tempDir); err != nil {
		t.Fatal(err)
	}

	watcher, _ := rnotify.NewWatcher()
	defer watcher.Close()

	done := make(chan bool)

	fileA := filepath.Join(tempDir, "a")
	dirB := filepath.Join(tempDir, "b")
	fileC := filepath.Join(tempDir, "foo", "c")
	dirD := filepath.Join(tempDir, "foo", "d")
	fileE := filepath.Join(tempDir, "foo", "d", "e")
	events := map[string]fsnotify.Op{
		fileA: fsnotify.Create,
		dirB:  fsnotify.Create,
		fileC: fsnotify.Create,
		dirD:  fsnotify.Create,
		fileE: fsnotify.Create,
		dirB:  fsnotify.Remove,
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if op, found := events[event.Name]; found && op == event.Op&op {
					delete(events, event.Name)
				}

				if len(events) == 0 {
					done <- true
				}
			case <-time.After(3 * time.Second):
				done <- true
			}
		}
	}()

	if err = watcher.Add(tempDir); err != nil {
		t.Fatal(err)
	}

	ioutil.WriteFile(fileA, []byte{'a'}, 0600)
	os.Mkdir(dirB, 0777)
	ioutil.WriteFile(fileC, []byte{'c'}, 0600)
	os.Mkdir(dirD, 0777)
	time.Sleep(100 * time.Millisecond)
	ioutil.WriteFile(fileE, []byte{'e'}, 0600)
	os.RemoveAll(dirB)

	<-done

	if len(events) != 0 {
		t.Fatalf("%+v events didn't occur", events)
	}
}
