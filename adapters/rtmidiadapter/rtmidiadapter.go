package rtmidiadapter

import (
	"github.com/gomidi/mid"
	//github.com/thestk/rtmidi/tree/master/contrib/go/rtmidi
	"github.com/gomidi/mid/imported/rtmidi"
)

func Out(o rtmidi.MIDIOut) mid.Out {
	return &out{o}
}

type out struct {
	rtmidi.MIDIOut
}

func (o *out) Send(b []byte) error {
	return o.SendMessage(b)
}

func In(i rtmidi.MIDIIn) mid.In {
	return &in{i}
}

type in struct {
	rtmidi.MIDIIn
}

func (i *in) SetListener(f func([]byte)) {
	i.SetCallback(func(_ rtmidi.MIDIIn, bt []byte, _ float64) {
		f(bt)
	})
}
