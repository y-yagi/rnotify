package rnotify_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

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

	watcher, _ := rnotify.NewWatcher()
	defer watcher.Close()

	events := []string{}
	done := make(chan bool)

	fileA := filepath.Join(tempDir, "a")
	dirB := filepath.Join(tempDir, "b")
	fileC := filepath.Join(tempDir, "foo", "c")
	dirD := filepath.Join(tempDir, "foo", "d")
	fileE := filepath.Join(tempDir, "foo", "d", "e")
	expected := []string{fileA, fileA, dirB, fileC, fileC, dirD, fileE, fileE}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				events = append(events, event.Name)
				if len(events) == len(expected) {
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

	ioutil.WriteFile(fileA, []byte{'a'}, 0644)
	os.Mkdir(dirB, 0777)
	ioutil.WriteFile(fileC, []byte{'c'}, 0644)
	os.Mkdir(dirD, 0777)
	time.Sleep(100 * time.Millisecond)
	ioutil.WriteFile(fileE, []byte{'e'}, 0644)

	<-done

	if !reflect.DeepEqual(expected, events) {
		t.Fatalf("Expect %v, but got %v", expected, events)
	}
}
