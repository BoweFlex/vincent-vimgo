// Defines the input processor which is used for:
// - tracking the current mode
// - processing user input
// - updating the screen
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
	normal mode = iota
	insert
	visual
	command
)

type position struct {
	x, y int
}

func setCoordinates(x, y int) position {
	return position{x: x, y: y}
}

type cursorInfo struct {
	position     position
	preferredCol int
}

func (c *cursorInfo) clampX(minX, maxX int) {
	if c.position.x < minX {
		c.position.x = minX
	} else if c.position.x >= maxX {
		c.position.x = maxX - 1
	}
}

func (c *cursorInfo) clampY(minY, maxY int) {
	if c.position.y < minY {
		c.position.y = minY
	} else if c.position.y >= maxY {
		c.position.y = maxY - 1
	}
}

func (c *cursorInfo) getCoordinates() (int, int) {
	return c.position.x, c.position.y
}

type inputProcessor struct {
	screen                    tcell.Screen
	screenWidth, screenHeight int
	buffer                    string
	cursor                    cursorInfo
	currentMode               mode
	command                   []rune
}

func (p *inputProcessor) updateScreenSize() {
	p.screen.Sync()
	p.screenWidth, p.screenHeight = p.screen.Size()
}

func (p *inputProcessor) drawStatusLine() {
	if command == p.currentMode {
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

func (p *inputProcessor) showCursor() {
	p.cursor.clampX(0, p.screenWidth)
	p.cursor.clampY(0, p.screenHeight)
	p.screen.ShowCursor(p.cursor.getCoordinates())
}

func (p *inputProcessor) showBuffer() {
	p.screen.Clear()
	for i, line := range strings.Split(p.buffer, "\n") {
		if i >= p.screenHeight {
			break
		}
		if i == p.cursor.position.y {
			p.cursor.position.x = len(line)
		}
		for j, char := range line {
			if j < p.screenWidth {
				p.screen.SetContent(j, i, char, nil, tcell.StyleDefault)
			}
		}
	}
}

func (p *inputProcessor) updateScreen() {
	p.showCursor()
	p.showBuffer()
	p.drawStatusLine()
	p.screen.Show()
}

func (p *inputProcessor) process(input *tcell.EventKey) (err error) {
	// This should be removed at some point but it's a good failsafe if command mode is broken for some reason
	if input.Key() == tcell.KeyCtrlC {
		panic(errors.New("Ctrl+C Entered"))
	}
	switch p.currentMode {
	case command:
		err = p.processInputCommand(input)
	case insert:
		err = p.processInputInsert(input)
	case visual:
		err = p.processInputVisual(input)
	default:
		err = p.processInputNormal(input)
	}
	if err != nil {
		return err
	}
	p.updateScreen()
	return nil
}

func (p *inputProcessor) processInputNormal(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.currentMode = normal
	case tcell.KeyRune:
		break
	case tcell.KeyLeft:
		p.cursor.position.x--
	case tcell.KeyDown:
		p.cursor.position.y++
	case tcell.KeyUp:
		p.cursor.position.y--
	case tcell.KeyRight:
		p.cursor.position.x++
	}
	switch input.Rune() {
	case 'i':
		p.currentMode = insert
	case 'v':
		p.currentMode = visual
	case ':':
		p.currentMode = command
	case 'h':
		p.cursor.position.x--
	case 'j':
		p.cursor.position.y++
	case 'k':
		p.cursor.position.y--
	case 'l':
		p.cursor.position.x++
	}
	return nil
}

func (p *inputProcessor) backSpace() {
	if size := len(p.buffer); size > 0 {
		p.buffer = p.buffer[:size-1]
	}
}

func (p *inputProcessor) processInputInsert(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.currentMode = normal
	case tcell.KeyEnter:
		p.cursor.position.y++
		p.cursor.position.x = 0
		p.buffer += "\n"
	case tcell.KeyDEL:
		p.backSpace()
		p.cursor.position.x--
	case tcell.KeyBS:
		p.backSpace()
		p.cursor.position.x--
	case tcell.KeyRune:
		p.cursor.position.x++
		p.buffer += string(input.Rune())
	case tcell.KeyLeft:
		p.cursor.position.x--
	case tcell.KeyDown:
		p.cursor.position.y++
	case tcell.KeyUp:
		p.cursor.position.y--
	case tcell.KeyRight:
		p.cursor.position.x++
	}

	return nil
}

func (p *inputProcessor) processInputVisual(input *tcell.EventKey) (err error) {
	if input.Key() == tcell.KeyEscape {
		p.currentMode = normal
		return nil
	}
	return nil
}

func (p *inputProcessor) popCommandRune() error {
	if size := len(p.command); size > 0 {
		p.command = p.command[:size-1]
		return nil
	}
	return errors.New("nothing in command buffer")
}

func (p *inputProcessor) sendCommand() {
	if slices.Equal([]rune{'q'}, p.command) || slices.Equal([]rune("quit"), p.command) {
		p.screen.Fini()
		os.Exit(0)
	}
}

func (p *inputProcessor) processInputCommand(input *tcell.EventKey) (err error) {
	leaveCommandMode := func() error {
		p.currentMode = normal
		p.command = []rune{}
		return nil
	}
	switch input.Key() {
	case tcell.KeyEscape:
		return leaveCommandMode()
	case tcell.KeyBS:
		fallthrough
	case tcell.KeyDEL:
		if err := p.popCommandRune(); err != nil {
			return leaveCommandMode()
		}
	case tcell.KeyEnter:
		p.sendCommand()
	case tcell.KeyRune:
		p.command = append(p.command, input.Rune())
	}
	return nil
}
