// simhTapeTool is a utility for manipulating SimH-encoded images of tapes for AOS/VS systems.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
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

var (
	createFlag     = flag.String("create", "", "Create a new SimH Tape Image file")
	csvFlag        = flag.Bool("csv", false, "Use/Generate CSV-format data")
	definitionFlag = flag.String("definition", "", "Use a definition file")
	scanFlag       = flag.String("scan", "", "Scan a SimH Tape Image file for correctness")
)

func main() {
	flag.Parse()

	switch {
	case *scanFlag != "":
		fmt.Printf("%s\n", ScanImage(*scanFlag, *csvFlag))
	case *createFlag != "":
		if *csvFlag == false || *definitionFlag == "" {
			log.Fatal("ERROR: Must specify --csv and provide a --definition file to create new image")
		}
		createImage()
	}
}

func createImage() {
	defCSVfile, err := os.Open(*definitionFlag)
	if err != nil {
		log.Fatalf("ERROR: Could not access CSV Definition file %s", *definitionFlag)
	}
	defer defCSVfile.Close()
	csvReader := csv.NewReader(defCSVfile)
	imgFile, err := os.Create(*createFlag)
	if err != nil {
		log.Fatalf("ERROR: Could not create new image file %s", *createFlag)
	}
	for {
		defRec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("ERROR: Could not parse CSV definition file")
		}
		// 1st field of defRec is the src file name, 2nd field is the block size
		thisSrcFile, err := os.Open(defRec[0])
		if err != nil {
			log.Fatalf("ERROR: Could not open input file %s", defRec[0])
		}
		thisBlkSize, err := strconv.Atoi(defRec[1])
		if err != nil {
			log.Fatalf("ERROR: Could not parse block size for input file %s", defRec[0])
		}
		switch thisBlkSize {
		case 2048, 4096, 8192, 16384:
			block := make([]byte, thisBlkSize)
			for {
				bytesRead, err := thisSrcFile.Read(block)
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				if bytesRead > 0 {
					WriteMetaData(imgFile, uint32(bytesRead))
					ok := WriteRecordData(imgFile, block)
					if !ok {
						log.Fatal("ERROR: Error writing image file")
					}
					WriteMetaData(imgFile, uint32(bytesRead))
				}
				if bytesRead == 0 || err == io.EOF { // End of this file
					WriteMetaData(imgFile, SimhMtrTmk)
					break
				}
			} // loop round for next block
		default:
			log.Fatalf("ERROR: Unsupported block size %d for input file %s", thisBlkSize, defRec[0])
		}
	}
	WriteMetaData(imgFile, SimhMtrEom)
	imgFile.Close()
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
		header, ok = ReadRecordHeaderTrailer(imgFile)
		if !ok {
			log.Fatal("Exiting")
		}
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
				log.Fatal("Exiting")
			}
			trailer, ok = ReadRecordHeaderTrailer(imgFile)
			if !ok {
				log.Fatal("Exiting")
			}
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

// ReadRecordHeaderTrailer reads a four byte (one doubleword) header or trailer record
// from the supplised tape image file
func ReadRecordHeaderTrailer(imgFile *os.File) (uint32, bool) {
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

// WriteMetaData writes a 4-byte header/trailer or other metadata
func WriteMetaData(imgFile *os.File, hdr uint32) bool {
	hdrBytes := make([]byte, 4)
	hdrBytes[3] = byte(hdr >> 24)
	hdrBytes[2] = byte(hdr >> 16)
	hdrBytes[1] = byte(hdr >> 8)
	hdrBytes[0] = byte(hdr)
	nb, err := imgFile.Write(hdrBytes)
	if err != nil || nb != 4 {
		log.Fatalf("ERROR: Could not write simh tape header record due to %s\n", err.Error())
	}
	return true
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
