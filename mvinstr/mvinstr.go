package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

const (
	maxTypes   = 20
	maxFormats = 40
	maxInstrs  = 500
)

var (
	window    *widgets.QMainWindow
	widget    *widgets.QWidget
	tabWidget *widgets.QTabWidget

	typesList, formatsList   []string
	typesModel, formatsModel *core.QAbstractListModel
	instrsTable              [][]string
	instrsModel              *core.QAbstractTableModel
	instrsView               *widgets.QTableView

	err error

	numTypes, numFormats, numInstrs int
)

func main() {
	widgets.NewQApplication(len(os.Args), os.Args)
	window = widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("MV/Instr - MV Instruction Set Maintenance")
	window.SetMinimumSize2(800, 600)

	tabWidget = widgets.NewQTabWidget(nil)

	populateMenus()

	createTypeFrame()
	createFormatFrame()
	createInstrFrame()

	window.SetCentralWidget(tabWidget)
	window.Show()
	widgets.QApplication_Exec()
}

func populateMenus() {
	fileMenu := window.MenuBar().AddMenu2("&File")

	lc := fileMenu.AddAction("&Load CSV")
	lc.ConnectTriggered(func(checked bool) { loadCSV() })

	sc := fileMenu.AddAction("&Save CSV")
	sc.ConnectTriggered(func(checked bool) { saveCSV() })

	ec := fileMenu.AddAction("Export for &C")
	_ = ec

	eg := fileMenu.AddAction("Export for &Go")
	eg.ConnectTriggered(func(checked bool) { exportGo() })

	em := fileMenu.AddAction("&Quit")
	em.ConnectTriggered(func(checked bool) { window.Close() })
}

func createTypeFrame() {
	//typesList = []string{"This", "is", "a", "test"}
	typesView := widgets.NewQListView(nil)
	typesModel = core.NewQAbstractListModel(nil)
	typesModel.ConnectRowCount(func(parent *core.QModelIndex) int {
		return len(typesList)
	})
	typesModel.ConnectData(func(index *core.QModelIndex, typ int) *core.QVariant {
		if typ != int(core.Qt__DisplayRole) {
			return core.NewQVariant()
		}
		return core.NewQVariant14(typesList[index.Row()])
	})
	typesView.SetModel(typesModel)
	//typesView.SetModel(core.NewQStringListModel2(typesList, nil))
	tabWidget.AddTab(typesView, "Types")
}

func createFormatFrame() {
	formatsView := widgets.NewQListView(nil)
	formatsModel = core.NewQAbstractListModel(nil)
	formatsModel.ConnectRowCount(func(parent *core.QModelIndex) int {
		return len(formatsList)
	})
	formatsModel.ConnectData(func(index *core.QModelIndex, fmt int) *core.QVariant {
		if fmt != int(core.Qt__DisplayRole) {
			return core.NewQVariant()
		}
		return core.NewQVariant14(formatsList[index.Row()])
	})
	formatsView.SetModel(formatsModel)
	tabWidget.AddTab(formatsView, "Formats")

}

func createInstrFrame() {
	instrsTable = [][]string{0: {"", "", "", "", "", "", ""}}
	instrsView = widgets.NewQTableView(nil)
	instrsModel = core.NewQAbstractTableModel(nil)

	instrsModel.ConnectRowCount(func(parent *core.QModelIndex) int {
		return len(instrsTable)
	})
	instrsModel.ConnectColumnCount(func(parent *core.QModelIndex) int {
		return len(instrsTable[0])
	})
	instrsModel.ConnectData(func(index *core.QModelIndex, instr int) *core.QVariant {
		if instr != int(core.Qt__DisplayRole) {
			return core.NewQVariant()
		}
		return core.NewQVariant14(instrsTable[index.Row()][index.Column()])
	})
	instrsView.SetModel(instrsModel)
	tabWidget.AddTab(instrsView, "Instructions")
}

