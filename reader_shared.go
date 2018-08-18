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

// Reader reads the midi messages coming from an SMF file or a live stream.
//
// The messages are dispatched to the corresponding functions that are not nil.
//
// The desired functions must be attached before Handler.ReadLive or Handler.ReadSMF is called
// and they must not be changed while these methods are running.
type Reader struct {
	tempoChanges []tempoChange
	header       smf.Header

	// callback functions for SMF (Standard MIDI File) header data
	SMFHeader func(smf.Header)

	// callback functions for MIDI messages
	Message struct {
		// is called in addition to other functions, if set.
		Each func(*SMFPosition, midi.Message)

		// undefined or unknown messages
		Unknown func(p *SMFPosition, msg midi.Message)

		// meta messages (only in SMF files)
		Meta struct {
			// SMF general settings
			Copyright     func(p SMFPosition, text string)
			Tempo         func(p SMFPosition, bpm uint32)
			TimeSignature func(p SMFPosition, num, denom uint8)
			KeySignature  func(p SMFPosition, key uint8, ismajor bool, num_accidentals uint8, accidentals_are_flat bool)

			// SMF tracks and sequence definitions
			Track          func(p SMFPosition, name string)
			Sequence       func(p SMFPosition, name string)
			SequenceNumber func(p SMFPosition, number uint16)

			// SMF text entries
			Marker   func(p SMFPosition, text string)
			Cuepoint func(p SMFPosition, text string)
			Text     func(p SMFPosition, text string)
			Lyric    func(p SMFPosition, text string)

			// SMF diverse
			EndOfTrack        func(p SMFPosition)
			DevicePort        func(p SMFPosition, name string)
			ProgramName       func(p SMFPosition, text string)
			SMPTEOffset       func(p SMFPosition, hour, minute, second, frame, fractionalFrame byte)
			SequencerSpecific func(p SMFPosition, data []byte)

			// deprecated
			MIDIChannel func(p SMFPosition, channel uint8)
			MIDIPort    func(p SMFPosition, port uint8)
		}

		// channel messages, may be in SMF files and in live data
		// for live data *SMFPosition is nil
		Channel struct {
			// NoteOn is just called for noteon messages with a velocity > 0
			// noteon messages with velocity == 0 will trigger NoteOff with a velocity of 0
			NoteOn func(p *SMFPosition, channel, key, velocity uint8)

			// NoteOff is triggered by noteoff messages (then the given velocity is passed)
			// and by noteon messages of velocity 0 (then velocity is 0)
			NoteOff func(p *SMFPosition, channel, key, velocity uint8)

			// PolyphonicAfterTouch aka key pressure
			PolyphonicAfterTouch func(p *SMFPosition, channel, key, pressure uint8)

			ControlChange func(p *SMFPosition, channel, controller, value uint8)
			ProgramChange func(p *SMFPosition, channel, program uint8)

			// AfterTouch aka channel pressure
			AfterTouch func(p *SMFPosition, channel, pressure uint8)
			PitchBend  func(p *SMFPosition, channel uint8, value int16)
		}

		// realtime messages: just in live data
		Realtime struct {
			Reset       func()
			Clock       func()
			Tick        func()
			Start       func()
			Continue    func()
			Stop        func()
			ActiveSense func()
		}

		// system common messages: just in live data
		SystemCommon struct {
			TuneRequest         func()
			SongSelect          func(num uint8)
			SongPositionPointer func(pos uint16)
			MIDITimingCode      func(frame uint8)
		}

		// system exclusive, may be in SMF files and in live data
		// for live data *SMFPosition is nil
		SystemExcluse struct {
			Complete func(p *SMFPosition, data []byte)
			Start    func(p *SMFPosition, data []byte)
			Continue func(p *SMFPosition, data []byte)
			End      func(p *SMFPosition, data []byte)
			Escape   func(p *SMFPosition, data []byte)
		}
	}

	// optional logger
	logger Logger

	pos    *SMFPosition
	errSMF error
}

