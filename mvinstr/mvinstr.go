package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

const (
	maxTypes   = 20
	maxFormats = 40
	maxInstrs  = 500
	instrAttrs = 6
)

var (
	window    *widgets.QMainWindow
	widget    *widgets.QWidget
	tabWidget *widgets.QTabWidget

	iiAction *widgets.QAction

	typesList                [maxTypes]string
	formatsList              [maxFormats]string
	typesModel, formatsModel *core.QAbstractListModel
	instrsTable              [maxInstrs][]string
	instrsModel              *core.QAbstractTableModel
	instrsView               *widgets.QTableView
	headers                  = [...]string{"#", "Mnem", "Bits", "BitMask", "Len", "Instruction Format", "Instruction Type"}
	err                      error

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

	editMenu := window.MenuBar().AddMenu2("&Edit")

	iiAction = editMenu.AddAction("Insert &Instruction")
	iiAction.SetDisabled(true)
	iiAction.ConnectTriggered(func(checked bool) { insertInstr() })

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
	instrsView = widgets.NewQTableView(nil)
	instrsModel = core.NewQAbstractTableModel(nil)

	instrsModel.ConnectRowCount(func(parent *core.QModelIndex) int {
		return len(instrsTable)
	})
	instrsModel.ConnectColumnCount(func(parent *core.QModelIndex) int {
		return len(instrsTable[0])
	})
	instrsModel.ConnectData(func(index *core.QModelIndex, instr int) *core.QVariant {
		if index.Row() < numInstrs && index.Column() < instrAttrs && instr == int(core.Qt__DisplayRole) {
			return core.NewQVariant14(instrsTable[index.Row()][index.Column()])
		}
		return core.NewQVariant()
	})
	instrsModel.ConnectHeaderData(headerData)

	instrsView.SetModel(instrsModel)

	tabWidget.AddTab(instrsView, "Instructions")
}

func headerData(section int, orientation core.Qt__Orientation, role int) *core.QVariant {
	if orientation == core.Qt__Horizontal && role == int(core.Qt__DisplayRole) {
		return core.NewQVariant14(headers[section+1])
	}
	if orientation == core.Qt__Vertical && role == int(core.Qt__DisplayRole) {
		return core.NewQVariant14(strconv.Itoa(section)) //headers[section])
	}
	return core.NewQVariant()
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

	// reset data counts
	numTypes = 0
	numInstrs = 0
	numInstrs = 0
	iiAction.SetDisabled(true)

	typesModel.BeginResetModel()
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
		formatsList[numFormats] = line[0]
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
		row := make([]string, 6)
		for c := 0; c < instrAttrs; c++ {
			row[c] = line[c]
		}
		instrsTable[numInstrs] = row
		numInstrs++
	}
	instrsModel.EndResetModel()

	csvFile.Close()
	iiAction.SetEnabled(true)
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

func insertInstr() {
	iiDialog := widgets.NewQDialog(window, core.Qt__Widget)
	iiDialog.SetWindowTitle("MV/Instr - Add new instruction")
	iiLayout := widgets.NewQGridLayout(iiDialog)

	mnemLab := widgets.NewQLabel(iiDialog, 0)
	mnemLab.SetText("Mnemonic")
	iiLayout.AddWidget(mnemLab, 0, 0, core.Qt__AlignVCenter)

	mnemEdit := widgets.NewQLineEdit(iiDialog)
	iiLayout.AddWidget(mnemEdit, 0, 1, core.Qt__AlignVCenter)

	bitsLab := widgets.NewQLabel(iiDialog, 0)
	bitsLab.SetText("Bit Pattern")
	iiLayout.AddWidget(bitsLab, 1, 0, core.Qt__AlignVCenter)

	bitsEdit := widgets.NewQLineEdit(iiDialog)
	iiLayout.AddWidget(bitsEdit, 1, 1, core.Qt__AlignVCenter)

	maskLab := widgets.NewQLabel(iiDialog, 0)
	maskLab.SetText("Bit Mask")
	iiLayout.AddWidget(maskLab, 2, 0, core.Qt__AlignVCenter)

	maskEdit := widgets.NewQLineEdit(iiDialog)
	iiLayout.AddWidget(maskEdit, 2, 1, core.Qt__AlignVCenter)

	lenLab := widgets.NewQLabel(iiDialog, 0)
	lenLab.SetText("Instruction Length")
	iiLayout.AddWidget(lenLab, 3, 0, core.Qt__AlignVCenter)

	lenEdit := widgets.NewQLineEdit(iiDialog)
	iiLayout.AddWidget(lenEdit, 3, 1, core.Qt__AlignVCenter)

	fmtLab := widgets.NewQLabel(iiDialog, 0)
	fmtLab.SetText("OpCode Format")
	iiLayout.AddWidget(fmtLab, 4, 0, core.Qt__AlignVCenter)

	fmtCombo := widgets.NewQComboBox(iiDialog)
	fmtCombo.AddItems(formatsList[:numFormats])
	iiLayout.AddWidget(fmtCombo, 4, 1, core.Qt__AlignVCenter)

	typLab := widgets.NewQLabel(iiDialog, 0)
	typLab.SetText("Instruction Type")
	iiLayout.AddWidget(typLab, 5, 0, core.Qt__AlignVCenter)

	typCombo := widgets.NewQComboBox(iiDialog)
	typCombo.AddItems(typesList[:numTypes])
	iiLayout.AddWidget(typCombo, 5, 1, core.Qt__AlignVCenter)

	buttonBox := widgets.NewQDialogButtonBox(nil)
	buttonBox.AddButton2("Cancel", widgets.QDialogButtonBox__RejectRole)
	buttonBox.AddButton2("Insert", widgets.QDialogButtonBox__AcceptRole)
	buttonBox.ConnectRejected(func() { iiDialog.Reject() })
	buttonBox.ConnectAccepted(func() { iiDialog.Accept() })
	iiLayout.AddWidget(buttonBox, 6, 1, core.Qt__AlignVCenter)

	iiDialog.SetLayout(iiLayout)
	if iiDialog.Exec() != int(widgets.QDialog__Accepted) {
		return
	}
	log.Println("Dialog Accepted")
	log.Printf("New Mnemonic is: %s\n", mnemEdit.Text())
	var afterNew int
	for afterNew = 0; afterNew < numInstrs; afterNew++ {
		if instrsTable[afterNew][0] > mnemEdit.Text() {
			break
		}
	}
	newIx := afterNew
	instrsModel.BeginResetModel()
	for shuffle := numInstrs; shuffle >= newIx; shuffle-- {
		row := make([]string, 6)
		for c := 0; c < instrAttrs; c++ {
			row[c] = instrsTable[shuffle-1][c]
		}
		instrsTable[shuffle] = row
	}
	instrsTable[newIx][0] = mnemEdit.Text()
	instrsTable[newIx][1] = bitsEdit.Text()
	instrsTable[newIx][2] = maskEdit.Text()
	instrsTable[newIx][3] = lenEdit.Text()
	instrsTable[newIx][4] = fmtCombo.CurrentText()
	instrsTable[newIx][5] = typCombo.CurrentText()
	numInstrs++
	instrsModel.EndResetModel()
}
