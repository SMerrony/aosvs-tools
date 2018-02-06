// simhTapeTool is a utility and various functions for manipulating SimH-encoded images of tapes for AOS/VS systems.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/SMerrony/aosvs-tools/simhTape"
)

var (
	createFlag     = flag.String("create", "", "Create a new SimH Tape Image file")
	csvFlag        = flag.Bool("csv", false, "Use/Generate CSV-format data")
	definitionFlag = flag.String("definition", "", "Use a definition file")
	dumpFlag       = flag.String("dump", "", "Dump all files in image as blobs in current directory")
	scanFlag       = flag.String("scan", "", "Scan a SimH Tape Image file for correctness")
	vFlag          = flag.Bool("v", false, "Be more verbose")
)

func main() {
	flag.Parse()

	switch {
	case *scanFlag != "":
		fmt.Printf("Scanning tape file : %s", *scanFlag)
		fmt.Printf("%s\n", simhTape.ScanImage(*scanFlag, *csvFlag))
	case *createFlag != "":
		if *csvFlag == false || *definitionFlag == "" {
			log.Fatal("ERROR: Must specify --csv and provide a --definition file to create new image")
		}
		createImage()
	case *dumpFlag != "":
		fmt.Println("Dumping files...")
		simhTape.DumpFiles(*dumpFlag)
		fmt.Println("...finished.")
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
		// read a line from the CSV definition file
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
			fmt.Printf("\nAdding file: %s with block size: %d ", defRec[0], thisBlkSize)
			block := make([]byte, thisBlkSize)
			for {
				bytesRead, err := thisSrcFile.Read(block)
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				if bytesRead > 0 {
					simhTape.WriteMetaData(imgFile, uint32(bytesRead)) // block header
					if *vFlag {
						fmt.Printf(" Wrote Header value: %d...", uint32(bytesRead))
					}
					ok := simhTape.WriteRecordData(imgFile, block[0:bytesRead]) // block
					if !ok {
						log.Fatal("ERROR: Error writing image file")
					}
					fmt.Printf(".")
					simhTape.WriteMetaData(imgFile, uint32(bytesRead)) // block trailer
					if *vFlag {
						fmt.Printf(" Wrote Trailer value: %d...", uint32(bytesRead))
					}
				}
				if bytesRead == 0 || err == io.EOF { // End of this file
					thisSrcFile.Close()
					simhTape.WriteMetaData(imgFile, simhTape.SimhMtrTmk)
					if *vFlag {
						fmt.Printf(" EOF: Wrote Tape Mark value: %d...", simhTape.SimhMtrTmk)
					}
					break
				}
			} // loop round for next block
		default:
			log.Fatalf("ERROR: Unsupported block size %d for input file %s", thisBlkSize, defRec[0])
		}
	}
	// // old EOT was 3 zero headers...
	// WriteMetaData(imgFile, 0)
	// WriteMetaData(imgFile, 0)

	simhTape.WriteMetaData(imgFile, simhTape.SimhMtrEom)
	if *vFlag {
		fmt.Printf(" EOM: Wrote Tape Mark value: %d...", simhTape.SimhMtrEom)
	}
	imgFile.Close()
	fmt.Printf("\nDone\n")
}
