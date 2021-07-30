.PHONY: build

bin_name=go-inotify

build:
		go build -o $(bin_name) *.go

.DEFAULT_GOAL := build