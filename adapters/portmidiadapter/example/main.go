package main

import (
	"bytes"
	"fmt"
	"github.com/gomidi/mid"
	"github.com/rakyll/portmidi"
	"io"
	"log"
	"time"
)

type writer struct {
	out *portmidi.Stream
}

func newWriter(out *portmidi.Stream) io.Writer {
	return &writer{out}
}

func (w *writer) Write(b []byte) (int, error) {
	err := w.out.WriteShort(int64(b[0]), int64(b[1]), int64(b[2]))
	return len(b), err
}

type reader struct {
	in *portmidi.Stream
	rd *mid.Reader
}

func newReader(rd *mid.Reader, in *portmidi.Stream) func() error {
	r := &reader{in, rd}

	return func() error {
		for {
			// r.Read()

			if r.Read() == io.EOF {
				return io.EOF
			}

		}
	}
}

func (r *reader) Read() error {
	//1024
	events, err := r.in.Read(3)

	if err != nil {
		return err
	}

	/*

		b[0] = byte(events[0].Status)
		b[1] = byte(events[0].Data1)
		b[2] = byte(events[0].Data2)
	*/

	if len(events) > 0 {
		fmt.Println(len(events))
	}
	for _, ev := range events {
		var b = make([]byte, 3)
		b[0] = byte(ev.Status)
		b[1] = byte(ev.Data1)
		b[2] = byte(ev.Data2)

		err = r.rd.Read(bytes.NewReader(b))
		// if err != nil {
		fmt.Println(err)
		// return err
		// }
	}

	return nil
}

func main() {
	portmidi.Initialize()

	// portmidi.CountDevices() // returns the number of MIDI devices
	// portmidi.Info(deviceID) // returns info about a MIDI device
	defIn := portmidi.DefaultInputDeviceID()   // returns the ID of the system default input
	defOut := portmidi.DefaultOutputDeviceID() // returns the ID of the system default output
	println(portmidi.Info(defOut))
	println(portmidi.Info(defIn))

	out, err := portmidi.NewOutputStream(defOut, 1024, 0)
	if err != nil {
		log.Fatal(err)
	}

	in, err := portmidi.NewInputStream(defIn, 1024)
	if err != nil {
		log.Fatal(err)
	}

	r := mid.NewReader()
	rdFunc := newReader(r, in)

	done := make(chan bool)
	go func() {
		rdFunc()
		done <- true
		return
	}()

	wr := mid.NewWriter(newWriter(out))
	wr.NoteOn(60, 100)
	time.Sleep(time.Nanosecond)
	wr.NoteOn(65, 100)
	time.Sleep(time.Nanosecond)
	wr.NoteOff(60)
	time.Sleep(time.Nanosecond)
	wr.NoteOff(65)
	time.Sleep(time.Second)
	fmt.Println("all written")
	<-done
	out.Close()
	//	out.WriteShort(0x90, 60, 100)

	// defer in.Close()

	/*
		events, err := in.Read(1024)
		if err != nil {
			log.Fatal(err)
		}
	*/
}
