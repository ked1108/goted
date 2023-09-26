package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
)

type erow struct {
	size int
	buf  string
}

type abuf struct {
	b string
	l int
}

type config struct {
	cx       int
	cy       int
	rowoff   int
	rows     int
	cols     int
	numrows  int
	row      []erow
	oldState *term.State
}

var conf config
var ABUF_INIT abuf = abuf{"", 0}

//goland:noinspection ALL
const (
	ARROW_LEFT int = iota + 1000
	ARROW_RIGHT
	ARROW_UP
	ARROW_DOWN
	DEL_KEY
	HOME_KEY
	END_KEY
	PAGE_UP
	PAGE_DOWN
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func runOnExit() {
	_, err := os.Stdout.WriteString("\x1b[2J")
	checkErr(err)

	_, err = os.Stdout.WriteString("\x1b[H")
	checkErr(err)

	err = term.Restore(int(os.Stdin.Fd()), conf.oldState)
	checkErr(err)

	os.Exit(0)
}

func ctrlkey(ch byte) int {
	return int(ch & 0x1f)
}

func abAppend(ab *abuf, s string, l int) {
	ab.b = ab.b + s
	ab.l = ab.l + l
}

func getSize() {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	checkErr(err)

	conf.rows, conf.cols = rows, cols
}

func initEditor() {
	conf.cx, conf.cy = 0, 0
	conf.rowoff = 0
	conf.numrows = 0
	conf.row = nil
	getSize()
}

func editorAppendRow(line string, lineLen int) {
	conf.row = append(conf.row, erow{lineLen, line})
	conf.numrows++
}

func editorOpen(filename string) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	checkErr(err)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		editorAppendRow(line, len(line))
	}
	err = scanner.Err()
	checkErr(err)
	err = file.Close()
	checkErr(err)
}

func editorDrawRows(ab *abuf) {
	for j := 0; j < conf.rows; j++ {
		filerow := j + conf.rowoff
		if filerow >= conf.numrows {
			if conf.numrows == 0 && j == conf.rows/3 {
				welcome := "Welcome to GoTed version --- 1.0"
				welcomeLen := len(welcome)
				padding := strings.Repeat(" ", (conf.cols-welcomeLen)/2)
				abAppend(ab, "~", 1)
				abAppend(ab, padding, (conf.cols-welcomeLen)/2)
				abAppend(ab, welcome, welcomeLen)
			} else {
				abAppend(ab, "~", 1)
			}

			abAppend(ab, "\x1b[K", 3)
			if j < conf.rows-1 {
				abAppend(ab, "\r\n", 2)
			}
		} else {
			l := conf.row[filerow].size
			if l > conf.cols {
				l = conf.cols
			}
			abAppend(ab, "\x1b[K", 3)
			abAppend(ab, conf.row[filerow].buf, l)
			if j < conf.rows-1 {
				abAppend(ab, "\r\n", 2)
			}
		}

	}
}

func editorMoveCursor(key int) {
	switch key {
	case ARROW_LEFT:
		if conf.cx > 0 {
			conf.cx--
		}
	case ARROW_UP:
		if conf.cy > 0 {
			conf.cy--
		}
	case ARROW_DOWN:
		if conf.cy < conf.numrows {
			conf.cy++
		}
	case ARROW_RIGHT:
		if conf.cx < conf.cols-1 {
			conf.cx++
		}
	}
}

func editorScroll() {
	if conf.cy < conf.rowoff {
		conf.rowoff = conf.cy
	}
	if conf.cy >= conf.rowoff+conf.rows {
		conf.rowoff = conf.cy - conf.rows + 1
	}
}

func editorRefreshScreen() {
	editorScroll()
	ab := ABUF_INIT
	abAppend(&ab, "\x1b[?25l", 6)
	abAppend(&ab, "\x1b[H", 4)

	editorDrawRows(&ab)

	buf := fmt.Sprintf("\x1b[%d;%dH", (conf.cy-conf.rowoff)+1, conf.cx+1)
	abAppend(&ab, buf, len(buf))

	abAppend(&ab, "\x1b[?25h", 6)
	_, err := os.Stdout.WriteString(ab.b)
	checkErr(err)
}

func editorReadKey() int {
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	checkErr(err)
	char := buf[0]
	if char == '\x1b' {
		var buf [3]byte
		s, err := os.Stdin.Read(buf[0:1])
		checkErr(err)
		if s != 1 {
			return '\x1b'
		}

		s, err = os.Stdin.Read(buf[1:2])
		checkErr(err)
		if s != 1 {
			return '\x1b'
		}

		if buf[0] == '[' {
			if buf[1] >= '0' && buf[1] <= '9' {
				s, err = os.Stdin.Read(buf[2:])
				checkErr(err)
				if s != 1 {
					return '\x1b'
				}
				if buf[2] == '~' {
					switch buf[1] {
					case '1':
						return HOME_KEY
					case '3':
						return DEL_KEY
					case '4':
						return END_KEY
					case '5':
						return PAGE_UP
					case '6':
						return PAGE_DOWN
					case '7':
						return HOME_KEY
					case '8':
						return END_KEY
					}
				}
			} else {
				switch buf[1] {
				case 'A':
					return ARROW_UP
				case 'B':
					return ARROW_DOWN
				case 'C':
					return ARROW_RIGHT
				case 'D':
					return ARROW_LEFT
				case 'H':
					return HOME_KEY
				case 'F':
					return END_KEY
				}
			}
		} else if buf[0] == 'O' {
			switch buf[1] {
			case 'H':
				return HOME_KEY
			case 'F':
				return END_KEY
			}
		} else {
			return '\x1b'
		}

	}
	return int(char)
}

func editorProcessKeys() {
	ch := int(editorReadKey())
	switch ch {
	case ctrlkey('q'):
		runOnExit()

	case HOME_KEY:
		conf.cx = 0

	case END_KEY:
		conf.cx = conf.cols - 1

	case PAGE_UP:
		reps := conf.cy
		for i := 0; i < reps; i++ {
			editorMoveCursor(ARROW_UP)
		}

	case PAGE_DOWN:
		reps := conf.rows - conf.cy
		for i := 0; i < reps; i++ {
			editorMoveCursor(ARROW_DOWN)
		}

	case ARROW_LEFT:
		fallthrough
	case ARROW_UP:
		fallthrough
	case ARROW_DOWN:
		fallthrough
	case ARROW_RIGHT:
		editorMoveCursor(ch)
	}
}

func main() {
	temp, err := term.MakeRaw(int(os.Stdin.Fd()))
	checkErr(err)

	conf.oldState = temp

	initEditor()
	if len(os.Args) >= 2 {
		editorOpen(os.Args[1])
	}
	for {
		editorRefreshScreen()
		editorProcessKeys()
	}
}
