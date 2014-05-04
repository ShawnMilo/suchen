#!/usr/bin/env bash

# Look how easy cross-compilation is in Go!

GOARCH=amd64 GOOS=linux go build -o suchen_linux_32-bit 
GOARCH=386 GOOS=linux go build -o suchen_linux_32-bit 

GOARCH=amd64 GOOS=darwin go build -o suchen_OSX_64-bit

GOARCH=amd64 GOOS=windows go build -o suchen_Win64
GOARCH=386 GOOS=windows go build -o suchen_Win32
