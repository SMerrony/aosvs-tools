// loadg.go

// Copyright (C) 2018  Steve Merrony

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

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const versionString = "1.2a"

// program flags (options)...
var (
	extract, ignoreErrors, list, summary, verbose, version bool
	dump                                                   string
)

func init() {
	flag.StringVar(&dump, "dumpFile", "", "DUMP_II or DUMP_III file to read/load")
	flag.StringVar(&dump, "d", "", "DUMP_II or DUMP_III file to read/load")
	flag.BoolVar(&extract, "extract", false, "extract the files from the DUMP_II/III into the current directory")
	flag.BoolVar(&extract, "e", false, "extract the files from the DUMP_II/III into the current directory")
	flag.BoolVar(&ignoreErrors, "ignoreErrors", false, "do not exit if a file cannot be created")
	flag.BoolVar(&ignoreErrors, "i", false, "do not exit if a file cannot be created")
	flag.BoolVar(&list, "list", false, "list the contents of the DUMP_II/III file")
	flag.BoolVar(&list, "l", false, "list the contents of the DUMP_II/III file")
	flag.BoolVar(&summary, "summary", true, "concise summary of the DUMP_II/III file contents")
	flag.BoolVar(&summary, "s", true, "concise summary of the DUMP_II/III file contents")
	flag.BoolVar(&verbose, "verbose", false, "be rather wordy about what loadg is doing")
	flag.BoolVar(&verbose, "v", false, "be rather wordy about what loadg is doing")
	flag.BoolVar(&version, "version", false, "show the version number of loadg and exit")
	flag.BoolVar(&version, "V", false, "show the version number of loadg and exit")
	flag.Parse()
	if !version && dump == "" {
		flag.PrintDefaults()
	}
}

func main() {
	if version {
		fmt.Printf("loadg version %s\n", versionString)
		return
	}
	dumpFile, err := os.Open(dump)
	if err != nil {
		log.Fatalf("ERROR: Could not open dump file <%s> due to %v", dump, err)
	}
	defer dumpFile.Close()

	// there should always be a SOD record...
	sod := readSod(dumpFile)
	if summary {
		fmt.Printf("AOS/VS dump version  : %d\n", sod.dumpFormatRevision)
		fmt.Printf("Dump date (y-m-d)    : %d-%d-%d\n", sod.dumpTimeYear, sod.dumpTimeMonth, sod.dumpTimeDay)
		fmt.Printf("Dump time( hh:mm:ss) : %02d:%02d:%02d\n", sod.dumpTimeHours, sod.dumpTimeMins, sod.dumpTimeSecs)
	}
}

func readAWord(file *os.File) WordT {
	var w WordT
	twoBytes := make([]byte, 2)
	n, err := file.Read(twoBytes)
	if n != 2 || err != nil {
		log.Fatalf("ERROR: Could not read Word in file <%s> due to %v", file.Name(), err)
	}
	w = WordT(twoBytes[0])<<8 | WordT(twoBytes[1])
	return w
}

func readHeader(file *os.File) recordHeaderT {
	var (
		hdr recordHeaderT
	)
	twoBytes := make([]byte, 2)
	n, err := file.Read(twoBytes)
	if n != 2 || err != nil {
		log.Fatalf("ERROR: Could not read header in file <%s> due to %v", file.Name(), err)
	}
	hdr.recordType = int(twoBytes[0]) >> 2 // 6-bit
	hdr.recordLength = int(twoBytes[0]&0x03)<<8 + int(twoBytes[1])
	return hdr
}

func readSod(file *os.File) sodT {
	var sod sodT
	sod.sodHeader = readHeader(file)
	if sod.sodHeader.recordType != startDumpType {
		log.Fatalln("ERROR: This does not appear to be an AOS/VS DUMP_II or DUMP_III file.")
	}
	sod.dumpFormatRevision = readAWord(file)
	sod.dumpTimeSecs = readAWord(file)
	sod.dumpTimeMins = readAWord(file)
	sod.dumpTimeHours = readAWord(file)
	sod.dumpTimeDay = readAWord(file)
	sod.dumpTimeMonth = readAWord(file)
	sod.dumpTimeYear = readAWord(file)
	return sod
}
