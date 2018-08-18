package mid

import (
	"bytes"
)

// In is an interface for external/over the wire MIDI input
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the In interface.
type In interface {
	SetListener(func([]byte))
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
	rd *Reader
	in In
}

func (r *inReader) handleMessage(b []byte) {
	r.rd.Read(bytes.NewReader(b))
}

// ListenTo configures the Reader to listen to the given MIDI in port.
// The gomidi/connect package provides adapters to rtmidi and portaudio
// that fullfill the In interface.
func (r *Reader) ListenTo(in In) {
	rd := &inReader{rd: r, in: in}
	rd.in.SetListener(rd.handleMessage)
}
