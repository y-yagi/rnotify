module github.com/y-yagi/rnotify

go 1.16

require (
	github.com/fsnotify/fsevents v0.1.1
	github.com/fsnotify/fsnotify v1.5.0
	golang.org/x/sys v0.0.0-20210820121016-41cdb8703e55 // indirect
)

replace github.com/fsnotify/fsnotify => github.com/y-yagi/fsnotify v1.4.10-0.20201227062311-078207fcf401
