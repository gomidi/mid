// Copyright (c) 2017 Marc René Arns. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package mid provides an easy abstraction for reading and writing of MIDI data live or from SMF files.
For examples see the examples folder.

Sharing callbacks for "over the wire" MIDI data and SMF MIDI data

The user attaches callback functions to the Reader and they get invoked as the MIDI data is read.
When SMF data is read, the SMFPosition is not nil. For over the wire MIDI it is nil.
To facilitate dealing with tracks and delta ticks, the SMFPosition provides absolute ticks and
the current track number.

*/
package mid
