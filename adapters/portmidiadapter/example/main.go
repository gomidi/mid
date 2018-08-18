package main

import (
	"bytes"
	"fmt"
	"github.com/gomidi/mid"
	"github.com/gomidi/mid/adapters/portmidiadapter"
	"github.com/rakyll/portmidi"
	"time"
)

func main() {

	// don't forget!
	portmidi.Initialize()

	{ // find the ports
		printPorts()
		fmt.Println(" \n--Messages--")
	}

	var ( // wire it up
		midiIn  = openMIDIIn(portmidi.DefaultInputDeviceID())
		midiOut = openMIDIOut(portmidi.DefaultOutputDeviceID())
		in, out = portmidiadapter.In(midiIn), portmidiadapter.Out(midiOut)
		rd      = mid.NewReader()
		wr      = mid.SpeakTo(out)
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
		midiIn.Close()
		midiOut.Close()
	}
}

func openMIDIIn(port portmidi.DeviceID) *portmidi.Stream {
	in, err := portmidi.NewInputStream(port, 1024)

	if err != nil {
		panic("can't open MIDI in port:" + err.Error())
	}

	return in
}

func openMIDIOut(port portmidi.DeviceID) *portmidi.Stream {
	out, err := portmidi.NewOutputStream(port, 1024, 0)

	if err != nil {
		panic("can't open MIDI out port:" + err.Error())
	}

	return out
}

func printPorts() {
	var ins, outs bytes.Buffer

	no := portmidi.CountDevices()

	for i := 0; i < no; i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		if info.IsInputAvailable {
			fmt.Fprintf(&ins, "%d %#v\n", i, info.Name)
		}

		if info.IsOutputAvailable {
			fmt.Fprintf(&outs, "%d %#v\n", i, info.Name)
		}
	}

	fmt.Println("\n---MIDI input ports---")
	fmt.Println(ins.String())

	fmt.Println("\n---MIDI output ports---")
	fmt.Println(outs.String())

}
