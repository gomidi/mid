package mid

import (
	"bytes"
	"time"

	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midireader"
	"github.com/gomidi/midi/smf"
)

// In is an interface for external/over the wire MIDI input
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the In interface.
type In interface {
	SetListener(func(msg []byte, deltaMicroseconds int64))
	StopListening()
}

// Out is an interface for external/over the wire MIDI output
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the Out interface.
type Out interface {

	// Send sends the given MIDI bytes over the wire.
	Send([]byte) error
}

type outWriter struct {
	out Out
}

func (w *outWriter) Write(b []byte) (int, error) {
	return len(b), w.out.Send(b)
}

// SpeakTo returns a Writer that outputs to the given MIDI out port.
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the Out interface.
func SpeakTo(out Out) *Writer {
	return NewWriter(&outWriter{out})
}

type inReader struct {
	rd         *Reader
	in         In
	midiReader midi.Reader
	bf         bytes.Buffer
}

func (r *inReader) handleMessage(b []byte, deltaMicroseconds int64) {
	// use the fake position to get the ticks for the current tempo
	r.rd.pos = &Position{}
	tempochange := r.rd.tempoChanges[len(r.rd.tempoChanges)-1]
	deltaticks := r.rd.resolution.Ticks(tempochange.bpm, time.Duration(deltaMicroseconds*1000))
	r.rd.pos.DeltaTicks = deltaticks
	r.rd.pos.AbsoluteTicks += uint64(deltaticks)
	r.bf.Write(b)
	r.rd.dispatchMessage(r.midiReader)
}

// ListenTo configures the Reader to listen to the given MIDI in port.
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the In interface.
func (r *Reader) ListenTo(in In) {
	r.resolution = smf.MetricTicks(1920)
	rd := &inReader{rd: r, in: in}
	rd.midiReader = midireader.New(&rd.bf, r.dispatchRealTime, r.midiReaderOptions...)
	rd.in.SetListener(rd.handleMessage)
}
