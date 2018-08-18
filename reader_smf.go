package mid

import (
	"io"

	"github.com/gomidi/midi/smf"
	"github.com/gomidi/midi/smf/smfreader"
)

// ReadSMFFile open, reads and closes a complete SMF file.
// If the read content was a valid midi file, nil is returned.
//
// The messages are dispatched to the corresponding attached functions of the handler.
//
// They must be attached before Handler.ReadSMF is called
// and they must not be unset or replaced until ReadSMF returns.
func (r *Reader) ReadSMFFile(file string, options ...smfreader.Option) error {
	r.errSMF = nil
	r.pos = &SMFPosition{}
	err := smfreader.ReadFile(file, r.readSMF, options...)
	if err != nil {
		return err
	}
	return r.errSMF
}

// ReadSMF reads midi messages from src (which is supposed to be the content of a standard midi file (SMF))
// until an error or io.EOF happens.
//
// ReadSMF does not close the src.
//
// If the read content was a valid midi file, nil is returned.
//
// The messages are dispatched to the corresponding attached functions of the Reader.
//
// They must be attached before Reader.ReadSMF is called
// and they must not be unset or replaced until ReadSMF returns.
func (r *Reader) ReadSMF(src io.Reader, options ...smfreader.Option) error {
	r.errSMF = nil
	r.pos = &SMFPosition{}
	rd := smfreader.New(src, options...)

	err := rd.ReadHeader()
	if err != nil {
		return err
	}
	r.readSMF(rd)
	return r.errSMF
}

func (r *Reader) readSMF(rd smf.Reader) {
	r.header = rd.Header()

	if r.SMFHeader != nil {
		r.SMFHeader(r.header)
	}

	err := r.read(rd)
	if err != io.EOF {
		r.errSMF = err
	}
}
