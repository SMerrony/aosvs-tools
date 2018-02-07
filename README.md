# AOS/VS Tools

This is a small collection of utility programs I have written in Go related to accessing Data General AOS/VS systems and files.

## DasherG
DasherG terminal emulator has moved to its own repo  [DasherG](https://github.com/SMerrony/DasherG)

*Work-in-Progress*

## DasherT
DasherT is a very minimal Dasher terminal emulator for use at the command line.  It provides just enough functionality to act as a console for an MV-class system.

It is only intended for emergency use where it is impossible to use either [DasherG](https://github.com/SMerrony/DasherG) or [DasherQ](https://github.com/SMerrony/DasherQ), maybe because you cannot build GUI applications or run their binaries.

## LoadG
LoadG loads (restores) AOS/VS DUMP_II, and maybe DUMP_III, files on any desktop system supported by Go.  It can be used to rescue data from legacy AOS/VS systems if the dumps are accessible on a modern system.  The current version handles at least versions 15 and 16 of the DUMP format.

## MV/Instr
No use to anyone yet.

## ST Parser
aosvs_st_parser takes an AOS/VS symbol table file (.ST) as its sole argument and emits a text stream of locations and symbol names found in the file.  It might be useful for understanding, documenting or reverse engineering AOS/VS programs where the source code has been lost or is unavailable.
