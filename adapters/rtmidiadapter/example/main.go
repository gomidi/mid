package main

import (
	"fmt"
	"github.com/gomidi/mid"
	"github.com/gomidi/mid/adapters/rtmidiadapter"
	"github.com/gomidi/mid/imported/rtmidi"
	"github.com/gomidi/midi"
	"io"
	"log"
	"time"
)

/*
in, err := NewMIDIInDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer in.Destroy()
	if err := in.OpenPort(0, "RtMidi"); err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	for {
		m, t, err := in.Message()
		if len(m) > 0 {
			log.Println(m, t, err)
		}
	}
*/

func main() {
	in, err := rtmidi.NewMIDIInDefault()
	if err != nil {
		log.Fatal(err)
	}

	api, err2 := in.API()
	if err2 != nil {
		log.Fatal(err2)
	}

	ports, err3 := in.PortCount()
	if err3 != nil {
		log.Fatal(err3)
	}

	fmt.Printf("%s ports in: %d\n", api, ports)

	name, err4 := in.PortName(0)
	if err4 != nil {
		log.Fatal(err4)
	}

	fmt.Printf("port0 in: %#v\n", name)

	err5 := in.OpenPort(0, "")
	if err5 != nil {
		log.Fatal(err5)
	}

	// in.Destroy()
	/*
		err5 = in.OpenPort(0, "")
		if err5 != nil {
			log.Fatal(err5)
		}
	*/
	out, err6 := rtmidi.NewMIDIOutDefault()
	if err6 != nil {
		log.Fatal(err6)
	}

	portsOut, err7 := out.PortCount()
	if err7 != nil {
		log.Fatal(err7)
	}

	fmt.Printf("ports out: %d\n", portsOut)

	nameOut, err8 := out.PortName(0)
	if err8 != nil {
		log.Fatal(err8)
	}

	fmt.Printf("port0 out: %#v\n", nameOut)

	err9 := out.OpenPort(0, "")
	if err9 != nil {
		log.Fatal(err9)
	}

	/*
		out.Destroy()
		err9 = out.OpenPort(0, "")
		if err9 != nil {
			log.Fatal(err9)
		}
	*/

	_ = io.EOF

	rd := mid.NewReader()
	rd.Message.Each = func(_ *mid.SMFPosition, msg midi.Message) {
		fmt.Printf("%s\n", msg)
	}
	go func() {
		rtmidiadapter.ReadFrom(rd, in)
		/*
			for {
				if rd.Read(rtmidiadapter.NewReader(in)) == io.EOF {
					break
				}
			}
		*/
	}()

	wr := mid.NewWriter(rtmidiadapter.NewWriter(out))
	wr.NoteOn(60, 100)
	time.Sleep(time.Nanosecond)
	wr.NoteOff(60)
	time.Sleep(time.Nanosecond)
	wr.SetChannel(1)
	wr.NoteOn(70, 100)
	time.Sleep(time.Nanosecond)
	wr.NoteOff(70)
	time.Sleep(time.Second * 1)
}
