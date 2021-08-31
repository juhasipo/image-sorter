package apitype

type Operation int

const (
	UNCATEGORIZE Operation = 0
	CATEGORIZE   Operation = 1
)

func (s Operation) NextOperation() Operation {
	return (s + 1) % 2
}

func (s Operation) AsId() int64 {
	return int64(s)
}

func OperationFromId(value int64) Operation {
	return Operation(value)
}
