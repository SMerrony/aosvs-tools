/* aosvs_st_parser.go
   ==================

   Parse an AOS/VS .ST file and emit info useful for reverse-engineering etc.

   S.Merrony 20170609 - First version

*/
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	args := os.Args[1:]
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Printf("Usage: %s -h|--help|SYMTABFILE.ST\n", os.Args[0])
		return
	}

	fileName := args[0]
	if !strings.HasSuffix(fileName, "ST") {
		fmt.Println("Error: Symbol Table file must have .ST suffix")
		return
	}

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Error while opening file ", err)
	}

	bslice := make([]byte, 256)

	for {

		// advance to next symbol, marked by an 0x20
		for bslice[0] != 0x20 {
			bslice = readBytes(file, 1)
		}

		// read 1 more byte
		bslice = readBytes(file, 1)
		symLength := int(bslice[0])

		// 4 byte address + 14-byte filler
		bslice = readBytes(file, 18)
		symAddr := binary.BigEndian.Uint32(bslice) //int(bslice[3])

		// symbol name
		bslice = readBytes(file, symLength)
		symName := string(bslice[:symLength])

		fmt.Printf("%d %s\n", symAddr, symName)
	}
}

func readBytes(file *os.File, nBytes int) []byte {
	bytes := make([]byte, nBytes)
	_, err := file.Read(bytes)
	if err != nil {
		if err == io.EOF {
			// fmt.Printf("\n=== EOF ===\n")
			os.Exit(0)
		}
		log.Fatal("Error reading file ", err)
	}
	return bytes
}
