// mvinstr is a command-line utility to check and convert CSV instruction definitions into
// files usable by various emulators.

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

package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	maxTypes   = 20
	maxFormats = 40
	maxInstrs  = 500
	instrAttrs = 6
)

var (
	// command arguments
	actionFlag = flag.String("action", "", "specify operation to perform ie. 'checkbits' or 'makego'")
	csvFlag    = flag.String("csv", "", "CSV file to source data from")
	goFlag     = flag.String("go", "", "Go filename")

	typesList   [maxTypes]string
	formatsList [maxFormats]string
	instrsTable [maxInstrs][]string

	headers = [...]string{"#", "Mnem", "Bits", "BitMask", "Len", "Instruction Format", "Instruction Type"}
	err     error

	numTypes, numFormats, numInstrs int
)

func main() {
	flag.Parse()

	if *csvFlag == "" {
		log.Fatalln("ERROR: Must specify source CSV file with -csv=<csvfile> argument")
	}
	if *actionFlag == "" {
		log.Fatalln("ERROR: Must specify action with -action=<action> argument i.e. checkbits or makego")
	}

	switch *actionFlag {
	case "checkbits":
		if loadCSV() {
			checkBits()
		}
	case "makego":
		if *goFlag == "" {
			log.Fatalln("ERROR: Must specify Go file for output with -go=<gofile> argument")
		}
		if loadCSV() {
			exportGo()
		}
	default:
		log.Fatalln("ERROR: No such action")
	}
}

func loadCSV() bool {

	csvFile, err := os.Open(*csvFlag)
	if err != nil {
		log.Fatalln(err)
	}
	csvReader := csv.NewReader(bufio.NewReader(csvFile))
	line, err := csvReader.Read()
	if line[0] != ";Types" {
		log.Printf("Error: expecting <;Types> got <%s>\n", line[0])
		return false
	}

	// reset data counts
	numTypes = 0
	numInstrs = 0
	numInstrs = 0

	numTypes = 0
	for {
		line, err = csvReader.Read()
		if line[0] == ";" {
			break
		}
		typesList[numTypes] = line[0]
		//log.Printf("Loading type #%d: %s\n", numTypes, line[0])
		numTypes++
	}

	line, err = csvReader.Read()
	if line[0] != ";Formats" {
		log.Printf("Error: expecting <;Formats> got <%s>\n", line[0])
		return false
	}

	numFormats = 0
	for {
		line, err = csvReader.Read()
		if line[0] == ";" {
			break
		}
		formatsList[numFormats] = line[0]
		//log.Printf("Loading format #%d: %s\n", numFormats, line[0])
		numFormats++
	}

	line, err = csvReader.Read()
	if line[0] != ";Instructions" {
		log.Printf("Error: expecting <;Instructions> got <%s>\n", line[0])
		return false
	}

	numInstrs = 0
	for {
		line, err = csvReader.Read()
		if line[0] == ";" {
			break
		}
		row := make([]string, 6)
		for c := 0; c < instrAttrs; c++ {
			row[c] = line[c]
		}
		instrsTable[numInstrs] = row
		numInstrs++
	}

	csvFile.Close()
	return true
}

func exportGo() bool {
	goFile, err := os.Create(*goFlag)
	if err != nil {
		log.Println(err)
		return false
	}
	goWriter := bufio.NewWriter(goFile)

	fmt.Fprintf(goWriter, `// InstructionDefinitions.go

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

// N.B. This file has been AUTOGENERATED by mvinstr...
// ===  ...you probably should not edit it!

package main

`)

	fmt.Fprintf(goWriter, "// Instruction Types\nconst (\n")
	fmt.Fprintf(goWriter, "\t%s = iota\n", typesList[0])
	for t := 1; t < numTypes; t++ {
		fmt.Fprintf(goWriter, "\t%s\n", typesList[t])
	}

	fmt.Fprintf(goWriter, ")\n\n// Instruction Formats\nconst (\n")
	fmt.Fprintf(goWriter, "\t%s = iota\n", formatsList[0])
	for f := 1; f < numFormats; f++ {
		fmt.Fprintf(goWriter, "\t%s\n", formatsList[f])
	}

	fmt.Fprintf(goWriter, ")\n\n// Instruction Mnemonic Consts\nconst (\n")
	fmt.Fprintf(goWriter, "\tinstr%s = iota\n", instrsTable[0][0])
	for i := 1; i < numInstrs; i++ {
		fmt.Fprintf(goWriter, "\tinstr%s\n", instrsTable[i][0])
	}

	fmt.Fprintf(goWriter, ")\n\n// InstructionsInit initialises the instruction characterstics for each instruction(\n")
	fmt.Fprintf(goWriter, "func instructionsInit() {\n")

	for i := 0; i < numInstrs; i++ {
		fmt.Fprintf(goWriter, "\tinstructionSet[instr%s] = instrChars{\"%s\", %s, %s, %s, %s, %s}\n",
			instrsTable[i][0],
			instrsTable[i][0],
			instrsTable[i][1],
			instrsTable[i][2],
			instrsTable[i][3],
			instrsTable[i][4],
			instrsTable[i][5])
	}

	fmt.Fprintf(goWriter, "}\n")
	goWriter.Flush()
	goFile.Close()
	fmt.Println("Go file written")
	return true
}

// checkBits tests every instruction to ensure that (at least) all set bits are covered by the
// associated bit mask
func checkBits() {
	errors := 0
	for i := 0; i < numInstrs; i++ {
		bitsUint, _ := strconv.ParseUint(instrsTable[i][1], 0, 16)
		maskUint, _ := strconv.ParseUint(instrsTable[i][2], 0, 16)
		diff := bitsUint ^ maskUint // XOR
		and := diff & bitsUint
		// fmt.Printf("%d %s %s ", instrsTable[i][0], instrsTable[i][1], instrsTable[i][2])
		// if and == 0 {
		// 	fmt.Printf("OK\n")
		// } else {
		// 	fmt.Println("*** Error ***\n")
		// }
		if and != 0 {
			errors++
			fmt.Printf("Bitmasking error in  %s\n", instrsTable[i][0])
		}
	}
	if errors == 0 {
		fmt.Println("No bitmasking errors detected")
	}
}
