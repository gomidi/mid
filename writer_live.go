package mid

import (
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midimessage/realtime"
	"github.com/gomidi/midi/midimessage/syscommon"
	"github.com/gomidi/midi/midiwriter"
	"io"
)

// Writer writes live MIDI data. Its methods must not be called concurrently
type Writer struct {
	*midiWriter
}

var _ midi.Writer = &Writer{}

// NewWriter creates and new Writer for writing of "live" MIDI data ("over the wire")
// By default it makes no use of the running status.
func NewWriter(dest io.Writer, options ...midiwriter.Option) *Writer {
	options = append(
		[]midiwriter.Option{
			midiwriter.NoRunningStatus(),
		}, options...)

	wr := midiwriter.New(dest, options...)
	return &Writer{&midiWriter{wr: wr, ch: channel.Channel0}}
}

// ActiveSensing writes the active sensing realtime message
func (w *Writer) ActiveSensing() error {
	return w.midiWriter.wr.Write(realtime.ActiveSensing)
}

// Continue writes the continue realtime message
func (w *Writer) Continue() error {
	return w.midiWriter.wr.Write(realtime.Continue)
}

// Reset writes the reset realtime message
func (w *Writer) Reset() error {
	return w.midiWriter.wr.Write(realtime.Reset)
}

// Start writes the start realtime message
func (w *Writer) Start() error {
	return w.midiWriter.wr.Write(realtime.Start)
}

// Stop writes the stop realtime message
func (w *Writer) Stop() error {
	return w.midiWriter.wr.Write(realtime.Stop)
}

// Tick writes the tick realtime message
func (w *Writer) Tick() error {
	return w.midiWriter.wr.Write(realtime.Tick)
}

// TimingClock writes the timing clock realtime message
func (w *Writer) TimingClock() error {
	return w.midiWriter.wr.Write(realtime.TimingClock)
}

// MIDITimingCode writes the MIDI Timing Code system message
func (w *Writer) MIDITimingCode(code uint8) error {
	return w.midiWriter.wr.Write(syscommon.MIDITimingCode(code))
}

// SongPositionPointer writes the song position pointer system message
func (w *Writer) SongPositionPointer(ptr uint16) error {
	return w.midiWriter.wr.Write(syscommon.SongPositionPointer(ptr))
}

// SongSelect writes the song select system message
func (w *Writer) SongSelect(song uint8) error {
	return w.midiWriter.wr.Write(syscommon.SongSelect(song))
}

// TuneRequest writes the tune request system message
func (w *Writer) TuneRequest() error {
	return w.midiWriter.wr.Write(syscommon.TuneRequest)
}
