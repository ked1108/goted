package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
)

type erow struct {
	size   int
	rsize  int
	buf    string
	render string
}

type abuf struct {
	b string
	l int
}

type config struct {
	cx       int
	cy       int
	rx       int
	rowoff   int
	coloff   int
	rows     int
	cols     int
	numrows  int
	row      []erow
	filename string
	oldState *term.State
}

var conf config

//goland:noinspection GoSnakeCaseUsage
var ABUF_INIT = abuf{"", 0}

const TAB_WIDTH = 8

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
		_ = term.Restore(int(os.Stdin.Fd()), conf.oldState)
		fmt.Println(err)
		os.Exit(1)
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
	conf.filename = ""
	getSize()

	conf.rows -= 1
}

func editorAppendRow(line string, lineLen int) {
	conf.row = append(conf.row, erow{lineLen, 0, line, ""})
	conf.row[conf.numrows].rsize = 0
	conf.row[conf.numrows].render = ""
	editorUpdateRows(&conf.row[conf.numrows])
	conf.numrows++
}

func editorOpen(filename string) {
	conf.filename = filename
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

func editorRowCxtoRx(row *erow, cx int) int {
	rx := 0
	for i := 0; i < cx; i++ {
		if row.buf[i] == '\t' {
			rx += (TAB_WIDTH - 1) - (rx % TAB_WIDTH)
		}
		rx++
	}

	return rx
}

func editorDrawStatusBar(ab *abuf) {
	abAppend(ab, "\x1b[7m", 4)
	status := ""
	if conf.filename == "" {
		status = fmt.Sprintf("%.20s - %d lines", "[No Name]", conf.numrows)
	} else {
		status = fmt.Sprintf("%.20s - %d lines", conf.filename, conf.numrows)
	}
	length := len(status)
	abAppend(ab, status, len(status))
	for length < conf.cols {
		abAppend(ab, " ", 1)
		length++
	}
	abAppend(ab, "\x1b[m", 3)
}

func editorUpdateRows(row *erow) {
	tabs := 0
	for char := range row.buf {
		if char == '\t' {
			tabs++
		}
	}
	var buf []byte
	idx := 0
	for _, char := range row.buf {
		if char == '\t' {
			buf = append(buf, ' ')
			idx++
			for idx%TAB_WIDTH != 0 {
				buf = append(buf, ' ')
				idx++
			}
		} else {
			buf = append(buf, byte(char))
			idx++
		}
	}
	row.render = string(buf)
	row.rsize = len(row.render)
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
			abAppend(ab, "\r\n", 2)
		} else {
			l := conf.row[filerow].size - conf.coloff
			if l < 0 {
				l = 0
			}
			if l > conf.cols {
				l = conf.cols
			}
			abAppend(ab, "\x1b[K", 3)
			abAppend(ab, conf.row[filerow].buf[conf.coloff:], l)
			abAppend(ab, "\r\n", 2)
		}

	}
}

func editorMoveCursor(key int) {
	var row *erow
	if conf.cy >= conf.numrows {
		row = nil
	} else {
		row = &conf.row[conf.cy]
	}
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
		if row != nil && conf.cx < row.size {
			conf.cx++
		}
	}
	if conf.cy > conf.numrows {
		row = nil
	} else {
		row = &conf.row[conf.cy]
	}
	var rowlen int
	if row != nil {
		rowlen = row.size
	} else {
		rowlen = 0
	}
	if conf.cx > rowlen {
		conf.cx = rowlen
	}
}

func editorScroll() {
	conf.rx = conf.cx

	if conf.cy < conf.numrows {
		conf.rx = editorRowCxtoRx(&conf.row[conf.cy], conf.cx)
	}

	if conf.cy < conf.rowoff {
		conf.rowoff = conf.cy
	}
	if conf.cy >= conf.rowoff+conf.rows {
		conf.rowoff = conf.cy - conf.rows + 1
	}
	if conf.rx < conf.coloff {
		conf.coloff = conf.rx
	}
	if conf.rx >= conf.coloff+conf.cols {
		conf.coloff = conf.rx - conf.cols + 1
	}
}

func editorRefreshScreen() {
	editorScroll()
	ab := ABUF_INIT
	abAppend(&ab, "\x1b[?25l", 6)
	abAppend(&ab, "\x1b[H", 4)

	editorDrawRows(&ab)
	editorDrawStatusBar(&ab)

	buf := fmt.Sprintf("\x1b[%d;%dH", (conf.cy-conf.rowoff)+1, (conf.rx-conf.coloff)+1)
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
	ch := editorReadKey()
	switch ch {
	case ctrlkey('q'):
		runOnExit()

	case HOME_KEY:
		conf.cx = 0

	case END_KEY:
		if conf.cy < conf.numrows {
			conf.cx = conf.row[conf.cy].size
		}

	case PAGE_UP:
		fallthrough
	case PAGE_DOWN:
		if ch == PAGE_UP {
			conf.cy = conf.rowoff
		} else if ch == PAGE_DOWN {
			conf.cy = conf.rowoff + conf.rows - 1
			if conf.cy > conf.numrows {
				conf.cy = conf.numrows
			}
		}

		times := conf.rows
		for i := 0; i < times; i++ {
			if ch == PAGE_UP {
				editorMoveCursor(ARROW_UP)
			} else {
				editorMoveCursor(ARROW_DOWN)
			}
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
