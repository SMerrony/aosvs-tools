// aosvsFstatPacket.go - AOS/VS FSTAT system call packet structure

// Based on info from System Call Dictionary 093-000241 p.2-162
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

const (
	slthW = 25
	slthB = 50

	flnk = 0
	fdir = 12
	fdmp = 64 // guessed symbol
	fstf = 67
	ftxt = 68
	fprv = 74
	fprg = 87
)

type fstatPktT struct {
	recordFormat ByteT
	entryType    ByteT
	stim, sacp, shfs, slau, smsh, smsl, smil, stch, stcl, stah, stal, stmh, stml, ssts,
	sefw, sefl, sfah, sfal, sidx, sopn, scsh, scsl WordT
}
