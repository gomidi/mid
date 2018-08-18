# mid
Porcelain library for reading and writing MIDI and SMF

Based on https://github.com/gomidi/midi.

[![Build Status Travis/Linux](https://travis-ci.org/gomidi/mid.svg?branch=master)](http://travis-ci.org/gomidi/mid) [![Coverage Status](https://coveralls.io/repos/github/gomidi/mid/badge.svg)](https://coveralls.io/github/gomidi/mid) [![Go Report](https://goreportcard.com/badge/github.com/gomidi/mid)](https://goreportcard.com/report/github.com/gomidi/mid) [![Documentation](http://godoc.org/github.com/gomidi/mid?status.png)](http://godoc.org/github.com/gomidi/mid)

## Example

```go
package main

import (
    "fmt"
    "github.com/gomidi/mid"
    "io"
    "time"
)

func noteOn(p *mid.SMFPosition, channel, key, vel uint8) {
    fmt.Printf("NoteOn (ch %v: key %v vel: %v)\n", channel, key, vel)
}

func noteOff(p *mid.SMFPosition, channel, key, vel uint8) {
    fmt.Printf("NoteOff (ch %v: key %v)\n", channel, key)
}

func main() {
    fmt.Println()

    // to disable logging, pass mid.NoLogger() as option
    rd := mid.NewReader()

    // set the functions for the messages you are interested in
    rd.Message.Channel.NoteOn = noteOn
    rd.Message.Channel.NoteOff = noteOff

    // to allow reading and writing concurrently in this example
    // we need a pipe
    piperd, pipewr := io.Pipe()

    go func() {
        wr := mid.NewWriter(pipewr)
        wr.SetChannel(11) // sets the channel for the next messages
        wr.NoteOn(120, 50)
        time.Sleep(time.Second) // let the note ring for 1 sec
        wr.NoteOff(120)
        pipewr.Close() // finishes the writing
    }()

    for {
        if rd.Read(piperd) == io.EOF {
            piperd.Close() // finishes the reading
            break
        }
    }

    // Output:
    // channel.NoteOn{channel:0xb, key:0x78, velocity:0x32}
    // NoteOn (ch 11: key 120 vel: 50)
    // channel.NoteOff{channel:0xb, key:0x78}
    // NoteOff (ch 11: key 120)
}
```


## Status

API mostly stable and complete

- Go version: >= 1.5
- OS/architectures: everywhere Go runs (tested on Linux and Windows).

## Installation

```
go get -d github.com/gomidi/mid/...
```

## License

MIT (see LICENSE file) 
