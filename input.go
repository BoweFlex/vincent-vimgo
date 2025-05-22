package main

import (
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type mode int

const (
	NORMAL mode = iota
	INSERT
	VISUAL
	COMMAND
)

type Position struct {
	x, y int
}

func SetCoordinates(x, y int) Position {
	return Position{x: x, y: y}
}

type CursorInfo struct {
	Position     Position
	PreferredCol int
}

func (c *CursorInfo) clampX(min, max int) {
	if c.Position.x < min {
		c.Position.x = min
	} else if c.Position.x >= max {
		c.Position.x = max - 1
	}
}

func (c *CursorInfo) clampY(min, max int) {
	if c.Position.y < min {
		c.Position.y = min
	} else if c.Position.y >= max {
		c.Position.y = max - 1
	}
}

func (c *CursorInfo) getCoordinates() (int, int) {
	return c.Position.x, c.Position.y
}

type InputProcessor struct {
	screen                    tcell.Screen
	screenWidth, screenHeight int
	buffer                    string
	cursor                    CursorInfo
	currentMode               mode
	command                   []rune
}

func (p *InputProcessor) updateScreenSize() {
	p.screen.Sync()
	p.screenWidth, p.screenHeight = p.screen.Size()
}

func (p *InputProcessor) drawStatusLine() {
	if COMMAND == p.currentMode {
		displayStr := []rune{':'}
		displayStr = append(displayStr, p.command...)
		for i := range p.screenWidth {
			var c rune
			if i < len(displayStr) {
				c = displayStr[i]
			} else {
				c = ' '
			}
			p.screen.SetContent(i, p.screenHeight-1, c, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite.TrueColor()))
		}
		return
	}

	modeStyles := []tcell.Style{
		tcell.StyleDefault.Foreground(tcell.ColorBlack.TrueColor()).Background(tcell.ColorBlue.TrueColor()),
		tcell.StyleDefault.Foreground(tcell.ColorBlack.TrueColor()).Background(tcell.ColorGreen.TrueColor()),
		tcell.StyleDefault.Foreground(tcell.ColorBlack.TrueColor()).Background(tcell.ColorPurple.TrueColor()),
	}
	modeStrings := []string{
		" NOR ",
		" INS ",
		" VIS ",
	}
	modeStyle, modeString := modeStyles[p.currentMode], modeStrings[p.currentMode]
	for i, char := range modeString {
		p.screen.SetContent(i, p.screenHeight-1, char, nil, modeStyle)
	}
	for i := len(modeString); i < p.screenWidth; i++ {
		p.screen.SetContent(i, p.screenHeight-1, ' ', nil, tcell.StyleDefault.Background(tcell.ColorBlack.TrueColor()))
	}
}

func (p *InputProcessor) showCursor() {
	p.cursor.clampX(0, p.screenWidth)
	p.cursor.clampY(0, p.screenHeight)
	p.screen.ShowCursor(p.cursor.getCoordinates())
}

func (p *InputProcessor) showBuffer() {
	p.screen.Clear()
	for i, line := range strings.Split(p.buffer, "\n") {
		if i >= p.screenHeight {
			break
		}
		if i == p.cursor.Position.y {
			p.cursor.Position.x = len(line)
		}
		for j, char := range line {
			if j < p.screenWidth {
				p.screen.SetContent(j, i, char, nil, tcell.StyleDefault)
			}
		}
	}
}

func (p *InputProcessor) updateScreen() {
	p.showCursor()
	p.showBuffer()
	p.drawStatusLine()
	p.screen.Show()
}

func (p *InputProcessor) Process(input *tcell.EventKey) (err error) {
	if input.Key() == tcell.KeyCtrlC {
		panic(errors.New("Ctrl+C Entered"))
	}
	if COMMAND == p.currentMode {
		err = p.processInputCommand(input)
	} else if INSERT == p.currentMode {
		err = p.processInputInsert(input)
	} else if VISUAL == p.currentMode {
		err = p.processInputVisual(input)
	} else {
		err = p.processInputNormal(input)
	}
	if err != nil {
		return err
	}
	p.updateScreen()
	return nil
}

func (p *InputProcessor) processInputNormal(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.currentMode = NORMAL
	case tcell.KeyRune:
		break
	case tcell.KeyLeft:
		p.cursor.Position.x -= 1
	case tcell.KeyDown:
		p.cursor.Position.y += 1
	case tcell.KeyUp:
		p.cursor.Position.y -= 1
	case tcell.KeyRight:
		p.cursor.Position.x += 1
	}
	switch input.Rune() {
	case 'i':
		p.currentMode = INSERT
	case 'v':
		p.currentMode = VISUAL
	case ':':
		p.currentMode = COMMAND
	case 'h':
		p.cursor.Position.x -= 1
	case 'j':
		p.cursor.Position.y += 1
	case 'k':
		p.cursor.Position.y -= 1
	case 'l':
		p.cursor.Position.x += 1
	}
	return nil
}

func (p *InputProcessor) backSpace() {
	if size := len(p.buffer); size > 0 {
		p.buffer = p.buffer[:size-1]
	}
}

func (p *InputProcessor) processInputInsert(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.currentMode = NORMAL
	case tcell.KeyEnter:
		p.cursor.Position.y += 1
		p.cursor.Position.x = 0
		p.buffer += "\n"
	case tcell.KeyDEL:
		p.backSpace()
		p.cursor.Position.x -= 1
	case tcell.KeyBS:
		p.backSpace()
		p.cursor.Position.x -= 1
	case tcell.KeyRune:
		p.cursor.Position.x += 1
		p.buffer += string(input.Rune())
	case tcell.KeyLeft:
		p.cursor.Position.x -= 1
	case tcell.KeyDown:
		p.cursor.Position.y += 1
	case tcell.KeyUp:
		p.cursor.Position.y -= 1
	case tcell.KeyRight:
		p.cursor.Position.x += 1
	}

	return nil
}

func (p *InputProcessor) processInputVisual(input *tcell.EventKey) (err error) {
	if input.Key() == tcell.KeyEscape {
		p.currentMode = NORMAL
		return nil
	}
	return nil
}

func (p *InputProcessor) popCommandRune() error {
	if size := len(p.command); size > 0 {
		p.command = p.command[:size-1]
		return nil
	}
	return errors.New("nothing in command buffer")
}

func (p *InputProcessor) sendCommand() {
	if slices.Equal([]rune{'q'}, p.command) || slices.Equal([]rune("quit"), p.command) {
		p.screen.Fini()
		os.Exit(0)
	}
}

func (p *InputProcessor) processInputCommand(input *tcell.EventKey) (err error) {
	leaveCommandMode := func() error {
		p.currentMode = NORMAL
		p.command = []rune{}
		return nil
	}
	if input.Key() == tcell.KeyEscape {
		return leaveCommandMode()
	} else if input.Key() == tcell.KeyBS || input.Key() == tcell.KeyDEL {
		if err := p.popCommandRune(); err != nil {
			return leaveCommandMode()
		}
	} else if input.Key() == tcell.KeyEnter {
		p.sendCommand()
	} else if input.Key() == tcell.KeyRune {
		p.command = append(p.command, input.Rune())
	}
	return nil
}
