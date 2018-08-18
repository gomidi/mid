package portmidiadapter

import (
	"github.com/gomidi/mid"
	"github.com/rakyll/portmidi"
	"sync"
)

func Out(o *portmidi.Stream) mid.Out {
	return &out{o}
}

type out struct {
	*portmidi.Stream
}

func (o *out) Send(b []byte) error {
	return o.WriteShort(int64(b[0]), int64(b[1]), int64(b[2]))
}

func In(i *portmidi.Stream) mid.In {
	return &in{Stream: i}
}

type in struct {
	*portmidi.Stream
	mx      sync.Mutex
	stopped bool
}

func (i *in) StopListening() {
	i.mx.Lock()
	i.stopped = true
	i.mx.Unlock()
}

func (i *in) read(cb func([]byte)) error {
	//1024
	//events, err := r.in.Read(3)
	events, err := i.Read(1024)

	if err != nil {
		return err
	}

	for _, ev := range events {
		var b = make([]byte, 3)
		b[0] = byte(ev.Status)
		b[1] = byte(ev.Data1)
		b[2] = byte(ev.Data2)
		cb(b)
	}

	return nil
}

func (i *in) SetListener(f func([]byte)) {
	for i.stopped == false {
		i.read(f)
	}
}

/*

Write to a MIDI Device
out, err := portmidi.NewOutputStream(deviceID, 1024, 0)
if err != nil {
	log.Fatal(err)
}

// note on events to play C major chord
out.WriteShort(0x90, 60, 100)
out.WriteShort(0x90, 64, 100)
out.WriteShort(0x90, 67, 100)

// notes will be sustained for 2 seconds
time.Sleep(2 * time.Second)

// note off events
out.WriteShort(0x80, 60, 100)
out.WriteShort(0x80, 64, 100)
out.WriteShort(0x80, 67, 100)

out.Close()
*/

/*
Read from a MIDI Device

in, err := portmidi.NewInputStream(deviceID, 1024)
if err != nil {
    log.Fatal(err)
}
defer in.Close()

events, err := in.Read(1024)
if err != nil {
    log.Fatal(err)
}

// alternatively you can filter the input to listen
// only a particular set of channels
in.SetChannelMask(portmidi.Channel(1) | portmidi.Channel.(2))
in.Read(1024) // will retrieve events from channel 1 and 2

// or alternatively listen events
ch := in.Listen()
event := <-ch
*/

/*
Cleanup

Cleanup your input and output streams once you're done. Likely to be called on graceful termination.

portmidi.Terminate()
*/
