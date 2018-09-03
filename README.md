# mid
Porcelain library for reading and writing MIDI and SMF (Standard MIDI File) 

Based on https://github.com/gomidi/midi.

[![Build Status Travis/Linux](https://travis-ci.org/gomidi/mid.svg?branch=master)](http://travis-ci.org/gomidi/mid) [![Coverage Status](https://coveralls.io/repos/github/gomidi/mid/badge.svg)](https://coveralls.io/github/gomidi/mid) [![Go Report](https://goreportcard.com/badge/github.com/gomidi/mid)](https://goreportcard.com/report/github.com/gomidi/mid) [![Documentation](http://godoc.org/github.com/gomidi/mid?status.png)](http://godoc.org/github.com/gomidi/mid)

## Description

Package mid provides an easy abstraction for reading and writing of "live" `MIDI` and `SMF` 
(Standard MIDI File) data.

`MIDI` data could be written the following ways:

- `NewWriter` is used to write "live" MIDI to an `io.Writer`.
- `NewSMF` is used to write SMF MIDI to an `io.Writer`.
- `NewSMFFile` is used to write a complete SMF file.
- `WriteTo` writes "live" MIDI to an `connect.Out`, aka MIDI out port

To read, create a `Reader` and attach callbacks to it.
Then MIDI data could be read the following ways:

- `Reader.Read` reads "live" MIDI from an `io.Reader`.
- `Reader.ReadSMF` reads SMF MIDI from an `io.Reader`.
- `Reader.ReadSMFFile` reads a complete SMF file.
- `Reader.ReadFrom` reads "live" MIDI from an `connect.In`, aka MIDI in port

For a simple example with "live" MIDI and `io.Reader` and `io.Writer` see the example below.

To connect with the MIDI ports of your computer (via connect.In and connect.Out), use it with one of the 
driver packages for `rtmidi` and `portaudio` at https://github.com/gomidi/connect.

There you can find a simple example how to do it.

## Example

We use an `io.Writer` to write to and `io.Reader` to read from. They are connected by the same `io.Pipe`.

```go
package main

import (
    "fmt"
    "github.com/gomidi/mid"
    "io"
    "time"
)

// callback for note on messages
func noteOn(p *mid.Position, channel, key, vel uint8) {
    fmt.Printf("NoteOn (ch %v: key %v vel: %v)\n", channel, key, vel)
}

// callback for note off messages
func noteOff(p *mid.Position, channel, key, vel uint8) {
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
    // channel.NoteOn channel 11 key 120 velocity 50
    // NoteOn (ch 11: key 120 vel: 50)
    // channel.NoteOff channel 11 key 120
    // NoteOff (ch 11: key 120)
}
```

## Status

API mostly stable and complete

- Go version: >= 1.10
- OS/architectures: everywhere Go runs (tested on Linux and Windows).

## Installation

It is recommended to use Go 1.11 with module support (`$GO111MODULE=on`).

```
go get -d github.com/gomidi/mid/...
```

## License

MIT (see LICENSE file) 
