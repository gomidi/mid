package mid

import (
	"fmt"

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

// Aftertouch writes a channel pressure message for the current channel
func (w *midiWriter) Aftertouch(pressure uint8) error {
	return w.wr.Write(w.ch.Aftertouch(pressure))
}

// PolyAftertouch writes a key pressure message for the current channel
func (w *midiWriter) PolyAftertouch(key, pressure uint8) error {
	return w.wr.Write(w.ch.PolyAftertouch(key, pressure))
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

// Pitchbend writes a pitch bend message for the current channel
// For reset value, use 0, for lowest -8191 and highest 8191
// Or use the pitch constants of midimessage/channel
func (w *midiWriter) Pitchbend(value int16) error {
	return w.wr.Write(w.ch.Pitchbend(value))
}

// ProgramChange writes a program change message for the current channel
// Program numbers start with 0 for program 1.
func (w *midiWriter) ProgramChange(program uint8) error {
	return w.wr.Write(w.ch.ProgramChange(program))
}

// PitchBendSensitivityRPN sets the pitch bend range via RPN
func (w *midiWriter) PitchBendSensitivityRPN(msbVal, lsbVal uint8) error {
	return w.RPN(0, 0, msbVal, lsbVal)
}

// FineTuningRPN
func (w *midiWriter) FineTuningRPN(msbVal, lsbVal uint8) error {
	return w.RPN(0, 1, msbVal, lsbVal)
}

// CoarseTuningRPN
func (w *midiWriter) CoarseTuningRPN(msbVal, lsbVal uint8) error {
	return w.RPN(0, 2, msbVal, lsbVal)
}

// TuningProgramSelectRPN
func (w *midiWriter) TuningProgramSelectRPN(msbVal, lsbVal uint8) error {
	return w.RPN(0, 3, msbVal, lsbVal)
}

// TuningBankSelectRPN
func (w *midiWriter) TuningBankSelectRPN(msbVal, lsbVal uint8) error {
	return w.RPN(0, 4, msbVal, lsbVal)
}

// ResetRPN aka Null
func (w *midiWriter) ResetRPN() error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(101, 127),
		w.ch.ControlChange(100, 127),
	)
	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write ResetRPN: %v", msg)
		}
	}
	return nil
}

// RPN message consisting of a val101 and val100 to identify the RPN and a msb and lsb for the value
func (w *midiWriter) RPN(val101, val100, msbVal, lsbVal uint8) error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(101, val101),
		w.ch.ControlChange(100, val100),
		w.ch.ControlChange(6, msbVal),
		w.ch.ControlChange(38, lsbVal))

	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write RPN(%v,%v): %v", val101, val100, msg)
		}
	}

	return w.ResetRPN()
}

func (w *midiWriter) RPNIncrement(val101, val100 uint8) error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(101, val101),
		w.ch.ControlChange(100, val100),
		w.ch.ControlChange(96, 0))

	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write RPNIncrement(%v,%v): %v", val101, val100, msg)
		}
	}

	return w.ResetRPN()
}

func (w *midiWriter) RPNDecrement(val101, val100 uint8) error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(101, val101),
		w.ch.ControlChange(100, val100),
		w.ch.ControlChange(97, 0))

	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write RPNDecrement(%v,%v): %v", val101, val100, msg)
		}
	}

	return w.ResetRPN()
}

func (w *midiWriter) NRPNIncrement(val99, val98 uint8) error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(99, val99),
		w.ch.ControlChange(98, val98),
		w.ch.ControlChange(96, 0))

	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write NRPNIncrement(%v,%v): %v", val99, val98, msg)
		}
	}
	return w.ResetNRPN()
}

func (w *midiWriter) NRPNDecrement(val99, val98 uint8) error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(99, val99),
		w.ch.ControlChange(98, val98),
		w.ch.ControlChange(97, 0))

	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write NRPNDecrement(%v,%v): %v", val99, val98, msg)
		}
	}
	return w.ResetNRPN()
}

// NRPN message consisting of a val99 and val98 to identify the RPN and a msb and lsb for the value
func (w *midiWriter) NRPN(val99, val98, msbVal, lsbVal uint8) error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(99, val99),
		w.ch.ControlChange(98, val98),
		w.ch.ControlChange(6, msbVal),
		w.ch.ControlChange(38, lsbVal))

	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write NRPN(%v,%v): %v", val99, val98, msg)
		}
	}
	return w.ResetNRPN()
}

// ResetNRPN aka Null
func (w *midiWriter) ResetNRPN() error {
	msgs := append([]midi.Message{},
		w.ch.ControlChange(99, 127),
		w.ch.ControlChange(98, 127),
	)
	for _, msg := range msgs {
		err := w.wr.Write(msg)
		if err != nil {
			return fmt.Errorf("can't write ResetNRPN: %v", msg)
		}
	}
	return nil
}

// MsbLsb writes a Msb control change message, followed by a Lsb control change message
// for the current channel
// For more comfortable use, used it in conjunction with the gomidi/cc package
func (w *midiWriter) MsbLsb(msbControl, lsbControl uint8, value uint16) error {

	var b = make([]byte, 2)
	b[1] = byte(value & 0x7F)
	b[0] = byte((value >> 7) & 0x7F)

	/*
		r := midilib.MsbLsbSigned(value)

		var b = make([]byte, 2)

		binary.BigEndian.PutUint16(b, r)
	*/
	err := w.ControlChange(msbControl, b[0])
	if err != nil {
		return err
	}
	return w.ControlChange(lsbControl, b[1])
}

// ControlChange writes a control change message for the current channel
// For more comfortable use, used it in conjunction with the gomidi/cc package
func (w *midiWriter) ControlChange(controller, value uint8) error {
	return w.wr.Write(w.ch.ControlChange(controller, value))
}

// CcOff writes a control change message with a value of 0 (=off) for the current channel
func (w *midiWriter) CcOff(controller uint8) error {
	return w.ControlChange(controller, 0)
}

// CcOn writes a control change message with a value of 127 (=on) for the current channel
func (w *midiWriter) CcOn(controller uint8) error {
	return w.ControlChange(controller, 127)
}

// SysEx writes system exclusive data
func (w *midiWriter) SysEx(data []byte) error {
	return w.wr.Write(sysex.SysEx(data))
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
