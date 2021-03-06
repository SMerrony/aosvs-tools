// loadg.go

// Copyright (C) 2018,2019  Steve Merrony

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
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const semVer = "v1.4.1"

// program flags (options)...
var (
	extract, ignoreErrors, list, summary, verbose, version bool
	dump                                                   string
)

var (
	fsbBlob                       []byte
	inFile, loadIt                bool
	totalFileSize                 int
	baseDir, fileName, workingDir string
	writeFile                     *os.File
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
}

func main() {
	if version || verbose {
		fmt.Printf("loadg version %s\n", semVer)
		if !verbose {
			return
		}
	}
	if len(dump) == 0 {
		log.Fatalln("ERROR: Must specify DUMP file name with -dumpFile <dumpfilename> option")
	}
	dumpFile, err := os.Open(dump)
	if err != nil {
		log.Fatalf("ERROR: Could not open dump file <%s> due to %v", dump, err)
	}
	defer dumpFile.Close()

	// dump images can legally contain 'too many' directory pops, so we
	// store the starting directory and never traverse above it...
	baseDir, _ = os.Getwd()
	workingDir = baseDir

	// there should always be a SOD record...
	sod := readSod(dumpFile)
	if summary || verbose {
		fmt.Printf("Summary of dump file : %s\n", dumpFile.Name())
		fmt.Printf("AOS/VS dump version  : %d\n", sod.dumpFormatRevision)
		fmt.Printf("Dump date (y-m-d)    : %d-%d-%d\n", sod.dumpTimeYear, sod.dumpTimeMonth, sod.dumpTimeDay)
		fmt.Printf("Dump time( hh:mm:ss) : %02d:%02d:%02d\n", sod.dumpTimeHours, sod.dumpTimeMins, sod.dumpTimeSecs)
	}

	// now go through the dump examining each block type and acting accordingly...
	done := false
	for !done {
		recHdr := readHeader(dumpFile)
		if verbose {
			fmt.Printf("Found block of type: %d, Length: %d\n", recHdr.recordType, recHdr.recordLength)
		}
		switch recHdr.recordType {
		case startDumpType:
			log.Fatalln("ERROR: Another START record found in DUMP - this should not happem.")
		case fsbType:
			fsbBlob = readBlob(recHdr.recordLength, dumpFile, "FSB")
			loadIt = false
		case nbType:
			fileName = processNameBlock(recHdr, fsbBlob, dumpFile)
		case udaType:
			// throw away for now
			_ = readBlob(recHdr.recordLength, dumpFile, "UDA")
		case aclType:
			aclBlob := readBlob(recHdr.recordLength, dumpFile, "ACL")
			if verbose {
				fmt.Printf(" ACL: %s\n", string(aclBlob))
			}
		case linkType:
			processLink(recHdr, fileName, dumpFile)
		case startBlockType:
			// nothing to do - it's just a recHdr
		case dataBlockType:
			processDataBlock(recHdr, fsbBlob, dumpFile)
		case endBlockType:
			processEndBlock()
		case endDumpType:
			fmt.Println("=== End of Dump ===")
			done = true
		default:
			log.Fatalf("ERROR: Unknown block type (%d) in dump file.  Giving up.", recHdr.recordType)
		}
	}
}

func processDataBlock(recHeader recordHeaderT, fsbBlob []byte, dumpFile *os.File) {
	var dhb dataHeaderT

	// first get the address and length
	fourBytes := readBlob(4, dumpFile, "byte address")
	dhb.byteAddress = DwordT(fourBytes[0])<<24 + DwordT(fourBytes[1])<<16 + DwordT(fourBytes[2])<<8 + DwordT(fourBytes[3])
	fourBytes = readBlob(4, dumpFile, "byte length")
	dhb.byteLength = DwordT(fourBytes[0])<<24 + DwordT(fourBytes[1])<<16 + DwordT(fourBytes[2])<<8 + DwordT(fourBytes[3])
	if dhb.byteLength > maxBlockSize {
		log.Fatalf("ERROR: Maximum block size exceeded ( %d vs. limit of %d).", dhb.byteLength, maxBlockSize)
	}
	twoBytes := readBlob(2, dumpFile, "alignment count")
	dhb.alignmentCount = WordT(twoBytes[0])<<8 + WordT(twoBytes[1])
	if verbose {
		fmt.Printf(" Data block: %d (bytes)\n", dhb.byteLength)
	}

	// skip any alignment bytes - usually just one
	if dhb.alignmentCount > 0 {
		if verbose {
			fmt.Printf("  Skipping %d alignment byte(s)\n", dhb.alignmentCount)
		}
		readBlob(int(dhb.alignmentCount), dumpFile, "alignment byte(s)")
	}

	dataBlob := readBlob(int(dhb.byteLength), dumpFile, "data block")

	// large areas of NULLs may be skipped over by DUMP_II/III
	// this is achieved by simply advancing the byte address so
	// we must pad out if byte address is beyond end of last block
	//if extract && writeFile != nil {

	if int(dhb.byteAddress) > totalFileSize+1 {
		paddingSize := int(dhb.byteAddress) - totalFileSize
		//fmt.Printf("File Size: %d, BA: %d, Padding Size: %d\n", totalFileSize, int(dhb.byteAddress), paddingSize)
		paddingBlock := make([]byte, paddingSize)
		if extract {
			if verbose {
				fmt.Println("  Padding with one block")
			}
			_, err := writeFile.Write(paddingBlock)
			if err != nil {
				log.Fatalf("ERROR: Could not write padding block due to %s", err.Error())
			}
		}
		totalFileSize += paddingSize
	}
	if extract {
		n, err := writeFile.Write(dataBlob)
		if n != int(dhb.byteLength) || err != nil {
			log.Fatalf("ERROR: Could not write out data due to %s", err.Error())
		}
	}
	//}
	totalFileSize += int(dhb.byteLength)
	inFile = true
}

