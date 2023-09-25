package main

import (
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type Config struct {
    rows int
    cols int
}

var config Config

func drawRows()  {
    for i:=0 ; i < config.rows; i++ {
        s := strconv.Itoa(i)
        os.Stdout.WriteString(s+"\r\n")
    }
}

func refreshScreen()  {
    os.Stdout.WriteString("\x1b[H")

    drawRows()

    os.Stdout.WriteString("\x1b[H")
}

func readKey(ev *tcell.EventKey)  {
    key := ev.Key()
    if key == tcell.KeyCtrlQ {
        os.Exit(0)
    }
}

func handleEvents(screen tcell.Screen)  {
    ev := screen.PollEvent()
    switch ev:= ev.(type) {


    case *tcell.EventKey:
        readKey(ev)
    }
}

func configUpdate(screen tcell.Screen)  {
    config.rows, config.cols = screen.Size()
}

func main() {
    // Create a new screen object
    screen, err := tcell.NewScreen()
    if err != nil {
        panic(err)
    }

    // Initialize the screen
    if err := screen.Init(); err != nil {
        panic(err)
    }
    defer screen.Fini()



    // Wait for events
    for {
        configUpdate(screen)
        screen.Clear()
        refreshScreen()
        handleEvents(screen)
    }
}
