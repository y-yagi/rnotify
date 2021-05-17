# rnotify

![Build Status](https://github.com/y-yagi/rnotify/workflows/CI/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/y-yagi/rnotify.svg)](https://pkg.go.dev/github.com/y-yagi/rnotify)

`rnotify` is a wrapper of [fsnotify](https://github.com/fsnotify/fsnotify), supports recursive directory watching on Go.

## Supported Platforms

Linux, Windows and macOS.

## Usage

```go
package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/y-yagi/rnotify"
)

func main() {
	watcher, err := rnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				// `Op` is the same as `fsnotify.Op`.
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/tmp/foo")
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

```
