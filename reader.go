package mid

import (
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/smf"
)

// Reader allows the reading of either "over the wire" MIDI
// data (via Read) or SMF MIDI data (via ReadSMF or ReadSMFFile).
//
// Before any of the Read* methods are called, callbacks for the MIDI messages of interest
// need to be attached to the Reader. This callbacks are then invoked when the corresponding
// MIDI message arrives. They must not be changed while any of the Read* methods is running.
//
// It is possible to share the same Reader for reading of the wire MIDI ("live")
// and SMF Midi data. However, only channel messages and system exclusive message
// may be used in both cases. To enable this, the corresponding callbacks
// get a pointer to the SMFPosition of the MIDI message. This pointer is always nil
// for "live" MIDI data and never nil when reading from a SMF.
//
// The SMF header callback and the meta message callbacks are only called, when reading data
// from an SMF. Therefore the SMFPosition is passed directly and can't be nil.
//
// System common and realtime message callbacks will only be called when reading "live" MIDI,
// so they get no SMFPosition.
type Reader struct {
	tempoChanges []tempoChange // track tempo changes
	header       smf.Header    // store the SMF header
	logger       Logger        // optional logger
	pos          *SMFPosition  // the current SMFPosition
	errSMF       error         // error when reading SMF

	// SMFHeader is the callback that gets SMF header data
	SMFHeader func(smf.Header)

	// Message provides callbacks for MIDI messages
	Message struct {

		// Each is called for every MIDI message in addition to the other callback.
		Each func(*SMFPosition, midi.Message)

		// Unknown is called for undefined or unknown messages
		Unknown func(p *SMFPosition, msg midi.Message)

		// Meta provides callbacks for meta messages (only in SMF files)
		Meta struct {

			// Copyright is called for the copyright message
			Copyright func(p SMFPosition, text string)

			// Tempo is called for the tempo (change) message
			Tempo func(p SMFPosition, bpm uint32)

			// TimeSignature is called for the time signature (change) message
			TimeSignature func(p SMFPosition, num, denom uint8)

			// KeySignature is called for the key signature (change) message
			KeySignature func(p SMFPosition, key uint8, ismajor bool, num_accidentals uint8, accidentals_are_flat bool)

			// Track is called for the track (name) message
			Track func(p SMFPosition, name string)

			// Sequence is called for the sequence (name) message
			Sequence func(p SMFPosition, name string)

			// SequenceNumber is called for the sequence number message
			SequenceNumber func(p SMFPosition, number uint16)

			// Marker is called for the marker message
			Marker func(p SMFPosition, text string)

			// Cuepoint is called for the cuepoint message
			Cuepoint func(p SMFPosition, text string)

			// Text is called for the text message
			Text func(p SMFPosition, text string)

			// Lyric is called for the lyric message
			Lyric func(p SMFPosition, text string)

			// EndOfTrack is called for the end of a track message
			EndOfTrack func(p SMFPosition)

			// DevicePort is called for the device port message
			DevicePort func(p SMFPosition, name string)

			// ProgramName is called for the program name message
			ProgramName func(p SMFPosition, text string)

			// SMPTEOffset is called for the smpte offset message
			SMPTEOffset func(p SMFPosition, hour, minute, second, frame, fractionalFrame byte)

			// SequencerSpecific is called for the sequencer specific message
			SequencerSpecific func(p SMFPosition, data []byte)

			// MIDIChannel is called for the deprecated MIDI channel message
			MIDIChannel func(p SMFPosition, channel uint8)

			// MIDIPort is called for the deprecated MIDI port message
			MIDIPort func(p SMFPosition, port uint8)
		}

		// Channel provides callbacks for channel messages
		// They may occur in SMF files and in live MIDI.
		// For live MIDI *SMFPosition is nil.
		Channel struct {

			// NoteOn is just called for noteon messages with a velocity > 0.
			// Noteon messages with velocity == 0 will trigger NoteOff with a velocity of 0.
			NoteOn func(p *SMFPosition, channel, key, velocity uint8)

			// NoteOff is called for noteoff messages (then the given velocity is passed)
			// and for noteon messages of velocity 0 (then velocity is 0).
			NoteOff func(p *SMFPosition, channel, key, velocity uint8)

			// PitchBend is called for pitch bend messages
			PitchBend func(p *SMFPosition, channel uint8, value int16)

			// ProgramChange is called for program change messages. Program numbers start with 0.
			ProgramChange func(p *SMFPosition, channel, program uint8)

			// AfterTouch is called for aftertouch messages  (aka "channel pressure")
			AfterTouch func(p *SMFPosition, channel, pressure uint8)

			// PolyphonicAfterTouch is called for polyphonic aftertouch messages (aka "key pressure").
			PolyphonicAfterTouch func(p *SMFPosition, channel, key, pressure uint8)

			// ControlChange is called for control change messages
			ControlChange func(p *SMFPosition, channel, controller, value uint8)
		}

		// Realtime provides callbacks for realtime messages.
		// They are only used with "live" MIDI
		Realtime struct {

			// Clock is called for a clock message
			Clock func()

			// Tick is called for a tick message
			Tick func()

			// ActiveSense is called for a active sense message
			ActiveSense func()

			// Start is called for a start message
			Start func()

			// Stop is called for a stop message
			Stop func()

			// Continue is called for a continue message
			Continue func()

			// Reset is called for a reset message
			Reset func()
		}

		// SystemCommon provides callbacks for system common messages.
		// They are only used with "live" MIDI
		SystemCommon struct {

			// TuneRequest is called for a tune request message
			TuneRequest func()

			// SongSelect is called for a song select message
			SongSelect func(num uint8)

			// SongPositionPointer is called for a song position pointer message
			SongPositionPointer func(pos uint16)

			// MIDITimingCode is called for a MIDI timing code message
			MIDITimingCode func(frame uint8)
		}

		// SystemExcluse provides callbacks for system exclusive messages.
		// They may occur in SMF files and in live MIDI.
		// For live MIDI *SMFPosition is nil.
		SystemExcluse struct {

			// Complete is called for a complete system exclusive message
			Complete func(p *SMFPosition, data []byte)

			// Start is called for a starting system exclusive message
			Start func(p *SMFPosition, data []byte)

			// Continue is called for a continuing system exclusive message
			Continue func(p *SMFPosition, data []byte)

			// End is called for an ending system exclusive message
			End func(p *SMFPosition, data []byte)

			// Escape is called for an escaping system exclusive message
			Escape func(p *SMFPosition, data []byte)
		}
	}
}
