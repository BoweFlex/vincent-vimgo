package internal

type CursorInfo struct {
	Position     Position
	PreferredCol int
}

func (p *Position) ClampX(minX, maxX int) {
	if p.X < minX {
		p.X = minX
	} else if p.X >= maxX {
		p.X = maxX - 1
	}
}

func (p *Position) ClampY(minY, maxY int) {
	if p.Y < minY {
		p.Y = minY
	} else if p.Y >= maxY {
		p.Y = maxY - 1
	}
}

func (c *CursorInfo) GetCoordinates() (int, int) {
	return c.Position.X, c.Position.Y
}

func (c *CursorInfo) AddDelta(xDelta, yDelta int, changePreferredCol bool) {
	c.Position.X += xDelta
	c.Position.Y += yDelta
	if changePreferredCol && xDelta != 0 {
		c.PreferredCol = c.Position.X
	}
}
