package internal

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type InputProcessor struct {
	Screen                    tcell.Screen
	Offset                    Position
	ScreenWidth, ScreenHeight int
	Buffer                    string
	Cursor                    CursorInfo
	CurrentMode               Mode
	Command                   []rune
}

func NewInputProcessor(screen tcell.Screen) *InputProcessor {
	return &InputProcessor{
		Screen: screen,
		Offset: setCoordinates(0, 0),
		Cursor: CursorInfo{
			Position:     setCoordinates(0, 0),
			PreferredCol: 0,
		},
		CurrentMode: NORMAL,
		Command:     []rune{},
	}
}

func (p *InputProcessor) UpdateScreenSize() {
	p.Screen.Sync()
	p.ScreenWidth, p.ScreenHeight = p.Screen.Size()
}

func (p *InputProcessor) drawStatusLine() {
	if COMMAND == p.CurrentMode {
		displayStr := []rune{':'}
		displayStr = append(displayStr, p.Command...)
		for i := range p.ScreenWidth {
			var c rune
			if i < len(displayStr) {
				c = displayStr[i]
			} else {
				c = ' '
			}
			p.Screen.SetContent(i, p.ScreenHeight-1, c, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite.TrueColor()))
		}
		return
	}

	cursorLocationString := fmt.Sprintf("%d:%d", p.Cursor.Position.Y, p.Cursor.Position.X)
	statusLineStyle := tcell.StyleDefault.Background(tcell.ColorBlack.TrueColor()).Foreground(tcell.ColorWhite.TrueColor())

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
	modeStyle, modeString := modeStyles[p.CurrentMode], modeStrings[p.CurrentMode]
	for i, char := range modeString {
		p.Screen.SetContent(i, p.ScreenHeight-1, char, nil, modeStyle)
	}
	for i := len(modeString); i < p.ScreenWidth-len(cursorLocationString); i++ {
		p.Screen.SetContent(i, p.ScreenHeight-1, ' ', nil, statusLineStyle)
	}
	startCursorLocationColumn := p.ScreenWidth - len(cursorLocationString)
	for i := startCursorLocationColumn; i < p.ScreenWidth; i++ {
		char := rune(cursorLocationString[i-startCursorLocationColumn])
		p.Screen.SetContent(i, p.ScreenHeight-1, char, nil, statusLineStyle)
	}
}

func (p *InputProcessor) setOffset(lines []string, usableHeight int) {
	p.Offset.ClampY(0, len(lines))
	numLinesShown := len(lines) - p.Offset.Y

	if p.Cursor.Position.Y >= usableHeight {
		diff := max(p.Cursor.Position.Y-usableHeight, 1)
		p.Offset.Y += diff
		p.Cursor.AddDelta(0, -diff, false)
	} else if p.Cursor.Position.Y < p.Offset.Y {
		diff := p.Offset.Y - max(p.Cursor.Position.Y, 0)
		p.Offset.Y -= diff
	}
	if numLinesShown > usableHeight {
		p.Cursor.Position.ClampY(p.Offset.Y, usableHeight)
	} else {
		p.Cursor.Position.ClampY(p.Offset.Y, numLinesShown)
	}
	if p.Cursor.Position.X >= p.ScreenWidth {
		diff := max(p.Cursor.Position.X-p.ScreenWidth, 1)
		p.Offset.X += diff
	} else if p.Cursor.Position.X < p.Offset.X {
		diff := p.Offset.X - max(p.Cursor.Position.X, 0)
		p.Offset.X -= diff
	}
	currLineVisibleLength := len(lines[p.Cursor.Position.Y]) - p.Offset.X + 1 // Cursor should be kept within the current bounds of the line, but allowed one in front of the text
	p.Cursor.Position.X = p.Cursor.PreferredCol
	p.Cursor.Position.ClampX(p.Offset.X, currLineVisibleLength)
}

func (p *InputProcessor) showBuffer() {
	p.Screen.Clear()
	usableHeight := p.ScreenHeight - 2 // Accounts for the space reserved for statusline
	lines := strings.Split(p.Buffer, "\n")
	p.setOffset(lines, usableHeight)

	for i := p.Offset.Y; i < usableHeight && i < len(lines); i++ {
		line := lines[i]
		chars := []rune(line)
		for j := p.Offset.X; j < len(chars); j++ {
			char := chars[j]
			p.Screen.SetContent(j-p.Offset.X, i-p.Offset.Y, char, nil, tcell.StyleDefault)
		}
	}
	p.Screen.ShowCursor(p.Cursor.GetCoordinates())
}

