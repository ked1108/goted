package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

type config struct {
    rows int
    cols int
    oldState *term.State
}

var conf config

func runOnExit(){
    err := term.Restore(int(os.Stdin.Fd()), conf.oldState)
    if err != nil {
        fmt.Println("Still Stuck")
    }
    os.Exit(0)
}

func ctrlkey(ch byte) byte {
    return ch&0x1f
}

func getSize()  {
    r, c, err := term.GetSize(int(os.Stdin.Fd()))
    if err != nil{
        panic(err)
    }
    conf.rows, conf.cols = r, c
}

func editorRefreshScreen(){
    _, err := os.Stdout.WriteString("\x1b[2J")
    if err != nil {
        panic(err)
    }

    _, err = os.Stdout.WriteString("\x1b[H")
    if err != nil {
        panic(err)
    }
    
}

func editorReadKey()  {
    var buf [1]byte
    _, err := os.Stdin.Read(buf[:])
    if err != nil {
        panic(err)
    }

    if(buf[0] == ctrlkey('q')){
        runOnExit()
    }
}

func main()  {
    temp, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        panic(err)
    }
    conf.oldState = temp

    for {
        editorRefreshScreen()
        editorReadKey()
    }
}
