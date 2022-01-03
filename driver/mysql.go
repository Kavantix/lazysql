package driver

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"

	"github.com/go-sql-driver/mysql"
)

type mysqlDriver struct {
	base
	config     *mysql.Config
	db         *sql.DB
	cancelFunc context.CancelFunc
	mutex      sync.Mutex
}

func NewMysqlDriver(dsn Dsn) (DatabaseDriver, error) {
	config := mysql.NewConfig()
	port := dsn.Port
	if port == "" {
		port = "3306"
	}
	config.Addr = dsn.Host + ":" + port
	config.User = dsn.User
	config.Passwd = dsn.Password

	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, err
	}

	driver := &mysqlDriver{
		config: config,
		db:     sql.OpenDB(connector),
	}

	return driver, nil
}

func (m *mysqlDriver) Databases() ([]Database, error) {
	databases := []Database{}
	rows, err := m.db.Query("Show databases")
	if err != nil {
		return databases, err
	}
	// checkErr(err)
	index := 0
	for rows.Next() {
		databases = append(databases, "")
		err := rows.Scan(&databases[index])
		if err != nil {
			return databases, err
		}
		index += 1
	}
	return databases, nil
}

func (m *mysqlDriver) SelectDatabase(db Database) error {
	if m.config.DBName == string(db) {
		return nil
	}
	if m.db != nil {
		go m.db.Close()
		m.config.DBName = string(db)
		connector, err := mysql.NewConnector(m.config)
		if err != nil {
			return err
		}
		m.db = sql.OpenDB(connector)
	}
	return nil
}

func (m *mysqlDriver) Tables() ([]Table, error) {
	if m.config.DBName == "" {
		return nil, errors.New("no database selected")
	}
	tables := []Table{}
	rows, err := m.db.Query("Show tables")
	if err != nil {
		return tables, err
	}
	index := 0
	for rows.Next() {
		tables = append(tables, "")
		err := rows.Scan(&tables[index])
		if err != nil {
			return tables, err
		}
		index += 1
	}
	return tables, nil
}

func (m *mysqlDriver) Query(query Query) (*QueryResult, error) {
	m.mutex.Lock()
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	context, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel
	m.currentQuery = query
	m.mutex.Unlock()

	data := [][]string{}
	rows, err := m.db.QueryContext(context, string(query))
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
