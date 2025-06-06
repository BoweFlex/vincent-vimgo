package internal

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

func ListenForInput(s tcell.Screen, p *InputProcessor) {
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
			p.UpdateScreenSize()
		case *tcell.EventKey:
			err := p.Process(ev)
			if err != nil {
				panic(err)
			}
		}
	}
}
