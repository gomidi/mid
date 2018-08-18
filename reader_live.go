package mid

import (
	"io"

	"github.com/gomidi/midi/midimessage/realtime"
	"github.com/gomidi/midi/midireader"
)

// Read reads midi messages from src until an error happens (for "live" MIDI data "over the wire").
// io.EOF is the expected error that is returned when reading should stop.
//
// Read does not close the src.
//
// The messages are dispatched to the corresponding attached functions of the handler.
//
// They must be attached before Reader.Read is called
// and they must not be unset or replaced until Read returns.
func (r *Reader) Read(src io.Reader, options ...midireader.Option) (err error) {
	r.pos = nil
	rtreader := func(m realtime.Message) {
		switch m {
		// ticks (most important, must be sent every 10 milliseconds) comes first
		case realtime.Tick:
			if r.Message.Realtime.Tick != nil {
				r.Message.Realtime.Tick()
			}

		// clock a bit slower synchronization method (24 MIDI Clocks in every quarter note) comes next
		case realtime.TimingClock:
			if r.Message.Realtime.Clock != nil {
				r.Message.Realtime.Clock()
			}

		// ok starting and continuing should not take too lpng
		case realtime.Start:
			if r.Message.Realtime.Start != nil {
				r.Message.Realtime.Start()
			}

		case realtime.Continue:
			if r.Message.Realtime.Continue != nil {
				r.Message.Realtime.Continue()
			}

		// Active Sense must come every 300 milliseconds (but is seldom implemented)
		case realtime.ActiveSensing:
			if r.Message.Realtime.ActiveSense != nil {
				r.Message.Realtime.ActiveSense()
			}

		// put any user defined realtime message here
		case realtime.Undefined4:
			if r.Message.Unknown != nil {
				r.Message.Unknown(r.pos, m)
			}

		// ok, stopping is not so urgent
		case realtime.Stop:
			if r.Message.Realtime.Stop != nil {
				r.Message.Realtime.Stop()
			}

		// reset may take some time
		case realtime.Reset:
			if r.Message.Realtime.Reset != nil {
				r.Message.Realtime.Reset()
			}

		}
	}

	rd := midireader.New(src, rtreader, options...)
	return r.read(rd)
}
