// aosvsDumpFmt.go - AOS/VS Dump Format structures

// Based on info from AOS/VS Systems Internals Reference Manual (AOS/VS Rev. 5.00)
// This file is part of loadg.

// Copyright 2018 Steve Merrony

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

type (
	// WordT - a DG Word is 16-bit unsigned
	WordT uint16
	// DwordT - a DG Double-Word is 32-bit unsigned
	DwordT uint32
	// ByteT - a DG Byte is 8-bit unsigned
	ByteT byte
)

const (
	startDumpType  = 0
	fsbType        = 1
	nbType         = 2
	udaType        = 3
	aclType        = 4
	linkType       = 5
	startBlockType = 6
	dataBlockType  = 7
	endBlockType   = 8
	endDumpType    = 9
)

const (
	maxBlockSize       = 32768
	maxAlignmentOffset = 256
	diskBlockBytes     = 512
)

type recordHeaderT struct {
	recordType   int
	recordLength int
}

// Start Of Dump
type sodT struct {
	sodHeader                                 recordHeaderT
	dumpFormatRevision                        WordT
	dumpTimeSecs, dumpTimeMins, dumpTimeHours WordT
	dumpTimeDay, dumpTimeMonth, dumpTimeYear  WordT
}

// FSB
type fsbT struct {
	fsbGeader recordHeaderT
	fstatPkt  fstatPktT
}

// Name Block
type nbT struct {
	nbHeader recordHeaderT
	fileName []byte
}

// User Data Area Block
type udaT struct {
	udaHeader recordHeaderT
	uda       [256]byte
}

// Access Control List block
type aclT struct {
	aclHeader recordHeaderT
	acl       []byte
}

// Link Block
type linkT struct {
	linkHeader         recordHeaderT
	linkResolutionName []byte
}

// Start Block
type startT struct {
	startBlockHeader recordHeaderT
}

// Data Header Block
type dataHeaderT struct {
	dataHeader     recordHeaderT
	byteAddress    DwordT
	byteLength     DwordT
	alignmentCount WordT
}

// End Block
type endT struct {
	endBlockHeader recordHeaderT
}

// End of Dump
type endOfDumpT struct {
	endOfDumpHeader recordHeaderT
}
