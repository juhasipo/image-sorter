package common

type Operation int

const (
	NONE Operation = 0
	MOVE Operation = 1
)

func (s Operation) NextOperation() Operation {
	return (s + 1) % 2
}
