package main

import (
	"os"
	"golang.org/x/term"
)

type config struct {
    rows int
    cols int
    oldState *term.State
}

type abuf struct {
    b string
    l int
}

var conf config
var ABUF_INIT abuf = abuf{"", 0}

func checkErr(err error){
    if err != nil {
        panic(err)
    }
}

func runOnExit(){
    _, err := os.Stdout.WriteString("\x1b[2J")
    checkErr(err)

    _, err = os.Stdout.WriteString("\x1b[H")
    checkErr(err)

    err = term.Restore(int(os.Stdin.Fd()), conf.oldState)
    checkErr(err)

    os.Exit(0)
}

func ctrlkey(ch byte) byte {
    return ch&0x1f
}

func abAppend(ab *abuf, s string, l int)  {
    ab.b = ab.b+s
    ab.l = ab.l + l
}

func getSize()  {
    cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
    checkErr(err)

    conf.rows, conf.cols = rows, cols
}

func editorDrawRows(ab *abuf)  {
    for i := 0; i < conf.rows; i++ {
        if i == conf.rows / 3 {
            welcome := "Welcome to GoTed version --- 1.0"
            welcomLen := len(welcome)
            abAppend(ab, welcome, welcomLen)
        } else {
            abAppend(ab, "~", 1)
        }

        abAppend(ab, "\x1b[K", 3)
        if i < conf.rows -1 {
            abAppend(ab, "\r\n", 2)
        }
    }
}

func editorRefreshScreen(){
    ab := ABUF_INIT
    abAppend(&ab, "\x1b[?25l", 6)
    abAppend(&ab, "\x1b[H", 4)

    editorDrawRows(&ab)

    abAppend(&ab, "\x1b[H", 4)
    abAppend(&ab, "\x1b[?25h", 6)
    _, err := os.Stdout.WriteString(ab.b)
    checkErr(err)
}

func editorReadKey() byte {
    var buf [1]byte
    _, err := os.Stdin.Read(buf[:])
    checkErr(err)

    return buf[0]
}

func editorProcessKeys() {
    ch := editorReadKey()
    switch ch {
    case ctrlkey('q'):
    runOnExit()
    }
}

func main()  {
    temp, err := term.MakeRaw(int(os.Stdin.Fd()))
    checkErr(err)

    conf.oldState = temp

    for {
        getSize()
        editorRefreshScreen()
        editorProcessKeys()
    }
}
