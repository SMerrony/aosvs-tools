// simhTape.go

// Copyright (C) 2017  Steve Merrony

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

package simhTape

import (
	"fmt"
	"log"
	"os"
)

// SimH Tape Image Markers
const (
	SimhMtrTmk    = 0          // tape mark
	SimhMtrEom    = 0xFFFFFFFF // end of medium
	SimhMtrGap    = 0xFFFFFFFE // primary gap
	SimhMtrMaxlen = 0x00FFFFFF // max len is 24b
	SimhMtrErf    = 0x80000000 // error flag
	SimhMaxRecLen = 32768
)

// ReadMetaData reads a four byte (one doubleword) header, trailer, or other metadata record
// from the supplied tape image file
func ReadMetaData(imgFile *os.File) (uint32, bool) {
	hdrBytes := make([]byte, 4)
	nb, err := imgFile.Read(hdrBytes)
	if err != nil {
		log.Printf("ERROR: Could not read simH Tape Image record header: due to: %s\n", err.Error())
		return 0, false
	}
	if nb != 4 {
		log.Printf("ERROR: Wrong length simH Tape Image record header: %d\n", nb)
		return 0, false
	}
	//logging.DebugPrint(logging.DEBUG_LOG,"Debug - Header bytes: %d %d %d %d\n", hdrBytes[0], hdrBytes[1], hdrBytes[2], hdrBytes[3])
	var hdr uint32
	hdr = uint32(hdrBytes[3]) << 24
	hdr |= uint32(hdrBytes[2]) << 16
	hdr |= uint32(hdrBytes[1]) << 8
	hdr |= uint32(hdrBytes[0])
	return hdr, true
}

// WriteMetaData writes a 4-byte header/trailer or other metadata
func WriteMetaData(imgFile *os.File, hdr uint32) bool {
	hdrBytes := make([]byte, 4)
	hdrBytes[3] = byte(hdr >> 24)
	hdrBytes[2] = byte(hdr >> 16)
	hdrBytes[1] = byte(hdr >> 8)
	hdrBytes[0] = byte(hdr)
	fmt.Printf("DEBUG: WriteMetaData got %d, writing bytes: %d %d %d %d\n", hdr, hdrBytes[0], hdrBytes[1], hdrBytes[2], hdrBytes[3])
	nb, err := imgFile.Write(hdrBytes)
	if err != nil || nb != 4 {
		log.Fatalf("ERROR: Could not write simh tape header record due to %s\n", err.Error())
	}
	return true
}

// ReadRecordData attempts to read a data record from SimH tape image, fails if wrong number of bytes read
// N.B. does not read the header and trailer
func ReadRecordData(imgFile *os.File, byteLen int) ([]byte, bool) {
	rec := make([]byte, byteLen)
	nb, err := imgFile.Read(rec)
	if err != nil {
		log.Printf("ERROR: Could not read simH Tape Image record due to: %s\n", err.Error())
		return nil, false
	}
	if nb != byteLen {
		log.Printf("ERROR: Could not read simH Tape Image record, got %d bytes, expecting %d\n", nb, byteLen)
		return nil, false
	}
	return rec, true
}

// WriteRecordData writes the actual data - not the header/trailer
func WriteRecordData(imgFile *os.File, rec []byte) bool {
	nb, err := imgFile.Write(rec)
	if err != nil {
		log.Printf("ERROR: Could not write simh tape record due to %s\n", err.Error())
		return false
	}
	if nb != len(rec) {
		log.Printf("ERROR: Could not write complete record (Wrote %d of %d bytes)\n", nb, len(rec))
		return false
	}
	return true
}

// Rewind simulates a tape rewind by seeking to the start of the tape image file
func Rewind(imgFile *os.File) bool {
	_, err := imgFile.Seek(0, 0)
	if err != nil {
		log.Printf("ERROR: Could not seek to start of SimH Tape Image due to %s\n", err.Error())
		return false
	}
	return true
}

// SpaceFwd advances the virtual tape by the specified amount (0 means 1 whole file)
func SpaceFwd(imgFile *os.File, recCnt int) bool {
	var hdr, trailer uint32
	done := false

	// special case when recCnt == 0 which means space forward one file...
	if recCnt == 0 {
		for !done {
			hdr, _ = ReadMetaData(imgFile)
			if hdr == SimhMtrTmk {
				done = true
			} else {
				// read record and throw it away
				ReadRecordData(imgFile, int(hdr))
				// read trailer
				trailer, _ = ReadMetaData(imgFile)
				if hdr != trailer {
					log.Fatal("ERROR: simhTape.SpaceFwd found non-matching header/trailer")
				}
			}
		}
	} else {
		log.Fatal("ERROR: simhTape.SpaceFwd called with record count != 0 - Not Yet Implemented")
	}

	return true
}

