package mid

import (
	"bytes"
	"time"

	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midireader"
	"github.com/gomidi/midi/smf"
)

// InConnection is an interface for external/over the wire MIDI input connections.
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the InConnection interface.
type InConnection interface {

	// SetListener sets the callback function that is called when data arrives
	// println(big.NewRat(math.MaxInt64,1000 /* milliseonds */ *1000 /* seconds */ *60 /* minutes */ *60 /* hours */ *24 /* days */ *365 /* years */).FloatString(0))
	// output: 292471
	// => a ascending timestamp based on microseconds would wrap after 292471 years
	SetListener(func(data []byte, deltaMicroseconds int64))

	// StopListening stops the listening
	StopListening()
}

// OutConnection is an interface for external/over the wire MIDI output connections.
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the OutConnection interface.
type OutConnection interface {

	// Send sends the given MIDI bytes over the wire.
	Send([]byte) error
}

type outWriter struct {
	out OutConnection
}

func (w *outWriter) Write(b []byte) (int, error) {
	return len(b), w.out.Send(b)
}

// SpeakTo returns a Writer that outputs to the given MIDI out port.
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the Out interface.
func SpeakTo(out OutConnection) *Writer {
	return NewWriter(&outWriter{out})
}

type inReader struct {
	rd         *Reader
	in         InConnection
	midiReader midi.Reader
	bf         bytes.Buffer
}

func (r *inReader) handleMessage(b []byte, deltaMicroseconds int64) {
	// use the fake position to get the ticks for the current tempo
	r.rd.pos = &Position{}
	r.rd.pos.DeltaTicks = r.rd.Ticks(time.Duration(deltaMicroseconds * 1000)) // deltaticks
	r.rd.pos.AbsoluteTicks += uint64(r.rd.pos.DeltaTicks)
	r.bf.Write(b)
	r.rd.dispatchMessage(r.midiReader)
}

// Ticks returns the ticks that correspond to a duration while respecting the current tempo
func (r *Reader) Ticks(d time.Duration) uint32 {
	tempochange := r.tempoChanges[len(r.tempoChanges)-1]
	return r.resolution.Ticks(tempochange.bpm, d)
}

// ListenTo configures the Reader to listen to the given MIDI in port.
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the In interface.
func (r *Reader) ListenTo(in InConnection) {
	r.resolution = smf.MetricTicks(1920)
	rd := &inReader{rd: r, in: in}
	rd.midiReader = midireader.New(&rd.bf, r.dispatchRealTime, r.midiReaderOptions...)
	rd.in.SetListener(rd.handleMessage)
}
