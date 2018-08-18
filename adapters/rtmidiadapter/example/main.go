package main

import (
	"fmt"
	"github.com/gomidi/mid"
	"github.com/gomidi/mid/adapters/rtmidiadapter"
	"github.com/gomidi/mid/imported/rtmidi"
	"time"
)

func main() {

	{ // find the ports
		printInPorts()
		printOutPorts()
		fmt.Println(" \n--Messages--")
	}

	var ( // wire it up
		midiIn, midiOut = openMIDIIn(0), openMIDIOut(0)
		in, out         = rtmidiadapter.In(midiIn), rtmidiadapter.Out(midiOut)
		rd              = mid.NewReader()
		wr              = mid.SpeakTo(out)
	)

	// listen for MIDI
	go rd.ListenTo(in)

	{ // write MIDI
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

	{ // clean up
		in.StopListening()
		midiIn.Destroy()
		midiOut.Destroy()
	}
}

func openMIDIIn(port int) rtmidi.MIDIIn {
	in, err := rtmidi.NewMIDIInDefault()
	if err != nil {
		panic("can't open default MIDI in:" + err.Error())
	}

	err = in.OpenPort(port, "")
	if err != nil {
		panic("can't open MIDI in port:" + err.Error())
	}

	return in
}

func openMIDIOut(port int) rtmidi.MIDIOut {
	out, err := rtmidi.NewMIDIOutDefault()
	if err != nil {
		panic("can't open default MIDI out:" + err.Error())
	}

	err = out.OpenPort(port, "")
	if err != nil {
		panic("can't open MIDI out port:" + err.Error())
	}

	return out
}

func printInPorts() {
	in, err := rtmidi.NewMIDIInDefault()
	if err != nil {
		panic("can't open default MIDI in:" + err.Error())
	}

	ports, errP := in.PortCount()
	if errP != nil {
		panic("can't get number of in ports:" + errP.Error())
	}

	fmt.Println("\n---MIDI input ports---")

	for i := 0; i < ports; i++ {
		name, _ := in.PortName(i)
		fmt.Printf("%d %#v\n", i, name)
	}
}

func printOutPorts() {
	out, err := rtmidi.NewMIDIOutDefault()
	if err != nil {
		panic("can't open default MIDI out:" + err.Error())
	}

	ports, errP := out.PortCount()
	if errP != nil {
		panic("can't get number of out ports:" + errP.Error())
	}

	fmt.Println("\n---MIDI output ports---")

	for i := 0; i < ports; i++ {
		name, _ := out.PortName(i)
		fmt.Printf("%d %#v\n", i, name)
	}
}
