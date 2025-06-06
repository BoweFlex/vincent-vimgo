package internal

type Mode int

const (
	NORMAL Mode = iota
	INSERT
	VISUAL
	COMMAND
)
