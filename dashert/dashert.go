// dashert.go

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
	"log"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh/terminalT"
)

// Dashert provides minimal DG DASHER terminalT emulation at an ANSI-compatible terminalT (shell).
//
// It is intended for use only where the fully-featured DasherQ or DasherJ terminalT emulators cannot be run
// and should provide just enough compatibility to run a console.
func main() {
	host, port := parseArgs()
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("Error: could not resolve host/port address <%s>:<%s>\n", host, port)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("Error: could not connect to host/port <%s>:<%s>\n", host, port)
	}

	oldState, err := terminalT.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminalT.Restore(0, oldState)

	go remoteListener(conn)
	kbdListener(conn)
}

// ParseArgs is a dumb argument splitter for the couple of required args.
func parseArgs() (host, port string) {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatalln("Error: dashert requires two arguments <host> and <port>\n")
	}
	h := args[0]
	p := args[1]
	return h, p
}

// KbdListener waits for input from the keyboard and sends it to the remote host.
//
// Ctrl-] is used to escape (terminate) the session as per telnet.
func kbdListener(conn *net.TCPConn) {
	var oneCharBuf = make([]byte, 1) // must be 1
	for {
		n, _ := os.Stdin.Read(oneCharBuf)

		switch oneCharBuf[0] {

		case 0x1D: // Ctrl-]
			os.Exit(0)

		default:
			if n > 0 {
				_, err := conn.Write(oneCharBuf)
				if err != nil {
					log.Fatalln("Error: fatal error sending to host")
				}
			}
		}

	}

}

// RemoteListener waits for data from the remote host and displays it on the local screen.
//
// A minimal amount of DASHER-to-ANSI decoding is done to correctly display some character attributes.
// The supported attributes are:
//   Underline
//   Dim (some terminalTs ignore this)
// NewLines are expanded to CR/LF
//
// The supprorted DASHER actions are:
//   Erase EOL
//   Erase Page
//   Write Cursor Address (Position in window)
func remoteListener(conn *net.TCPConn) {
	const ansiEsc byte = 033
	for {
		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			log.Fatalln("Error: fatal error reading host response")
		}
		if n > 0 {
			ansi := make([]byte, 2048)
			for c := 0; c < n; c++ {
				dasherChar := response[c]
				switch dasherChar {
				case 012: // Dasher NL
					ansi = append(ansi, 012, 015)
				case 013: // Erase EOL
					ansi = append(ansi, ansiEsc, '[', 'K')
				case 014: // Erase Page
					ansi = append(ansi, ansiEsc, '[', '2', 'J')
				case 020: // write window address followed by col and row
					row := []byte(strconv.Itoa(int(response[c+2] + 1)))
					col := []byte(strconv.Itoa(int(response[c+1] + 1)))
					c += 2
					ansi = append(ansi, ansiEsc, '[')
					ansi = append(ansi, row...)
					ansi = append(ansi, ';')
					ansi = append(ansi, col...)
					ansi = append(ansi, 'f')
				case 024: // underline on
					ansi = append(ansi, ansiEsc, '[', '4', 'm')
				case 025, 035: // ...off
					ansi = append(ansi, ansiEsc, '[', '0', 'm')
				case 031: // cursor left
					ansi = append(ansi, ansiEsc, '[', '1', 'D')
				case 034: // dim on
					ansi = append(ansi, ansiEsc, '[', '2', 'm')
				default:
					ansi = append(ansi, dasherChar)
				}
			}
			_, err = os.Stdout.Write(ansi)
			if err != nil {
				log.Fatalln("Error: fatal error writing host response to console")
			}
		}
	}
}