func (p *InputProcessor) updateScreen() {
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
	p.Screen.Show()
}

func (p *InputProcessor) Process(input *tcell.EventKey) (err error) {
	// This should be removed at some point but it's a good failsafe if command mode is broken for some reason
	if input.Key() == tcell.KeyCtrlC {
		panic(errors.New("Ctrl+C Entered"))
	}
	switch p.CurrentMode {
	case COMMAND:
		err = p.processInputCommand(input)
	case INSERT:
		err = p.processInputInsert(input)
	case VISUAL:
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

func (p *InputProcessor) processInputNormal(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.CurrentMode = NORMAL
	case tcell.KeyRune:
		break
	case tcell.KeyLeft:
		p.Cursor.AddDelta(-1, 0, true)
	case tcell.KeyDown:
		p.Cursor.AddDelta(0, 1, false)
	case tcell.KeyUp:
		p.Cursor.AddDelta(0, -1, false)
	case tcell.KeyRight:
		p.Cursor.AddDelta(1, 0, true)
	}
	switch input.Rune() {
	case 'i':
		p.CurrentMode = INSERT
	case 'v':
		p.CurrentMode = VISUAL
	case ':':
		p.CurrentMode = COMMAND
	case 'h':
		p.Cursor.AddDelta(-1, 0, true)
	case 'j':
		p.Cursor.AddDelta(0, 1, false)
	case 'k':
		p.Cursor.AddDelta(0, -1, false)
	case 'l':
		p.Cursor.AddDelta(1, 0, true)
	}
	return nil
}

func (p *InputProcessor) backSpace() {
	if size := len(p.Buffer); size > 0 {
		p.Buffer = p.Buffer[:size-1]
	}
}

func (p *InputProcessor) processInputInsert(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.CurrentMode = NORMAL
	case tcell.KeyEnter:
		p.Cursor.Position.Y++
		p.Cursor.Position.X = 0
		p.Buffer += "\n"
	case tcell.KeyDEL:
		p.backSpace()
		p.Cursor.Position.X--
		p.Cursor.PreferredCol = p.Cursor.Position.X
	case tcell.KeyBS:
		p.backSpace()
		p.Cursor.Position.X--
		p.Cursor.PreferredCol = p.Cursor.Position.X
	case tcell.KeyRune:
		p.Cursor.Position.X++
		p.Cursor.PreferredCol = p.Cursor.Position.X
		p.Buffer += string(input.Rune())
	case tcell.KeyLeft:
		p.Cursor.AddDelta(-1, 0, true)
	case tcell.KeyDown:
		p.Cursor.AddDelta(0, 1, false)
	case tcell.KeyUp:
		p.Cursor.AddDelta(0, -1, false)
	case tcell.KeyRight:
		p.Cursor.AddDelta(1, 0, true)
	}

	return nil
}

func (p *InputProcessor) processInputVisual(input *tcell.EventKey) (err error) {
	switch input.Key() {
	case tcell.KeyEscape:
		p.CurrentMode = NORMAL
	case tcell.KeyRune:
		break
	case tcell.KeyLeft:
		p.Cursor.AddDelta(-1, 0, true)
	case tcell.KeyDown:
		p.Cursor.AddDelta(0, 1, false)
	case tcell.KeyUp:
		p.Cursor.AddDelta(0, -1, false)
	case tcell.KeyRight:
		p.Cursor.AddDelta(1, 0, true)
	}
	switch input.Rune() {
	case 'i':
		p.CurrentMode = INSERT
	case 'v':
		p.CurrentMode = NORMAL
	case ':':
		p.CurrentMode = COMMAND
	case 'h':
		p.Cursor.AddDelta(-1, 0, true)
	case 'j':
		p.Cursor.AddDelta(0, 1, false)
	case 'k':
		p.Cursor.AddDelta(0, -1, false)
	case 'l':
		p.Cursor.AddDelta(1, 0, true)
	}
	return nil
}

func (p *InputProcessor) popCommandRune() error {
	if size := len(p.Command); size > 0 {
		p.Command = p.Command[:size-1]
		return nil
	}
	return errors.New("nothing in command buffer")
}

func (p *InputProcessor) sendCommand() {
	if slices.Equal([]rune{'q'}, p.Command) || slices.Equal([]rune("quit"), p.Command) {
		p.Screen.Fini()
		os.Exit(0)
	}
}

func (p *InputProcessor) processInputCommand(input *tcell.EventKey) (err error) {
	leaveCommandMode := func() error {
		p.CurrentMode = NORMAL
		p.Command = []rune{}
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
		p.Command = append(p.Command, input.Rune())
	}
	return nil
}
