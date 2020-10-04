package apitype

type PersistCategorizationCommand struct {
	keepOriginals  bool
	fixOrientation bool
	quality        int
}

func NewPersistCategorizationCommand(keepOriginals bool, fixOrientation bool, quality int) PersistCategorizationCommand {
	return PersistCategorizationCommand{
		keepOriginals:  keepOriginals,
		fixOrientation: fixOrientation,
		quality:        quality,
	}
}

func (s *PersistCategorizationCommand) ShouldKeepOriginals() bool {
	return s.keepOriginals
}
func (s *PersistCategorizationCommand) ShouldFixOrientation() bool {
	return s.fixOrientation
}

func (s *PersistCategorizationCommand) GetQuality() int {
	return s.quality
}
