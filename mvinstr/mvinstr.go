package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

const (
	maxTypes   = 20
	maxFormats = 40
	maxInstrs  = 500
)

var (
	window              *gtk.Window
	notebook            *gtk.Notebook
	typeEntry           [maxTypes]*gtk.Entry
	formatEntry         [maxFormats]*gtk.Entry
	instrTable          *gtk.Table
	instrEntry          [][5]*gtk.Entry
	instrType, instrFmt []*gtk.ComboBoxText
	typeRenameButton    [maxTypes]*gtk.Button
	formatRenameButton  [maxFormats]*gtk.Button
	insertInstrButton   *gtk.Button
	// instrSaveButton                 [maxInstrs]*gtk.Button
	numTypes, numFormats, numInstrs int
)

func main() {
	gtk.Init(&os.Args)
	window = gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle("MV/Instr - MV Instruction Set Maintenance")
	window.Connect("destroy", gtk.MainQuit)

	vbox := gtk.NewVBox(false, 1)

	instrEntry = make([][5]*gtk.Entry, maxInstrs)
	instrType = make([]*gtk.ComboBoxText, maxInstrs)
	instrFmt = make([]*gtk.ComboBoxText, maxInstrs)

	menuBar := gtk.NewMenuBar()
	vbox.PackStart(menuBar, false, false, 0)
	populateMenus(menuBar)

	notebook = gtk.NewNotebook()
	createTypeFrame()
	createFormatFrame()
	createInstrFrame()
	vbox.Add(notebook)
	window.Add(vbox)
	window.SetSizeRequest(800, 600)
	window.ShowAll()

	gtk.Main()
}

func populateMenus(menuBar *gtk.MenuBar) {
	fileMenu := gtk.NewMenuItemWithMnemonic("_File")
	menuBar.Append(fileMenu)
	subMenu := gtk.NewMenu()
	fileMenu.SetSubmenu(subMenu)

	var menuItem *gtk.MenuItem

	menuItem = gtk.NewMenuItemWithMnemonic("_Load CSV")
	menuItem.Connect("activate", loadCSV)
	subMenu.Append(menuItem)

	menuItem = gtk.NewMenuItemWithMnemonic("_Save CSV")
	menuItem.Connect("activate", saveCSV)
	subMenu.Append(menuItem)

	menuItem = gtk.NewMenuItemWithMnemonic("Export for _C")
	subMenu.Append(menuItem)

	menuItem = gtk.NewMenuItemWithMnemonic("Export for _Go")
	menuItem.Connect("activate", exportGo)
	subMenu.Append(menuItem)

	menuItem = gtk.NewMenuItemWithMnemonic("E_xit")
	menuItem.Connect("activate", func() {
		gtk.MainQuit()
	})
	subMenu.Append(menuItem)
}

func createTypeFrame() {
	page := gtk.NewFrame("Types")
	notebook.AppendPage(page, gtk.NewLabel("Types"))
	typesScroller := gtk.NewScrolledWindow(nil, nil)
	typesScroller.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	typesTable := gtk.NewTable(20, 3, false)
	for y := uint(0); y < maxTypes; y++ {
		typesTable.Attach(gtk.NewLabel(strconv.Itoa(int(y))), 0, 1, y, y+1, gtk.FILL, gtk.FILL, 5, 5)
		typeEntry[y] = gtk.NewEntry()
		typesTable.Attach(typeEntry[y], 1, 2, y, y+1, gtk.FILL, gtk.FILL, 5, 5)
		typeRenameButton[y] = gtk.NewButtonWithLabel("Rename")
		typeRenameButton[y].Connect(
			"clicked",
			func(ctx *glib.CallbackContext) {
				renameType(ctx)
			},
			y)
		typeRenameButton[y].SetSensitive(false)
		// enable the rename button when data changes
		typeEntry[y].Connect(
			"changed",
			func(ctx *glib.CallbackContext) {
				typeRenameButton[ctx.Data().(uint)].SetSensitive(true)
			},
			y)
		typesTable.Attach(typeRenameButton[y], 2, 3, y, y+1, gtk.FILL, gtk.FILL, 5, 5)
	}
	typesScroller.AddWithViewPort(typesTable)
	page.Add(typesScroller)
}

