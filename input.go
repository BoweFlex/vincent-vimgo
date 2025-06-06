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

func (p *position) clampX(minX, maxX int) {
	if p.x < minX {
		p.x = minX
	} else if p.x >= maxX {
		p.x = maxX - 1
	}
}

func (p *position) clampY(minY, maxY int) {
	if p.y < minY {
		p.y = minY
	} else if p.y >= maxY {
		p.y = maxY - 1
	}
}

func (c *cursorInfo) getCoordinates() (int, int) {
	return c.position.x, c.position.y
}

func (c *cursorInfo) addDelta(xDelta, yDelta int, changePreferredCol bool) {
	c.position.x += xDelta
	c.position.y += yDelta
	if changePreferredCol && xDelta != 0 {
		c.preferredCol = c.position.x
	}
}

type inputProcessor struct {
	screen                    tcell.Screen
	offset                    position
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

func (p *inputProcessor) setOffset(lines []string, usableHeight int) {
	p.offset.clampY(0, len(lines))
	numLinesShown := len(lines) - p.offset.y

	if p.cursor.position.y >= usableHeight {
		diff := max(p.cursor.position.y-usableHeight, 1)
		p.offset.y += diff
		p.cursor.addDelta(0, -diff, false)
	} else if p.cursor.position.y < p.offset.y {
		diff := p.offset.y - max(p.cursor.position.y, 0)
		p.offset.y -= diff
	}
	if numLinesShown > usableHeight {
		p.cursor.position.clampY(p.offset.y, usableHeight)
	} else {
		p.cursor.position.clampY(p.offset.y, numLinesShown)
	}
	if p.cursor.position.x >= p.screenWidth {
		diff := max(p.cursor.position.x-p.screenWidth, 1)
		p.offset.x += diff
	} else if p.cursor.position.x < p.offset.x {
		diff := p.offset.x - max(p.cursor.position.x, 0)
		p.offset.x -= diff
	}
	currLineVisibleLength := len(lines[p.cursor.position.y]) - p.offset.x + 1 // Cursor should be kept within the current bounds of the line, but allowed one in front of the text
	p.cursor.position.x = p.cursor.preferredCol
	p.cursor.position.clampX(p.offset.x, currLineVisibleLength)
}

func (p *inputProcessor) showBuffer() {
	p.screen.Clear()
	usableHeight := p.screenHeight - 2 // Accounts for the space reserved for statusline
	lines := strings.Split(p.buffer, "\n")
	p.setOffset(lines, usableHeight)

	for i := p.offset.y; i < usableHeight && i < len(lines); i++ {
		line := lines[i]
		chars := []rune(line)
		for j := p.offset.x; j < len(chars); j++ {
			char := chars[j]
			p.screen.SetContent(j-p.offset.x, i-p.offset.y, char, nil, tcell.StyleDefault)
		}
	}
	p.screen.ShowCursor(p.cursor.getCoordinates())
}

func (p *inputProcessor) updateScreen() {
	// var wg sync.WaitGroup
	drawFuncs := []func(){
		p.showBuffer,
		p.drawStatusLine,
	}

	// wg.Add(len(drawFuncs))
	for _, drawFunc := range drawFuncs {
		drawFunc()
		// go func(draw func()) {
		// 	defer wg.Done()
		// 	draw()
		// }(drawFunc)
	}

	// wg.Wait()
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
		p.cursor.addDelta(-1, 0, true)
	case tcell.KeyDown:
		p.cursor.addDelta(0, 1, false)
	case tcell.KeyUp:
		p.cursor.addDelta(0, -1, false)
	case tcell.KeyRight:
		p.cursor.addDelta(1, 0, true)
	}
	switch input.Rune() {
	case 'i':
		p.currentMode = insert
	case 'v':
		p.currentMode = visual
	case ':':
		p.currentMode = command
	case 'h':
		p.cursor.addDelta(-1, 0, true)
	case 'j':
		p.cursor.addDelta(0, 1, false)
	case 'k':
		p.cursor.addDelta(0, -1, false)
	case 'l':
		p.cursor.addDelta(1, 0, true)
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
		p.cursor.preferredCol = p.cursor.position.x
	case tcell.KeyBS:
		p.backSpace()
		p.cursor.position.x--
		p.cursor.preferredCol = p.cursor.position.x
	case tcell.KeyRune:
		p.cursor.position.x++
		p.cursor.preferredCol = p.cursor.position.x
		p.buffer += string(input.Rune())
	case tcell.KeyLeft:
		p.cursor.addDelta(-1, 0, true)
	case tcell.KeyDown:
		p.cursor.addDelta(0, 1, false)
	case tcell.KeyUp:
		p.cursor.addDelta(0, -1, false)
	case tcell.KeyRight:
		p.cursor.addDelta(1, 0, true)
	}

	return nil
}

func (p *inputProcessor) processInputVisual(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.currentMode = normal
	case tcell.KeyRune:
		break
	case tcell.KeyLeft:
		p.cursor.addDelta(-1, 0, true)
	case tcell.KeyDown:
		p.cursor.addDelta(0, 1, false)
	case tcell.KeyUp:
		p.cursor.addDelta(0, -1, false)
	case tcell.KeyRight:
		p.cursor.addDelta(1, 0, true)
	}
	switch input.Rune() {
	case 'i':
		p.currentMode = insert
	case 'v':
		p.currentMode = normal
	case ':':
		p.currentMode = command
	case 'h':
		p.cursor.addDelta(-1, 0, true)
	case 'j':
		p.cursor.addDelta(0, 1, false)
	case 'k':
		p.cursor.addDelta(0, -1, false)
	case 'l':
		p.cursor.addDelta(1, 0, true)
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
