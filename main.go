package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type config struct {
    cx int
    cy int
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

func initEditor() {
    conf.cx, conf.cy = 0, 0
    getSize()
}

func editorDrawRows(ab *abuf)  {
    for i := 0; i < conf.rows; i++ {
        abAppend(ab, "~", 1)
        if i == conf.rows / 3 {
            welcome := "Welcome to GoTed version --- 1.0"
            welcomLen := len(welcome)
            padding := strings.Repeat(" ", (conf.cols - welcomLen)/2)
            abAppend(ab, padding, (conf.cols - welcomLen)/2)
            abAppend(ab, welcome, welcomLen)
        }

        abAppend(ab, "\x1b[K", 3)
        if i < conf.rows -1 {
            abAppend(ab, "\r\n", 2)
        }
    }
}

func editorMoveCursor(key byte)  {
    switch key {
    case 'h':
        conf.cx--
    case 'j':
        conf.cy--
    case 'k':
        conf.cy++
    case 'l':
        conf.cx--
    }
}

func editorRefreshScreen(){
    ab := ABUF_INIT
    abAppend(&ab, "\x1b[?25l", 6)
    abAppend(&ab, "\x1b[H", 4)

    editorDrawRows(&ab)

    buf := fmt.Sprintf("\x1b[%d;%dH", conf.cx+1, conf.cy+1)
    abAppend(&ab, buf, len(buf))

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


    case 'h':
        fallthrough
    case 'j':
        fallthrough
    case 'k':
        fallthrough
    case 'l':
        editorMoveCursor(ch)
    }
}

func main()  {
    temp, err := term.MakeRaw(int(os.Stdin.Fd()))
    checkErr(err)

    conf.oldState = temp

    initEditor()

    for {
        getSize()
        editorRefreshScreen()
        editorProcessKeys()
    }
}
