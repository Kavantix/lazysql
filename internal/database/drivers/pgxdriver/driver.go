package pgxdriver

import (
	"errors"
	"fmt"

	"github.com/Kavantix/lazysql/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

var _ database.Driver = &pgxDriver{}

type pgxDriver struct {
	database.BaseDriver
	config pgx.ConnConfig
}

// QueryForTable implements Driver.
func (m *pgxDriver) QueryForTable(table database.Table, limit int) database.Query {
	return database.Query(fmt.Sprintf("SELECT *\nFROM `%s`\nLIMIT %d", table, limit))
}

func NewPgxDriver(dsn database.Dsn) (database.Driver, error) {
	config := pgx.ConnConfig{}
	port := dsn.Port
	if port == 0 {
		port = 5432
	}
	config.Host = dsn.Host
	config.Port = dsn.Port
	config.User = dsn.User
	config.Password = dsn.Password
	// config.TLSConfig = nil

	driver := &pgxDriver{
		config: config,
		BaseDriver: database.BaseDriver{
			Db: stdlib.OpenDB(config),
		},
	}

	return driver, nil
}

func (m *pgxDriver) Databases() ([]database.Database, error) {
	databases := []database.Database{}
	rows, err := m.Db.Query("SHOW DATABASES")
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

func (m *pgxDriver) SelectDatabase(db database.Database) error {
	if m.config.Database == string(db) {
		return nil
	}
	if m.Db != nil {
		go m.Db.Close()
		m.config.Database = string(db)
		m.Db = stdlib.OpenDB(m.config)
	}
	return nil
}

func (m *pgxDriver) Tables() ([]database.Table, error) {
	if m.config.Database == "" {
		return nil, errors.New("no database selected")
	}
	tables := []database.Table{}
	rows, err := m.Db.Query("SHOW TABLES")
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
