# Termloader
[![Build Status](https://travis-ci.org/bharath-srinivas/termloader.svg?branch=master)](https://travis-ci.org/bharath-srinivas/termloader)
[![GoDoc](https://godoc.org/github.com/bharath-srinivas/termloader?status.svg)](https://godoc.org/github.com/bharath-srinivas/termloader)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Termloader is a simple library to add a loading screen to your terminal applications. Termloader will render the 
loader at the center of your terminal screen. Currently termloader is supported only on *nix operating systems.

## Installation
```bash
go get github.com/bharath-srinivas/term-loader
``` 

## Example Usage
```go
package main

import (
	"time"

	"github.com/bharath-srinivas/termloader"
)

func main() {
	loader := termloader.New(termloader.Charsets[0], 100 * time.Millisecond) // construct a new loader with config

	loader.Start() // start the loader
	time.Sleep(5 * time.Second) // sleep for sometime to simulate a task
	loader.Stop() // stop the loader
}
```

## Loader color
```go
loader.Color = termloader.Green // provide a color for the loader (white if not provided)
```

## Provide a loading text
```go
loader.Text = "Loading" // provide a text to show above the loader
```

## Loading text color
```go
loadingText := termloader.ColorString("Now Loading", termloader.Green) // color the string
loader.Text = loadingText // provide the colored string as loading text
```

## Provide your own character set for loader
```go
charSet := []string{"|", "/", "-", "\"}
loader := termloader.New(charSet, 100 * time.Millisecond)
```

## Todo
- [x] Loader
- [x] Optional loading text support
- [ ] Optional image/icon support
- [ ] Add a gif

## License
MIT