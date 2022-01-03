package driver

import "unsafe"

type Dsn struct {
	Host, Port     string
	User, Password string
}

type QueryResult struct {
	Columns []string
	Data    [][]string
}

type Query string
type Database string
type Table string

type DatabaseDriver interface {
	Databases() ([]Database, error)
	SelectDatabase(db Database) error
	Tables() ([]Table, error)
	Query(query Query) (*QueryResult, error)
}

type base struct {
	currentQuery Query
}

func (b *base) CurrentQuery() Query {
	return b.currentQuery
}

func DatabaseNames(databases []Database) []string {
	return *(*[]string)(unsafe.Pointer(&databases))
}

func TableNames(tables []Table) []string {
	return *(*[]string)(unsafe.Pointer(&tables))
}
