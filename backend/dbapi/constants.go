package dbapi

type TableExist bool

const (
	TableNotExist TableExist = false
	TableExists   TableExist = true
)