// ScanImage - attempt to read a whole tape image ensuring headers, record sizes, and trailers match
// if csv is true then output is in CSV format
func ScanImage(imgFileName string, csv bool) (res string) {

	imgFile, err := os.Open(imgFileName)
	if err != nil {
		log.Fatalf("ERROR: Could not open tape image file %s for ScanImage function", err.Error())
	}
	defer imgFile.Close()

	var (
		fileSize, markCount, fileCount, recNum int
		header, trailer                        uint32 // a DG-DoubleWord
		ok                                     bool
	)
	fileCount = -1

recLoop:
	for {
		header, ok = ReadMetaData(imgFile)
		if !ok {
			log.Fatal("Exiting")
		}
		// if *vFlag {
		// 	res += fmt.Sprintf("...Read Header(meta) value: %d...", trailer)
		// }
		switch header {
		case SimhMtrTmk:
			if fileSize > 0 {
				fileCount++
				if csv {
					res += fmt.Sprintf("file%d,%d\n", fileCount, fileSize/recNum)
				} else {
					res += fmt.Sprintf("\nFile %d : %12d bytes in %6d block(s) avg. block size %d",
						fileCount, fileSize, recNum, fileSize/recNum)
				}
				fileSize = 0
				recNum = 0
			}
			markCount++
			if markCount == 3 {
				if csv {
					res += "EOT,0"
				} else {
					res += "\nTriple Mark (old End Of Tape indicator)"
				}
				break recLoop
			}
		case SimhMtrEom:
			res += "\nEnd of Medium"
			break recLoop
		case SimhMtrGap:
			res += "\nErase Gap"
			markCount = 0
		default:
			recNum++
			markCount = 0
			_, ok := ReadRecordData(imgFile, int(header)) // read record and throw away
			if !ok {
				fmt.Printf("%s\n", res)
				log.Fatal("Exiting")
			}
			trailer, ok = ReadMetaData(imgFile)
			if !ok {
				fmt.Printf("%s\n", res)
				log.Fatal("Exiting")
			}
			// if *vFlag {
			// 	res += fmt.Sprintf("...Read Trailer value: %d...", trailer)
			// }
			//logging.DebugPrint(logging.DEBUG_LOG,"Debug: got trailer value: %d\n", trailer)
			if header == trailer {
				fileSize += int(header)
			} else {
				res += "\nNon-matching trailer found."
			}
		}
	}
	return res
}

// DumpFiles - attempt to read a whole tape image and dump its contents
// to appropriately named files in the current directory
func DumpFiles(imgFileName string) {

	imgFile, err := os.Open(imgFileName)
	if err != nil {
		log.Fatalf("ERROR: Could not open tape image file %s for DumpFiles function", err.Error())
	}
	defer imgFile.Close()

	var (
		fileSize, markCount, fileCount, recNum int
		header, trailer                        uint32 // a DG-DoubleWord
		fileName                               string
		writeFile                              *os.File
		ok                                     bool
	)
	fileCount = 0
	fileName = fmt.Sprintf("file%d", fileCount)
	writeFile, err = os.Create(fileName)
	if err != nil {
		log.Fatalf("ERROR: Could not create file %s due to %v", fileName, err)
	}
recLoop:
	for {
		header, ok = ReadMetaData(imgFile)
		if !ok {
			log.Fatal("Exiting")
		}

		switch header {
		case SimhMtrTmk:
			if fileSize > 0 {
				writeFile.Close()
				fileCount++
				fileName = fmt.Sprintf("file%d", fileCount)
				writeFile, err = os.Create(fileName)
				if err != nil {
					log.Fatalf("ERROR: Could not create file %s due to %v", fileName, err)
				}
				fileSize = 0
				recNum = 0
			}
			markCount++
			if markCount == 3 {
				break recLoop
			}
		case SimhMtrEom:
			if fileSize > 0 {
				writeFile.Close()
			}
			break recLoop
		case SimhMtrGap:
			markCount = 0
		default:
			recNum++
			markCount = 0
			blob, ok := ReadRecordData(imgFile, int(header))
			if !ok {
				log.Fatal("Exiting")
			}
			trailer, ok = ReadMetaData(imgFile)
			if !ok {
				log.Fatal("Exiting")
			}
			if header == trailer {
				writeFile.Write(blob)
				fileSize += int(header)
			}
		}
	}
	writeFile.Close()
	if fileSize == 0 {
		os.Remove(fileName)
	}
}
