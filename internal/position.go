package internal

type Position struct {
	X, Y int
}

func setCoordinates(x, y int) Position {
	return Position{X: x, Y: y}
}
