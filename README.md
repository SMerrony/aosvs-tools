# AOS/VS Tools

This is a small collection of utility programs I have written in Go related to accessing Data General AOS/VS systems and files.

## DasherG
DasherG will be a port of my [DasherQ](https://github.com/SMerrony/DasherQ) terminal emulator written in Go and (currently) using the [Go-GTK](https://github.com/mattn/go-gtk) GUI toolkit.  If all goes well I shall retire DasherJ once this is working.

*Work-in-Progress*

## DasherT
DasherT is a very minimal Dasher terminal emulator for use at the command line.  It provides just enough functionality to act as a console for an MV-class system.

It is only intended for use where it is impossible to use either [DasherQ](https://github.com/SMerrony/DasherQ) or [DasherJ](https://github.com/SMerrony/DasherJ).

## MV/Instr
No use to anyone yet.

## SimhTape
simhTape is a Go package that provides some low-level functions for handling SimH-standard tape file images.  Functions available include...
 * ReadMetaData and WriteMetaData for handling headers, trailers, and inter-file gaps
 * ReadRecordData and WriteRecordData for handling data blocks (without their associated headers and trailers)
 * Rewind and SpaceFwd for positioning the virtual tape image
 * ScanImage for examining/verifying a tape image file

## SimhTapeTool
simhTapeTool provides a command-line utility for handling SimH-compatible images of AOS/VS tapes.  Images may be tested for structural validity (-scan) and created (-create) using a simple CSV file to specify the contents of the tape where each line contains a filename and a block-size. 

Uses the simhTape package above.

## ST Parser
aosvs_st_parser takes an AOS/VS symbol table file (.ST) as its sole argument and emits a text stream of locations and symbol names found in the file.  It might be useful for understanding, documenting or reverse engineering AOS/VS programs where the source code has been lost or is unavailable.
