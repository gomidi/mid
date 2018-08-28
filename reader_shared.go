package mid

import (
	"time"

	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midimessage/meta"
	"github.com/gomidi/midi/midimessage/syscommon"
	"github.com/gomidi/midi/midimessage/sysex"
	"github.com/gomidi/midi/smf"
)

func (r *Reader) saveTempoChange(pos Position, bpm uint32) {
	r.tempoChanges = append(r.tempoChanges, tempoChange{pos.AbsoluteTicks, bpm})
}

// TimeAt returns the time.Duration at the given absolute position counted
// from the beginning of the file, respecting all the tempo changes in between.
// If the time format is not of type smf.MetricTicks, nil is returned.
func (r *Reader) TimeAt(absTicks uint64) *time.Duration {
	if r.resolution == 0 {
		return nil
	}

	var tc = tempoChange{0, 120}
	var lastTick uint64
	var lastDur time.Duration
	for _, t := range r.tempoChanges {
		if t.absTicks >= absTicks {
			// println("stopping")
			break
		}
		// println("pre", "lastDur", lastDur, "lastTick", lastTick, "bpm", tc.bpm)
		lastDur += calcDeltaTime(r.resolution, uint32(t.absTicks-lastTick), tc.bpm)
		tc = t
		lastTick = t.absTicks
	}
	result := lastDur + calcDeltaTime(r.resolution, uint32(absTicks-lastTick), tc.bpm)
	return &result
}

// log does the logging
func (r *Reader) log(m midi.Message) {
	if r.pos != nil {
		r.logger.Printf("#%v [%v d:%v] %s\n", r.pos.Track, r.pos.AbsoluteTicks, r.pos.DeltaTicks, m)
	} else {
		r.logger.Printf("%s\n", m)
	}
}

// dispatch dispatches the messages from the midi.Reader (which might be an smf reader)
// for realtime reading, the passed *SMFPosition is nil
func (r *Reader) dispatch(rd midi.Reader) (err error) {
	for {
		err = r.dispatchMessage(rd)
		if err != nil {
			return
		}
	}
}

