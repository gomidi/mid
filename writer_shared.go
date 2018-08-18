package mid

import (
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midimessage/sysex"
)

type midiWriter struct {
	wr midi.Writer
	ch channel.Channel
}

// SetChannel sets the channel for the following midi messages
// Channel numbers are counted from 0 to 15 (MIDI channel 1 to 16).
// The initial channel number is 0.
func (w *midiWriter) SetChannel(no uint8 /* 0-15 */) {
	w.ch = channel.Channel(no)
}

// ChannelPressure writes a channel pressure message for the current channel
func (w *midiWriter) ChannelPressure(pressure uint8) error {
	return w.wr.Write(w.ch.ChannelPressure(pressure))
}

// KeyPressure writes a key pressure message for the current channel
func (w *midiWriter) KeyPressure(key, pressure uint8) error {
	return w.wr.Write(w.ch.KeyPressure(key, pressure))
}

// NoteOff writes a note off message for the current channel
func (w *midiWriter) NoteOff(key uint8) error {
	return w.wr.Write(w.ch.NoteOff(key))
}

// NoteOn writes a note on message for the current channel
func (w *midiWriter) NoteOn(key, veloctiy uint8) error {
	return w.wr.Write(w.ch.NoteOn(key, veloctiy))
}

// PitchBend writes a pitch bend message for the current channel
// For reset value, use 0, for lowest -8191 and highest 8191
// Or use the pitch constants of midimessage/channel
func (w *midiWriter) PitchBend(value int16) error {
	return w.wr.Write(w.ch.PitchBend(value))
}

// ProgramChange writes a program change message for the current channel
// Program numbers start with 0 for program 1.
func (w *midiWriter) ProgramChange(program uint8) error {
	return w.wr.Write(w.ch.ProgramChange(program))
}

// MsbLsb writes a Msb control change message, followed by a Lsb control change message
// for the current channel
// For more comfortable use, used it in conjunction with the gomidi/cc package
func (w *midiWriter) MsbLsb(msb, lsb uint8, value uint16) error {

	var b = make([]byte, 2)
	b[1] = byte(value & 0x7F)
	b[0] = byte((value >> 7) & 0x7F)

	/*
		r := midilib.MsbLsbSigned(value)

		var b = make([]byte, 2)

		binary.BigEndian.PutUint16(b, r)
	*/
	err := w.ControlChange(msb, b[0])
	if err != nil {
		return err
	}
	return w.ControlChange(lsb, b[1])
}

// ControlChange writes a control change message for the current channel
// For more comfortable use, used it in conjunction with the gomidi/cc package
func (w *midiWriter) ControlChange(controller, value uint8) error {
	return w.wr.Write(w.ch.ControlChange(controller, value))
}

// ControlChangeOff writes a control change message with a value of 0 (=off) for the current channel
func (w *midiWriter) ControlChangeOff(controller uint8) error {
	return w.ControlChange(controller, 0)
}

// ControlChangeOn writes a control change message with a value of 127 (=on) for the current channel
func (w *midiWriter) ControlChangeOn(controller uint8) error {
	return w.ControlChange(controller, 127)
}

// SystemExclusive writes system exclusive data
func (w *midiWriter) SystemExclusive(data []byte) error {
	return w.wr.Write(sysex.SysEx(data))
}
