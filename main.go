package main

import (
	"golang.org/x/term"
	"os"
)

func ctrl_key(ascii byte) byte {
	var mask int8 = 0b00011111
	var code int8 = int8(ascii) & mask
	return byte(code)
}

func die(err error){
	_, _ = os.Stdout.WriteString("\x1b[2J")
	_, _ = os.Stdout.WriteString("\x1b[H")
    editorRefresh()
    os.Exit(1)
}

func termRestore(oldStateIn *term.State, oldStateOut *term.State)  {
    err := term.Restore(int(os.Stdin.Fd()), oldStateIn) 
    if err != nil {
        die(err)
    }

    err = term.Restore(int(os.Stdout.Fd()), oldStateOut) 
    if err != nil {
        die(err)
    }
}

func editorRefresh() {
	_, err := os.Stdout.WriteString("\x1b[2J")
	if err != nil {
		die(err)
	}
	_, err = os.Stdin.WriteString("\x1b[H")
	if err != nil {
		die(err)
	}
}

func editorReadKey() byte {
	var buf []byte = make([]byte, 1)
	_, err := os.Stdin.Read(buf)
	if err != nil {
		die(err)
	}

	return buf[0]
}

func editorProcessKeypress() {
	var char byte = editorReadKey()
	switch char {
	case ctrl_key('q'):
		os.Exit(0)
	}
}

func main() {
    oldStateIn, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		die(err)
	}

    oldStateOut, err := term.GetState(int(os.Stdout.Fd()))
	if err != nil {
		die(err)
	}

	_, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		die(err)
	}
	defer termRestore(oldStateIn, oldStateOut)

	for {
		editorRefresh()
		editorProcessKeypress()
	}
}