func processEndBlock() {
	if inFile {
		if extract && loadIt {
			writeFile.Close()
		}
		if summary {
			fmt.Printf(" %12d bytes\n", totalFileSize)
		}
		totalFileSize = 0
		inFile = false
	} else {
		if workingDir != baseDir { // don't go up from start dir
			workingDir = filepath.Dir(workingDir)
		}
		if verbose {
			fmt.Printf(" Popped dir - new dir is: %s\n", workingDir)
		}
	}
	if verbose {
		fmt.Println("End Block processed")
	}
}

func processLink(recHeader recordHeaderT, linkName string, dumpFile *os.File) {
	linkTargetBA := readBlob(recHeader.recordLength, dumpFile, "link target")
	linkTargetBA = bytes.Trim(linkTargetBA, "\x00")
	// convert AOS/VS : directory separators to platform-specific separators ("\", or "/")
	linkTarget := strings.ToUpper(strings.Replace(string(linkTargetBA), ":", string(os.PathSeparator), -1))
	if summary || verbose {
		fmt.Printf(" -> Link Target: %s\n", linkTarget)
	}
	if extract {
		var oldName string
		if len(workingDir) == 0 {
			oldName = linkTarget
		} else {
			oldName = filepath.Join(workingDir, linkTarget)
			linkName = filepath.Join(workingDir, linkName)
		}
		err := os.Symlink(oldName, linkName)
		if err != nil {
			log.Printf("ERROR: Could not create symbolic link, existing file %s, link name: %s, due to %v\n",
				oldName, linkName, err)
			if !ignoreErrors {
				log.Fatalln("Giving up.")
			}
		}
	}
}

func processNameBlock(recHeader recordHeaderT, fsbBlob []byte, dumpFile *os.File) string {
	var fileType string
	nameBytes := readBlob(recHeader.recordLength, dumpFile, "file name")
	fileName := strings.ToUpper(string(bytes.Trim(nameBytes, "\x00")))
	if summary && verbose {
		fmt.Println()
	}
	thisEntryType, known := KnownFstatEntryTypes[fsbBlob[1]]
	if known {
		fileType = thisEntryType.Desc
		loadIt = thisEntryType.HasPayload
		if thisEntryType.IsDir {
			workingDir = filepath.Join(workingDir, fileName)
			if extract {
				err := os.MkdirAll(workingDir, os.ModePerm)
				if err != nil {
					log.Printf("ERROR: Could not create directory <%s> due to %v", workingDir, err)
					if !ignoreErrors {
						log.Fatalln("Giving up.")
					}
				}
			}
		}
	} else {
		fileType = "Unknown File"
		loadIt = true
	}

	if summary {
		var displayPath string
		if len(workingDir) == 0 {
			displayPath = fileName
		} else {
			displayPath = filepath.Join(workingDir, fileName)
		}
		fmt.Printf("%-20s: %-48s", fileType, displayPath)
		if verbose || (known && thisEntryType.IsDir) {
			fmt.Println()
		} else {
			fmt.Printf("\t")
		}
	}

	if extract && loadIt {
		var writePath string
		if len(workingDir) == 0 {
			writePath = fileName
		} else {
			writePath = filepath.Join(workingDir, fileName)
		}
		if verbose {
			fmt.Printf(" Creating file: '%s'\n", writePath)
		}
		var err error
		writeFile, err = os.Create(writePath)
		if err != nil {
			log.Printf("ERROR: Could not create file %s due to %s", writePath, err.Error())
			if !ignoreErrors {
				log.Fatalln("Giving up.")
			}
		}
	}
	return fileName
}

func readBlob(byteLen int, dumpFile *os.File, desc string) []byte {
	ba := make([]byte, byteLen)
	n, err := dumpFile.Read(ba)
	if n != byteLen || err != nil {
		log.Fatalf("ERROR: Could not read %s record due to %v", desc, err)
	}
	return ba
}

func readAWord(dumpFile *os.File) WordT {
	twoBytes := readBlob(2, dumpFile, "DG Word")
	return WordT(twoBytes[0])<<8 | WordT(twoBytes[1])
}

func readHeader(dumpFile *os.File) recordHeaderT {
	var hdr recordHeaderT
	twoBytes := readBlob(2, dumpFile, "Header")
	hdr.recordType = int(twoBytes[0]) >> 2 // 6-bit
	hdr.recordLength = int(twoBytes[0]&0x03)<<8 + int(twoBytes[1])
	return hdr
}

func readSod(dumpFile *os.File) sodT {
	var sod sodT
	sod.sodHeader = readHeader(dumpFile)
	if sod.sodHeader.recordType != startDumpType {
		log.Fatalln("ERROR: This does not appear to be an AOS/VS DUMP_II or DUMP_III file.")
	}
	sod.dumpFormatRevision = readAWord(dumpFile)
	sod.dumpTimeSecs = readAWord(dumpFile)
	sod.dumpTimeMins = readAWord(dumpFile)
	sod.dumpTimeHours = readAWord(dumpFile)
	sod.dumpTimeDay = readAWord(dumpFile)
	sod.dumpTimeMonth = readAWord(dumpFile)
	sod.dumpTimeYear = readAWord(dumpFile)
	return sod
}