func createFormatFrame() {
	page := gtk.NewFrame("Formats")
	notebook.AppendPage(page, gtk.NewLabel("Formats"))
	fmtScroller := gtk.NewScrolledWindow(nil, nil)
	fmtScroller.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	fmtTable := gtk.NewTable(maxFormats, 3, false)
	for y := uint(0); y < maxFormats; y++ {
		fmtTable.Attach(gtk.NewLabel(strconv.Itoa(int(y))), 0, 1, y, y+1, gtk.FILL, gtk.FILL, 5, 5)
		formatEntry[y] = gtk.NewEntry()
		formatEntry[y].SetWidthChars(32)
		fmtTable.Attach(formatEntry[y], 1, 2, y, y+1, gtk.FILL, gtk.FILL, 5, 5)
		formatRenameButton[y] = gtk.NewButtonWithLabel("Rename")
		formatRenameButton[y].Connect(
			"clicked",
			func(ctx *glib.CallbackContext) {
				renameFormat(ctx)
			},
			y)
		formatRenameButton[y].SetSensitive(false)
		// enable the rename button when data changes
		formatEntry[y].Connect(
			"changed",
			func(ctx *glib.CallbackContext) {
				formatRenameButton[ctx.Data().(uint)].SetSensitive(true)
			},
			y)
		fmtTable.Attach(formatRenameButton[y], 2, 3, y, y+1, gtk.FILL, gtk.FILL, 5, 5)
	}
	fmtScroller.AddWithViewPort(fmtTable)
	page.Add(fmtScroller)
}

func createInstrFrame() {
	page := gtk.NewFrame("Instructions")
	notebook.AppendPage(page, gtk.NewLabel("Instructions"))
	instrScroller := gtk.NewScrolledWindow(nil, nil)
	instrScroller.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	instrTable = gtk.NewTable(maxInstrs, 7, false)
	instrTable.SetRowSpacings(1)
	for y := uint(0); y < maxInstrs; y++ {
		instrTable.Attach(gtk.NewLabel(strconv.Itoa(int(y))), 0, 1, y, y+1, gtk.FILL, gtk.FILL, 2, 2)
		instrEntry[y][0] = gtk.NewEntry()
		instrEntry[y][0].SetWidthChars(8)
		instrTable.Attach(instrEntry[y][0], 1, 2, y, y+1, gtk.FILL, gtk.FILL, 2, 2)
		instrEntry[y][1] = gtk.NewEntry()
		instrEntry[y][1].SetWidthChars(6)
		instrTable.Attach(instrEntry[y][1], 2, 3, y, y+1, gtk.FILL, gtk.FILL, 2, 2)
		instrEntry[y][2] = gtk.NewEntry()
		instrEntry[y][2].SetWidthChars(6)
		instrTable.Attach(instrEntry[y][2], 3, 4, y, y+1, gtk.FILL, gtk.FILL, 2, 2)
		instrEntry[y][3] = gtk.NewEntry()
		instrEntry[y][3].SetWidthChars(2)
		instrTable.Attach(instrEntry[y][3], 4, 5, y, y+1, gtk.FILL, gtk.FILL, 2, 2)
		instrFmt[y] = gtk.NewComboBoxText()
		instrFmt[y].AppendText("DUMMY_INSTR_FORMAT")
		instrTable.Attach(instrFmt[y], 5, 6, y, y+1, gtk.FILL, gtk.FILL, 2, 2)
		instrType[y] = gtk.NewComboBoxText()
		instrType[y].AppendText("DUMMY_INSTR_TYPE")
		instrTable.Attach(instrType[y], 6, 7, y, y+1, gtk.FILL, gtk.FILL, 2, 2)
	}
	instrScroller.AddWithViewPort(instrTable)
	page.Add(instrScroller)
}

