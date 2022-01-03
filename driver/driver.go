package driver

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"unsafe"
)

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
	CancelQuery() bool
}

type BaseDriver struct {
	currentQuery Query
	Db           *sql.DB
	context      context.Context
	cancelFunc   context.CancelFunc
	queryMutex   sync.Mutex
}

func (b *BaseDriver) CurrentQuery() Query {
	return b.currentQuery
}

func DatabaseNames(databases []Database) []string {
	return *(*[]string)(unsafe.Pointer(&databases))
}

func TableNames(tables []Table) []string {
	return *(*[]string)(unsafe.Pointer(&tables))
}

func (b *BaseDriver) CancelQuery() bool {
	b.queryMutex.Lock()
	defer b.queryMutex.Unlock()
	if b.cancelFunc != nil {
		b.cancelFunc()
		return true
	} else {
		return false
	}
}

func (b *BaseDriver) Query(query Query) (*QueryResult, error) {
	b.queryMutex.Lock()
	if b.cancelFunc != nil {
		b.cancelFunc()
	}
	context, cancel := context.WithCancel(context.Background())
	b.context = context
	b.cancelFunc = cancel
	b.currentQuery = query
	b.queryMutex.Unlock()
	defer func() {
		b.queryMutex.Lock()
		defer b.queryMutex.Unlock()
		if b.context == context {
			b.context = nil
			b.cancelFunc = nil
		}
	}()

	data := [][]string{}
	rows, err := b.Db.QueryContext(context, string(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	index := 0
	columns, err := rows.Columns()
	numColumns := len(columns)
	if err != nil {
		return nil, err
	}
	for rows.Next() && index < 9999 {
		if context.Err() != nil {
			return nil, context.Err()
		}
		row := make([]sql.NullString, numColumns)
		scannableRow := make([]interface{}, numColumns)
		for i := range row {
			scannableRow[i] = &row[i]
		}
		err := rows.Scan(scannableRow...)
		rowValues := make([]string, numColumns)
		for i, column := range row {
			if column.Valid {
				rowValues[i] = strings.ReplaceAll(column.String, "\r", "")
			} else {
				rowValues[i] = "NULL"
			}
		}
		data = append(data, rowValues)
		if err != nil {
			return nil, err
		}
		index += 1
	}
	if context.Err() != nil {
		return nil, context.Err()
	}
	return &QueryResult{
		Columns: columns,
		Data:    data,
	}, nil
}
