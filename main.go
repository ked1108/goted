package main

import (
    "github.com/gdamore/tcell/v2"
    "os"
)

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
        ev := screen.PollEvent()
        switch ev := ev.(type) {
        case *tcell.EventKey:
            key := ev.Key()
			if key == tcell.KeyCtrlC {
				os.Exit(0)
			}
        }
    }
}