// dispatchMessage dispatches a single message from the midi.Reader (which might be an smf reader)
// for realtime reading, the passed *SMFPosition is nil
func (r *Reader) dispatchMessage(rd midi.Reader) (err error) {
	var m midi.Message
	m, err = rd.Read()
	if err != nil {
		return
	}

	if frd, ok := rd.(smf.Reader); ok && r.pos != nil {
		r.pos.DeltaTicks = frd.Delta()
		r.pos.AbsoluteTicks += uint64(r.pos.DeltaTicks)
		r.pos.Track = frd.Track()
	}

	if r.logger != nil {
		r.log(m)
	}

	if r.Msg.Each != nil {
		r.Msg.Each(r.pos, m)
	}

	switch msg := m.(type) {

	// most common event, should be exact
	case channel.NoteOn:
		if r.Msg.Channel.NoteOn != nil {
			r.Msg.Channel.NoteOn(r.pos, msg.Channel(), msg.Key(), msg.Velocity())
		}

	// proably second most common
	case channel.NoteOff:
		if r.Msg.Channel.NoteOff != nil {
			r.Msg.Channel.NoteOff(r.pos, msg.Channel(), msg.Key(), 0)
		}

	case channel.NoteOffVelocity:
		if r.Msg.Channel.NoteOff != nil {
			r.Msg.Channel.NoteOff(r.pos, msg.Channel(), msg.Key(), msg.Velocity())
		}

	// if send there often are a lot of them
	case channel.Pitchbend:
		if r.Msg.Channel.Pitchbend != nil {
			r.Msg.Channel.Pitchbend(r.pos, msg.Channel(), msg.Value())
		}

	case channel.PolyAftertouch:
		if r.Msg.Channel.PolyAftertouch != nil {
			r.Msg.Channel.PolyAftertouch(r.pos, msg.Channel(), msg.Key(), msg.Pressure())
		}

	case channel.Aftertouch:
		if r.Msg.Channel.Aftertouch != nil {
			r.Msg.Channel.Aftertouch(r.pos, msg.Channel(), msg.Pressure())
		}

	case channel.ControlChange:
		if r.Msg.Channel.ControlChange != nil {
			r.Msg.Channel.ControlChange(r.pos, msg.Channel(), msg.Controller(), msg.Value())
		}

	case meta.SMPTE:
		if r.Msg.Meta.SMPTE != nil {
			r.Msg.Meta.SMPTE(*r.pos, msg.Hour, msg.Minute, msg.Second, msg.Frame, msg.FractionalFrame)
		}

	case meta.Tempo:
		r.saveTempoChange(*r.pos, msg.BPM())
		if r.Msg.Meta.Tempo != nil {
			r.Msg.Meta.Tempo(*r.pos, msg.BPM())
		}

	case meta.TimeSig:
		if r.Msg.Meta.TimeSig != nil {
			r.Msg.Meta.TimeSig(*r.pos, msg.Numerator, msg.Denominator)
		}

		// may be for karaoke we need to be fast
	case meta.Lyric:
		if r.Msg.Meta.Lyric != nil {
			r.Msg.Meta.Lyric(*r.pos, msg.Text())
		}

	// may be useful to synchronize by sequence number
	case meta.SequenceNo:
		if r.Msg.Meta.SequenceNo != nil {
			r.Msg.Meta.SequenceNo(*r.pos, msg.Number())
		}

	case meta.Marker:
		if r.Msg.Meta.Marker != nil {
			r.Msg.Meta.Marker(*r.pos, msg.Text())
		}

	case meta.Cuepoint:
		if r.Msg.Meta.Cuepoint != nil {
			r.Msg.Meta.Cuepoint(*r.pos, msg.Text())
		}

	case meta.Program:
		if r.Msg.Meta.Program != nil {
			r.Msg.Meta.Program(*r.pos, msg.Text())
		}

	case meta.SequencerData:
		if r.Msg.Meta.SequencerData != nil {
			r.Msg.Meta.SequencerData(*r.pos, msg.Data())
		}

	case sysex.SysEx:
		if r.Msg.SysEx.Complete != nil {
			r.Msg.SysEx.Complete(r.pos, msg.Data())
		}

	case sysex.Start:
		if r.Msg.SysEx.Start != nil {
			r.Msg.SysEx.Start(r.pos, msg.Data())
		}

	case sysex.End:
		if r.Msg.SysEx.End != nil {
			r.Msg.SysEx.End(r.pos, msg.Data())
		}

	case sysex.Continue:
		if r.Msg.SysEx.Continue != nil {
			r.Msg.SysEx.Continue(r.pos, msg.Data())
		}

	case sysex.Escape:
		if r.Msg.SysEx.Escape != nil {
			r.Msg.SysEx.Escape(r.pos, msg.Data())
		}

	// this usually takes some time
	case channel.ProgramChange:
		if r.Msg.Channel.ProgramChange != nil {
			r.Msg.Channel.ProgramChange(r.pos, msg.Channel(), msg.Program())
		}

	// the rest is not that interesting for performance
	case meta.Key:
		if r.Msg.Meta.Key != nil {
			r.Msg.Meta.Key(*r.pos, msg.Key, msg.IsMajor, msg.Num, msg.IsFlat)
		}

	case meta.Sequence:
		if r.Msg.Meta.Sequence != nil {
			r.Msg.Meta.Sequence(*r.pos, msg.Text())
		}

	case meta.Track:
		if r.Msg.Meta.Track != nil {
			r.Msg.Meta.Track(*r.pos, msg.Text())
		}

	case meta.Channel:
		if r.Msg.Meta.Deprecated.Channel != nil {
			r.Msg.Meta.Deprecated.Channel(*r.pos, msg.Number())
		}

	case meta.Port:
		if r.Msg.Meta.Deprecated.Port != nil {
			r.Msg.Meta.Deprecated.Port(*r.pos, msg.Number())
		}

	case meta.Text:
		if r.Msg.Meta.Text != nil {
			r.Msg.Meta.Text(*r.pos, msg.Text())
		}

	case syscommon.SongSelect:
		if r.Msg.SysCommon.SongSelect != nil {
			r.Msg.SysCommon.SongSelect(msg.Number())
		}

	case syscommon.SPP:
		if r.Msg.SysCommon.SPP != nil {
			r.Msg.SysCommon.SPP(msg.Number())
		}

	case syscommon.MTC:
		if r.Msg.SysCommon.MTC != nil {
			r.Msg.SysCommon.MTC(msg.QuarterFrame())
		}

	case meta.Copyright:
		if r.Msg.Meta.Copyright != nil {
			r.Msg.Meta.Copyright(*r.pos, msg.Text())
		}

	case meta.Device:
		if r.Msg.Meta.Device != nil {
			r.Msg.Meta.Device(*r.pos, msg.Text())
		}

	//case meta.Undefined, syscommon.Undefined4, syscommon.Undefined5:
	case meta.Undefined:
		if r.Msg.Unknown != nil {
			r.Msg.Unknown(r.pos, m)
		}

	default:
		switch m {
		case syscommon.Tune:
			if r.Msg.SysCommon.Tune != nil {
				r.Msg.SysCommon.Tune()
			}
		case meta.EndOfTrack:
			if _, ok := rd.(smf.Reader); ok && r.pos != nil {
				r.pos.DeltaTicks = 0
				r.pos.AbsoluteTicks = 0
			}
			if r.Msg.Meta.EndOfTrack != nil {
				r.Msg.Meta.EndOfTrack(*r.pos)
			}
		default:

			if r.Msg.Unknown != nil {
				r.Msg.Unknown(r.pos, m)
			}

		}

	}

	return
}

type tempoChange struct {
	absTicks uint64
	bpm      uint32
}

func calcDeltaTime(mt smf.MetricTicks, deltaTicks uint32, bpm uint32) time.Duration {
	return mt.Duration(bpm, deltaTicks)
}
