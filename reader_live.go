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
	rd := midireader.New(src, r.dispatchRealTime, options...)
	return r.dispatch(rd)
}

func (r *Reader) dispatchRealTime(m realtime.Message) {

	// ticks (most important, must be sent every 10 milliseconds) comes first
	if m == realtime.Tick {
		if r.Message.Realtime.Tick != nil {
			r.Message.Realtime.Tick()
		}
		return
	}

	// clock a bit slower synchronization method (24 MIDI Clocks in every quarter note) comes next
	if m == realtime.TimingClock {
		if r.Message.Realtime.Clock != nil {
			r.Message.Realtime.Clock()
		}
		return
	}

	// starting should not take too long
	if m == realtime.Start {
		if r.Message.Realtime.Start != nil {
			r.Message.Realtime.Start()
		}
		return
	}

	// continuing should not take too long
	if m == realtime.Continue {
		if r.Message.Realtime.Continue != nil {
			r.Message.Realtime.Continue()
		}
		return
	}

	// Active Sense must come every 300 milliseconds
	// (but is seldom implemented)
	if m == realtime.ActiveSensing {
		if r.Message.Realtime.ActiveSense != nil {
			r.Message.Realtime.ActiveSense()
		}
		return
	}

	// put any user defined realtime message here
	if m == realtime.Undefined4 {
		if r.Message.Unknown != nil {
			r.Message.Unknown(r.pos, m)
		}
		return
	}

	// stopping is not so urgent
	if m == realtime.Stop {
		if r.Message.Realtime.Stop != nil {
			r.Message.Realtime.Stop()
		}
		return
	}

	// reset may take some time anyway
	if m == realtime.Reset {
		if r.Message.Realtime.Reset != nil {
			r.Message.Realtime.Reset()
		}
		return
	}
}