// NewReader returns a new reader
func NewReader(opts ...ReaderOption) *Reader {
	h := &Reader{logger: logfunc(printf)}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (r *Reader) registerTempoChange(pos SMFPosition, bpm uint32) {
	r.tempoChanges = append(r.tempoChanges, tempoChange{pos.AbsTime, bpm})
}

// TimeAt returns the time.Duration at the given absolute position counted
// from the beginning of the file, respecting all the tempo changes in between.
// If the time format is not of type smf.MetricTicks, nil is returned.
func (r *Reader) TimeAt(absTicks uint64) *time.Duration {
	mt, isMetric := r.header.TimeFormat.(smf.MetricTicks)
	if !isMetric {
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
		lastDur += calcDeltaTime(mt, uint32(t.absTicks-lastTick), tc.bpm)
		tc = t
		lastTick = t.absTicks
	}
	result := lastDur + calcDeltaTime(mt, uint32(absTicks-lastTick), tc.bpm)
	return &result
}

// log does the logging
func (r *Reader) log(m midi.Message) {
	if r.pos != nil {
		r.logger.Printf("#%v [%v d:%v] %#v\n", r.pos.Track, r.pos.AbsTime, r.pos.Delta, m)
	} else {
		r.logger.Printf("%#v\n", m)
	}
}

// dispatch dispatches the messages from the midi.Reader (which might be an smf reader)
// for realtime reading, the passed *SMFPosition is nil
func (r *Reader) dispatch(rd midi.Reader) (err error) {
	var m midi.Message

	for {
		m, err = rd.Read()
		if err != nil {
			break
		}

		if frd, ok := rd.(smf.Reader); ok && r.pos != nil {
			r.pos.Delta = frd.Delta()
			r.pos.AbsTime += uint64(r.pos.Delta)
			r.pos.Track = frd.Track()
		}

		if r.logger != nil {
			r.log(m)
		}

		if r.Message.Each != nil {
			r.Message.Each(r.pos, m)
		}

		switch msg := m.(type) {

		// most common event, should be exact
		case channel.NoteOn:
			if r.Message.Channel.NoteOn != nil {
				r.Message.Channel.NoteOn(r.pos, msg.Channel(), msg.Key(), msg.Velocity())
			}

		// proably second most common
		case channel.NoteOff:
			if r.Message.Channel.NoteOff != nil {
				r.Message.Channel.NoteOff(r.pos, msg.Channel(), msg.Key(), 0)
			}

		case channel.NoteOffVelocity:
			if r.Message.Channel.NoteOff != nil {
				r.Message.Channel.NoteOff(r.pos, msg.Channel(), msg.Key(), msg.Velocity())
			}

		// if send there often are a lot of them
		case channel.PitchBend:
			if r.Message.Channel.PitchBend != nil {
				r.Message.Channel.PitchBend(r.pos, msg.Channel(), msg.Value())
			}

		case channel.PolyphonicAfterTouch:
			if r.Message.Channel.PolyphonicAfterTouch != nil {
				r.Message.Channel.PolyphonicAfterTouch(r.pos, msg.Channel(), msg.Key(), msg.Pressure())
			}

		case channel.AfterTouch:
			if r.Message.Channel.AfterTouch != nil {
				r.Message.Channel.AfterTouch(r.pos, msg.Channel(), msg.Pressure())
			}

		case channel.ControlChange:
			if r.Message.Channel.ControlChange != nil {
				r.Message.Channel.ControlChange(r.pos, msg.Channel(), msg.Controller(), msg.Value())
			}

		case meta.SMPTEOffset:
			if r.Message.Meta.SMPTEOffset != nil {
				r.Message.Meta.SMPTEOffset(*r.pos, msg.Hour, msg.Minute, msg.Second, msg.Frame, msg.FractionalFrame)
			}

		case meta.Tempo:
			r.registerTempoChange(*r.pos, msg.BPM())
			if r.Message.Meta.Tempo != nil {
				r.Message.Meta.Tempo(*r.pos, msg.BPM())
			}

		case meta.TimeSignature:
			if r.Message.Meta.TimeSignature != nil {
				r.Message.Meta.TimeSignature(*r.pos, msg.Numerator, msg.Denominator)
			}

			// may be for karaoke we need to be fast
		case meta.Lyric:
			if r.Message.Meta.Lyric != nil {
				r.Message.Meta.Lyric(*r.pos, msg.Text())
			}

		// may be useful to synchronize by sequence number
		case meta.SequenceNumber:
			if r.Message.Meta.SequenceNumber != nil {
				r.Message.Meta.SequenceNumber(*r.pos, msg.Number())
			}

		case meta.Marker:
			if r.Message.Meta.Marker != nil {
				r.Message.Meta.Marker(*r.pos, msg.Text())
			}

		case meta.Cuepoint:
			if r.Message.Meta.Cuepoint != nil {
				r.Message.Meta.Cuepoint(*r.pos, msg.Text())
			}

		case meta.ProgramName:
			if r.Message.Meta.ProgramName != nil {
				r.Message.Meta.ProgramName(*r.pos, msg.Text())
			}

		case meta.SequencerSpecific:
			if r.Message.Meta.SequencerSpecific != nil {
				r.Message.Meta.SequencerSpecific(*r.pos, msg.Data())
			}

		case sysex.SysEx:
			if r.Message.SystemExcluse.Complete != nil {
				r.Message.SystemExcluse.Complete(r.pos, msg.Data())
			}

		case sysex.Start:
			if r.Message.SystemExcluse.Start != nil {
				r.Message.SystemExcluse.Start(r.pos, msg.Data())
			}

		case sysex.End:
			if r.Message.SystemExcluse.End != nil {
				r.Message.SystemExcluse.End(r.pos, msg.Data())
			}

		case sysex.Continue:
			if r.Message.SystemExcluse.Continue != nil {
				r.Message.SystemExcluse.Continue(r.pos, msg.Data())
			}

		case sysex.Escape:
			if r.Message.SystemExcluse.Escape != nil {
				r.Message.SystemExcluse.Escape(r.pos, msg.Data())
			}

		// this usually takes some time
		case channel.ProgramChange:
			if r.Message.Channel.ProgramChange != nil {
				r.Message.Channel.ProgramChange(r.pos, msg.Channel(), msg.Program())
			}

		// the rest is not that interesting for performance
		case meta.KeySignature:
			if r.Message.Meta.KeySignature != nil {
				r.Message.Meta.KeySignature(*r.pos, msg.Key, msg.IsMajor, msg.Num, msg.IsFlat)
			}

		case meta.Sequence:
			if r.Message.Meta.Sequence != nil {
				r.Message.Meta.Sequence(*r.pos, msg.Text())
			}

		case meta.Track:
			if r.Message.Meta.Track != nil {
				r.Message.Meta.Track(*r.pos, msg.Text())
			}

		case meta.MIDIChannel:
			if r.Message.Meta.MIDIChannel != nil {
				r.Message.Meta.MIDIChannel(*r.pos, msg.Number())
			}

		case meta.MIDIPort:
			if r.Message.Meta.MIDIPort != nil {
				r.Message.Meta.MIDIPort(*r.pos, msg.Number())
			}

		case meta.Text:
			if r.Message.Meta.Text != nil {
				r.Message.Meta.Text(*r.pos, msg.Text())
			}

		case syscommon.SongSelect:
			if r.Message.SystemCommon.SongSelect != nil {
				r.Message.SystemCommon.SongSelect(msg.Number())
			}

		case syscommon.SongPositionPointer:
			if r.Message.SystemCommon.SongPositionPointer != nil {
				r.Message.SystemCommon.SongPositionPointer(msg.Number())
			}

		case syscommon.MIDITimingCode:
			if r.Message.SystemCommon.MIDITimingCode != nil {
				r.Message.SystemCommon.MIDITimingCode(msg.QuarterFrame())
			}

		case meta.Copyright:
			if r.Message.Meta.Copyright != nil {
				r.Message.Meta.Copyright(*r.pos, msg.Text())
			}

		case meta.DevicePort:
			if r.Message.Meta.DevicePort != nil {
				r.Message.Meta.DevicePort(*r.pos, msg.Text())
			}

		//case meta.Undefined, syscommon.Undefined4, syscommon.Undefined5:
		case meta.Undefined:
			if r.Message.Unknown != nil {
				r.Message.Unknown(r.pos, m)
			}

		default:
			switch m {
			case syscommon.TuneRequest:
				if r.Message.SystemCommon.TuneRequest != nil {
					r.Message.SystemCommon.TuneRequest()
				}
			case meta.EndOfTrack:
				if _, ok := rd.(smf.Reader); ok && r.pos != nil {
					r.pos.Delta = 0
					r.pos.AbsTime = 0
				}
				if r.Message.Meta.EndOfTrack != nil {
					r.Message.Meta.EndOfTrack(*r.pos)
				}
			default:

				if r.Message.Unknown != nil {
					r.Message.Unknown(r.pos, m)
				}

			}

		}

	}

	return
}

// SMFPosition is the position of the event inside a standard midi file (SMF).
type SMFPosition struct {
	// the Track number
	Track int16

	// the delta time to the previous message in the same track
	Delta uint32

	// the absolute time from the beginning of the track
	AbsTime uint64
}

type tempoChange struct {
	absTicks uint64
	bpm      uint32
}

func calcDeltaTime(mt smf.MetricTicks, deltaTicks uint32, bpm uint32) time.Duration {
	return mt.Duration(bpm, deltaTicks)
}