func loadCSV() {
	dialog := gtk.NewFileChooserDialog(
		"Open CSV File",
		window,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		gtk.STOCK_CANCEL,
		gtk.RESPONSE_CANCEL,
		gtk.STOCK_OPEN,
		gtk.RESPONSE_ACCEPT)
	filter := gtk.NewFileFilter()
	filter.AddPattern("*.csv")
	dialog.AddFilter(filter)
	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		csvFilename := dialog.GetFilename()
		dialog.Destroy()
		csvFile, err := os.Open(csvFilename)
		if err != nil {
			md := gtk.NewMessageDialog(window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Could not open CSV file")
			md.Run()
			md.Destroy()
			return
		}
		csvReader := csv.NewReader(bufio.NewReader(csvFile))
		line, err := csvReader.Read()
		if line[0] != ";Types" {
			log.Printf("Error: expecting <;Types> got <%s>\n", line[0])
			md := gtk.NewMessageDialog(window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Wrong format CSV file")
			md.Run()
			md.Destroy()
			return
		}

		for {
			line, err = csvReader.Read()
			if line[0] == ";" {
				break
			}
			typeEntry[numTypes].SetText(line[0])
			typeRenameButton[numTypes].SetSensitive(false) // disable the button until data changes
			// now update combos in the Instruction tab
			for i := 0; i < maxInstrs; i++ {
				instrType[i].InsertText(numTypes, line[0])
			}
			numTypes++
		}

		line, err = csvReader.Read()
		if line[0] != ";Formats" {
			log.Printf("Error: expecting <;Formats> got <%s>\n", line[0])
			md := gtk.NewMessageDialog(window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Wrong format CSV file")
			md.Run()
			md.Destroy()
			return
		}

		for {
			line, err = csvReader.Read()
			if line[0] == ";" {
				break
			}
			formatEntry[numFormats].SetText(line[0])
			formatRenameButton[numFormats].SetSensitive(false) // disable the button until data changes
			// now update combos in the Instruction tab
			for i := 0; i < maxInstrs; i++ {
				instrFmt[i].InsertText(numFormats, line[0])
			}
			numFormats++
		}

		line, err = csvReader.Read()
		if line[0] != ";Instructions" {
			log.Printf("Error: expecting <;Instructions> got <%s>\n", line[0])
			md := gtk.NewMessageDialog(window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Wrong format CSV file")
			md.Run()
			md.Destroy()
			return
		}
		numInstrs = 0
		for {
			line, err = csvReader.Read()
			if line[0] == ";" {
				break
			}

			for c := 0; c < 4; c++ {
				instrEntry[numInstrs][c].SetText(line[c])
			}
			f, _ := strconv.Atoi(line[4])
			instrFmt[numInstrs].SetActive(f)
			t, _ := strconv.Atoi(line[5])
			instrType[numInstrs].SetActive(t)
			numInstrs++
		}
		insertInstrButton = gtk.NewButtonWithLabel("Insert")
		instrTable.Attach(insertInstrButton, 7, 8, uint(numInstrs), uint(numInstrs)+1, gtk.FILL, gtk.FILL, 1, 1)
		insertInstrButton.Connect(
			"clicked",
			func(ctx *glib.CallbackContext) {
				insertInstruction(ctx)
			},
			numInstrs)
		insertInstrButton.Show()
		//instrTable.Resize(uint(numInstrs+1), 7)

		csvFile.Close()
	}
	dialog.Destroy()
}

func saveCSV() {
	dialog := gtk.NewFileChooserDialog("Save CSV File",
		window, gtk.FILE_CHOOSER_ACTION_SAVE,
		gtk.STOCK_CANCEL, gtk.RESPONSE_CANCEL,
		gtk.STOCK_SAVE, gtk.RESPONSE_ACCEPT)
	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		csvFilename := dialog.GetFilename()
		csvFile, err := os.Create(csvFilename)
		if err != nil {
			md := gtk.NewMessageDialog(window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Could not create CSV file")
			md.Run()
		}
		csvWriter := bufio.NewWriter(csvFile)
		fmt.Fprintf(csvWriter, ";Types\n")
		for t := 0; t < numTypes; t++ {
			fmt.Fprintf(csvWriter, "%s\n", typeEntry[t].GetText())
		}
		fmt.Fprintf(csvWriter, ";\n;Formats\n")
		for f := 0; f < numFormats; f++ {
			fmt.Fprintf(csvWriter, "%s\n", formatEntry[f].GetText())
		}
		fmt.Fprintf(csvWriter, ";\n;Instructions\n")
		for i := 0; i < numInstrs; i++ {
			fmt.Fprintf(csvWriter, "%s,%s,%s,%s,%d,%d\n",
				instrEntry[i][0].GetText(),
				instrEntry[i][1].GetText(),
				instrEntry[i][2].GetText(),
				instrEntry[i][3].GetText(),
				instrFmt[i].GetActive(),
				instrType[i].GetActive())
		}

		fmt.Fprintf(csvWriter, ";\n")
		csvWriter.Flush()
		csvFile.Close()
	}
	dialog.Destroy()
}

func exportGo() {
	dialog := gtk.NewFileChooserDialog("Save Go language File",
		window, gtk.FILE_CHOOSER_ACTION_SAVE,
		gtk.STOCK_CANCEL, gtk.RESPONSE_CANCEL,
		gtk.STOCK_SAVE, gtk.RESPONSE_ACCEPT)
	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		goFilename := dialog.GetFilename()
		goFile, err := os.Create(goFilename)
		if err != nil {
			md := gtk.NewMessageDialog(window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Could not create Go file")
			md.Run()
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
		fmt.Fprintf(goWriter, "\t%s = iota\n", typeEntry[0].GetText())
		for t := 1; t < numTypes; t++ {
			fmt.Fprintf(goWriter, "\t%s\n", typeEntry[t].GetText())
		}
		fmt.Fprintf(goWriter, ")\n\n// Instruction Formats\nconst (\n")
		fmt.Fprintf(goWriter, "\t%s = iota\n", formatEntry[0].GetText())
		for f := 1; f < numFormats; f++ {
			fmt.Fprintf(goWriter, "\t%s\n", formatEntry[f].GetText())
		}
		fmt.Fprintf(goWriter, ")\n\n// InstructionsInit initialises the instruction characterstics for each instruction(\n")
		fmt.Fprintf(goWriter, "func instructionsInit() {\n")

		for i := 0; i < numInstrs; i++ {
			fmt.Fprintf(goWriter, "\tinstructionSet[\"%s\"] = instrChars{%s, %s, %s, %s, %s}\n",
				instrEntry[i][0].GetText(),
				instrEntry[i][1].GetText(),
				instrEntry[i][2].GetText(),
				instrEntry[i][3].GetText(),
				instrFmt[i].GetActiveText(),
				instrType[i].GetActiveText())
		}

		fmt.Fprintf(goWriter, "}\n")
		goWriter.Flush()
		goFile.Close()
	}
	dialog.Destroy()
}

func insertInstruction(ctx *glib.CallbackContext) {
	log.Printf("Not yet...(%d)\n", ctx.Data().(uint))
}

func renameFormat(ctx *glib.CallbackContext) {
	log.Printf("Not yet...(%d)\n", ctx.Data().(uint))
}

func renameType(ctx *glib.CallbackContext) {
	log.Printf("Not yet...(%d)\n", ctx.Data().(uint))
}
