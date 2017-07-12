# AOS/VS Tools

This is a small collection of utility programs I have written in Go related to accessing Data General AOS/VS systems and files.

## DasherT

DasherT is a very minimal Dasher terminal emulator for use at the command line.  It provides just enough functionality to act as a console for an MV-class system.

It is only intended for use where it is impossible to use either [DasherQ](https://github.com/SMerrony/DasherQ) or [DasherJ](https://github.com/SMerrony/DasherJ).

## MV/Instr

No use to anyone yet.

## ST Parser

aosvs_st_parser takes an AOS/VS symbol table file (.ST) as its sole argument and emits a text stream of locations and symbol names found in the file.  It might be useful for understanding, documenting or reverse engineering AOS/VS programs where the source code has been lost or is unavailable.
