package mid

import (
	"bytes"
)

// In is an interface for external/over the wire MIDI input
type In interface {
	SetListener(func([]byte))
	StopListening()
}

// Out is an interface for external/over the wire MIDI output
type Out interface {
	Send([]byte) error
}

type outWriter struct {
	out Out
}

func (w *outWriter) Write(b []byte) (int, error) {
	return len(b), w.out.Send(b)
}

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

func (r *Reader) ListenTo(in In) {
	rd := &inReader{rd: r, in: in}
	rd.in.SetListener(rd.handleMessage)
}
