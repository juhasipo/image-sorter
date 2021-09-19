package apitype

type Command interface {
	IsThrottled() bool
}

type Throttled struct {
}

type NotThrottled struct {
}

func (s *Throttled) IsThrottled() bool {
	return true
}

func (s *NotThrottled) IsThrottled() bool {
	return false
}
