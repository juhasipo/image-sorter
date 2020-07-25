package common

type Operation int

const (
	NONE Operation = 0
	MOVE Operation = 1
)

func (s Operation) NextOperation() Operation {
	return (s + 1) % 2
}

type PersistCategorizationCommand struct {
	keepOriginals  bool
	fixOrientation bool
}

func PersistCategorizationCommandNew(keepOriginals bool, fixOrientation bool) PersistCategorizationCommand {
	return PersistCategorizationCommand{
		keepOriginals:  keepOriginals,
		fixOrientation: fixOrientation,
	}
}

func (s *PersistCategorizationCommand) ShouldKeepOriginals() bool {
	return s.keepOriginals
}
func (s *PersistCategorizationCommand) ShouldFixOrientation() bool {
	return s.fixOrientation
}
