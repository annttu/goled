all: build

build:
	GOPATH=$(shell pwd) go install goled