func loadCSV() {

	fileDialog := widgets.NewQFileDialog2(window,
		"Open CSV File",
		"",
		"*.csv")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptOpen)
	fileDialog.SetFileMode(widgets.QFileDialog__ExistingFile)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}

	csvFilename := fileDialog.SelectedFiles()[0]

	csvFile, err := os.Open(csvFilename)
	if err != nil {
		widgets.QMessageBox_Warning(window,
			"Error", "Could not open CSV file",
			widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)
		return
	}
	csvReader := csv.NewReader(bufio.NewReader(csvFile))
	line, err := csvReader.Read()
	if line[0] != ";Types" {
		log.Printf("Error: expecting <;Types> got <%s>\n", line[0])
		widgets.QMessageBox_Warning(window,
			"Error", "Wrong format CSV file",
			widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)
		return
	}

	typesModel.BeginResetModel()
	numTypes = 0
	for {
		line, err = csvReader.Read()
		if line[0] == ";" {
			break
		}
		typesList = append(typesList, line[0])
		//log.Printf("Loading type #%d: %s\n", numTypes, line[0])
		numTypes++
	}
	typesModel.EndResetModel()
	line, err = csvReader.Read()
	if line[0] != ";Formats" {
		log.Printf("Error: expecting <;Formats> got <%s>\n", line[0])
		widgets.QMessageBox_Warning(window,
			"Error", "Wrong format CSV file",
			widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)
		return
	}

	formatsModel.BeginResetModel()
	numFormats = 0
	for {
		line, err = csvReader.Read()
		if line[0] == ";" {
			break
		}
		formatsList = append(formatsList, line[0])
		//log.Printf("Loading format #%d: %s\n", numFormats, line[0])
		numFormats++
	}
	formatsModel.EndResetModel()

	line, err = csvReader.Read()
	if line[0] != ";Instructions" {
		log.Printf("Error: expecting <;Instructions> got <%s>\n", line[0])
		widgets.QMessageBox_Warning(window,
			"Error", "Wrong format CSV file",
			widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)
		return
	}

	instrsModel.BeginResetModel()
	numInstrs = 0
	for {
		line, err = csvReader.Read()
		if line[0] == ";" {
			break
		}
		row := make([]string, 7)
		for c := 0; c < 6; c++ {
			row[c] = line[c]
		}
		if numInstrs == 0 {
			instrsTable[0] = row
		} else {
			instrsTable = append(instrsTable, row)
		}
		numInstrs++
	}

	instrsView.ResizeColumnsToContentsDefault()
	instrsModel.EndResetModel()
	csvFile.Close()

}

func saveCSV() {
	fileDialog := widgets.NewQFileDialog2(nil,
		"Save CSV File",
		"",
		"*.csv")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	csvFilename := fileDialog.SelectedFiles()[0]

	csvFile, err := os.Create(csvFilename)
	if err != nil {
		widgets.QMessageBox_Warning(window,
			"Error", "Could not create CSV file",
			widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)
		return
	}
	csvWriter := bufio.NewWriter(csvFile)
	fmt.Fprintf(csvWriter, ";Types\n")
	for t := 0; t < numTypes; t++ {
		fmt.Fprintf(csvWriter, "%s\n", typesList[t])
	}
	fmt.Fprintf(csvWriter, ";\n;Formats\n")
	for f := 0; f < numFormats; f++ {
		fmt.Fprintf(csvWriter, "%s\n", formatsList[f])
	}
	fmt.Fprintf(csvWriter, ";\n;Instructions\n")
	for i := 0; i < numInstrs; i++ {
		fmt.Fprintf(csvWriter, "%s,%s,%s,%s,%s,%s\n",
			instrsTable[i][0],
			instrsTable[i][1],
			instrsTable[i][2],
			instrsTable[i][3],
			instrsTable[i][4],
			instrsTable[i][5])
	}

	fmt.Fprintf(csvWriter, ";\n")
	csvWriter.Flush()
	csvFile.Close()
	widgets.QMessageBox_Information(window,
		"MV/Instr", "CSV file written",
		widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)

}

func exportGo() {
	fileDialog := widgets.NewQFileDialog2(nil,
		"Save Go language File",
		"",
		"*.go")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	goFilename := fileDialog.SelectedFiles()[0]

	goFile, err := os.Create(goFilename)
	if err != nil {
		widgets.QMessageBox_Warning(window,
			"Error", "Could not create Go file",
			widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)
		return
	}
	goWriter := bufio.NewWriter(goFile)

	fmt.Fprintf(goWriter, `// InstructionDefinitions.go

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
	fmt.Fprintf(goWriter, ")\n\n// InstructionsInit initialises the instruction characterstics for each instruction(\n")
	fmt.Fprintf(goWriter, "func instructionsInit() {\n")

	for i := 0; i < numInstrs; i++ {
		fmt.Fprintf(goWriter, "\tinstructionSet[\"%s\"] = instrChars{%s, %s, %s, %s, %s}\n",
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
	widgets.QMessageBox_Information(window,
		"MV/Instr", "Go file written",
		widgets.QMessageBox__Close, widgets.QMessageBox__NoButton)
}
