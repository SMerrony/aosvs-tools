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
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const versionString = "1.2a"

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

	baseDir, _ = os.Getwd()
	workingDir = baseDir

	// there should always be a SOD record...
	sod := readSod(dumpFile)
	if summary {
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
		case fsbType:
			//dumpFile.Seek(int64(recHdr.recordLength), 1)
			fsbBlob = make([]byte, recHdr.recordLength)
			n, err := dumpFile.Read(fsbBlob)
			if n != recHdr.recordLength || err != nil {
				log.Fatalf("ERROR: Could not read FSB due to %v", err)
			}
			loadIt = false
		case nbType:
			fileName = processNameBlock(recHdr, fsbBlob, dumpFile)
		case udaType:
			// throw away for now
			udaBlob := make([]byte, recHdr.recordLength)
			n, err := dumpFile.Read(udaBlob)
			if n != recHdr.recordLength || err != nil {
				log.Fatalf("ERROR: Could not read UDA due to %v", err)
			}
		case aclType:
			aclBlob := make([]byte, recHdr.recordLength)
			n, err := dumpFile.Read(aclBlob)
			if n != recHdr.recordLength || err != nil {
				log.Fatalf("ERROR: Could not read ACL due to %v", err)
			}
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

	var (
		dhb dataHeaderT
	)

	// first get the address and length
	fourBytes := make([]byte, 4)
	dumpFile.Read(fourBytes)
	dhb.byteAddress = DwordT(fourBytes[0])<<24 + DwordT(fourBytes[1])<<16 + DwordT(fourBytes[2])<<8 + DwordT(fourBytes[3])
	dumpFile.Read(fourBytes)
	dhb.byteLength = DwordT(fourBytes[0])<<24 + DwordT(fourBytes[1])<<16 + DwordT(fourBytes[2])<<8 + DwordT(fourBytes[3])
	if dhb.byteLength > maxBlockSize {
		log.Fatalf("ERROR: Maximum block size exceeded ( %d vs. limit of %d).", dhb.byteLength, maxBlockSize)
	}

	twoBytes := make([]byte, 2)
	dumpFile.Read(twoBytes)
	dhb.alignmentCount = WordT(twoBytes[0])<<8 + WordT(twoBytes[1])

	if summary && verbose {
		fmt.Printf(" Data block: %d (bytes)\n", dhb.byteLength)
	}

	// skip any alignment bytes - usually just one
	if dhb.alignmentCount > 0 {
		if verbose {
			fmt.Printf("  Skipping %d alignment byte(s)\n", dhb.alignmentCount)
		}
		alignment := make([]byte, dhb.alignmentCount)
		dumpFile.Read(alignment)
	}

	dataBlob := make([]byte, dhb.byteLength)
	n, err := dumpFile.Read(dataBlob)
	if n != int(dhb.byteLength) || err != nil {
		log.Fatalf("ERROR: Could not read data block due to %v", err)
	}

	if extract && writeFile != nil {
		// pad out if block address is beyond end of last block
		if int(dhb.byteAddress) > totalFileSize+1 {
			paddingSize := int(dhb.byteAddress) - totalFileSize
			paddingBlocks := paddingSize / diskBlockBytes
			paddingBlock := make([]byte, diskBlockBytes)
			for p := 0; p < paddingBlocks; p++ {
				writeFile.Write(paddingBlock)
				totalFileSize += diskBlockBytes
			}
		}
		n, err := writeFile.Write(dataBlob)
		if n != int(dhb.byteLength) || err != nil {
			log.Fatalf("ERROR: Could not write out data due to %v", err)
		}
	}

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
		// not in the middle of a file, this must be a directory pop instruction
		// if len(workingDir) > 0 {
		// 	lastSlashPos := strings.LastIndex(workingDir, "/")
		// 	if lastSlashPos != -1 {
		// 		workingDir = workingDir[0:lastSlashPos]
		// 	}
		// }
		if workingDir != baseDir { // don't go up from start dir
			workingDir = filepath.Dir(workingDir)
		}
		if verbose {
			fmt.Printf("Popped dir - new dir is: %s\n", workingDir)
		}
	}
	if verbose {
		fmt.Println("End Block processed")
	}
}

func processLink(recHeader recordHeaderT, linkName string, dumpFile *os.File) {
	linkTargetBA := make([]byte, recHeader.recordLength)
	dumpFile.Read(linkTargetBA)
	linkTargetBA = bytes.Trim(linkTargetBA, "\x00")
	// convert AOS/VS : directory separators to Posix slashes
	linkTarget := strings.ToUpper(strings.Replace(string(linkTargetBA), ":", "/", -1))
	if summary {
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
	var (
		fileType string
	)
	nameBytes := make([]byte, recHeader.recordLength)
	n, err := dumpFile.Read(nameBytes)
	if n != recHeader.recordLength || err != nil {
		log.Fatalf("ERROR: Could not read file name in Name Block in file <%s> due to %v", dumpFile.Name(), err)
	}
	fileName := strings.ToUpper(string(bytes.Trim(nameBytes, "\x00")))
	if summary && verbose {
		fmt.Println()
	}
	switch fsbBlob[1] {
	case flnk:
		fileType = "=>Link=>"
		loadIt = false
	case fdir:
		fileType = "<Directory>"
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
		loadIt = false
	case fstf:
		fileType = "Symbol Table"
		loadIt = true
	case ftxt:
		fileType = "Text file"
		loadIt = true
	case fprg, fprv:
		fileType = "Program File"
		loadIt = true
	default: // we don't explicitly recognise the type
		// TODO: get definitive list from paru.32.sr
		fileType = "File"
		loadIt = true
	}

	if summary {
		var displayPath string
		if len(workingDir) == 0 {
			displayPath = fileName
		} else {
			displayPath = filepath.Join(workingDir, fileName)
		}
		fmt.Printf("%-12s: %-48s", fileType, displayPath)
		if verbose || fsbBlob[1] == fdir {
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
		writeFile, err = os.Create(writePath)
		if err != nil {
			log.Printf("ERROR: Could not create file %s due to %v", writePath, err)
			if !ignoreErrors {
				log.Fatalln("Giving up.")
			}
		}
	}
	return fileName
}

func readAWord(dumpFile *os.File) WordT {
	var w WordT
	twoBytes := make([]byte, 2)
	n, err := dumpFile.Read(twoBytes)
	if n != 2 || err != nil {
		log.Fatalf("ERROR: Could not read Word in file <%s> due to %v", dumpFile.Name(), err)
	}
	w = WordT(twoBytes[0])<<8 | WordT(twoBytes[1])
	return w
}

func readHeader(dumpFile *os.File) recordHeaderT {
	var (
		hdr recordHeaderT
	)
	twoBytes := make([]byte, 2)
	n, err := dumpFile.Read(twoBytes)
	if n != 2 || err != nil {
		log.Fatalf("ERROR: Could not read header in file <%s> due to %v", dumpFile.Name(), err)
	}
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
