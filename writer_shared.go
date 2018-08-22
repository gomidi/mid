package mid

import (
	"fmt"
	"github.com/gomidi/cc"
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midimessage/sysex"
)

type midiWriter struct {
	wr              midi.Writer
	ch              channel.Channel
	noteState       [16][128]bool
	noConsolidation bool
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
// By default, midi notes are consolidated (see ConsolidateNotes method)
func (w *midiWriter) NoteOff(key uint8) error {
	return w.Write(w.ch.NoteOff(key))
}

// NoteOffVelocity writes a note off message for the current channel with a velocity.
// By default, midi notes are consolidated (see ConsolidateNotes method)
func (w *midiWriter) NoteOffVelocity(key, velocity uint8) error {
	return w.Write(w.ch.NoteOffVelocity(key, velocity))
}

// NoteOn writes a note on message for the current channel
// By default, midi notes are consolidated (see ConsolidateNotes method)
func (w *midiWriter) NoteOn(key, veloctiy uint8) error {
	return w.Write(w.ch.NoteOn(key, veloctiy))
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

// silentium for a single channel
func (w *midiWriter) silentium(ch uint8, force bool) (err error) {
	for key, val := range w.noteState[ch] {
		if force || val {
			err = w.wr.Write(channel.Channel(ch).NoteOff(uint8(key)))
		}
	}

	if force {
		err = w.wr.Write(channel.Channel(ch).ControlChange(cc.AllNotesOff, 0))
		err = w.wr.Write(channel.Channel(ch).ControlChange(cc.AllSoundOff, 0))
	}
	return
}

// Silence sends a  note off message for every running note
// on the given channel. If the channel is -1, every channel is affected.
// If note consolidation is switched off, notes aren't tracked and therefor
// every possible note will get a note off. This could also be enforced by setting
// force to true. If force is true, additionally the cc messages AllNotesOff (123) and AllSoundOff (120)
// are send.
// If channel is > 15, the method panics.
// The error or not error of the last write is returned.
// (i.e. it does not return on the first error, but tries everything instead to make it silent)
func (w *midiWriter) Silence(ch int8, force bool) (err error) {
	if ch > 15 {
		panic("invalid channel number")
	}
	if w.noConsolidation {
		force = true
	}

	// single channel
	if ch >= 0 {
		err = w.silentium(uint8(ch), force)
		// set note states for the channel
		w.noteState[ch] = [128]bool{}
		return
	}

	// all channels
	for c := 0; c < 16; c++ {
		err = w.silentium(uint8(c), force)
	}
	// reset all note states
	w.noteState = [16][128]bool{}
	return
}

// ConsolidateNotes enables/disables the midi note consolidation (default: enabled)
// When enabled, midi note on/off messages are consolidated, that means the on/off state of
// every possible note on every channel is tracked and note on messages are only
// written, if the corresponding note is off and vice versa. Note on messages
// with a velocity of 0 behave the same way as note offs.
// The tracking of the notes is not cross SMF tracks, i.e. a meta.EndOfTrack
// message will reset the tracking.
// The consolidation should prevent "hanging" notes in most cases.
// If on is true, the note will be started tracking again (fresh state), assuming no note is currently running.
func (w *midiWriter) ConsolidateNotes(on bool) {
	if on {
		w.noteState = [16][128]bool{}
	}
	w.noConsolidation = !on
}

// Write writes the given midi.Message. By default, midi notes are consolidated (see ConsolidateNotes method)
func (w *midiWriter) Write(msg midi.Message) error {
	if w.noConsolidation {
		return w.wr.Write(msg)
	}
	switch m := msg.(type) {
	case channel.NoteOn:
		if m.Velocity() > 0 && w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note already running.", msg)
		}
		if m.Velocity() == 0 && !w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note is not running.", msg)
		}
		w.noteState[m.Channel()][m.Key()] = m.Velocity() > 0
	case channel.NoteOff:
		if !w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note is not running.", msg)
		}
		w.noteState[m.Channel()][m.Key()] = false
	case channel.NoteOffVelocity:
		if !w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note is not running.", msg)
		}
		w.noteState[m.Channel()][m.Key()] = false
	}
	return w.wr.Write(msg)
}
