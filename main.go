package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	s.SetStyle(defStyle)

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()
	proc := inputProcessor{
		screen: s,
		offset: setCoordinates(0, 0),
		cursor: cursorInfo{
			position:     setCoordinates(0, 0),
			preferredCol: 0,
		},
		currentMode: normal,
		command:     []rune{},
	}
	// For some reason even if proc.updateScreen() was called here, the starting
	// screen was not drawn. Using PostEvent to send an immediate, and invisible,
	// `ESC` key to the editor so screen will be drawn.
	if err := s.PostEvent(tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone)); err != nil {
		err = fmt.Errorf("%w, why was queue full immediately", err)
		panic(err)
	}

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			proc.updateScreenSize()
		case *tcell.EventKey:
			err := proc.process(ev)
			if err != nil {
				panic(err)
			}
		}
	}
}
