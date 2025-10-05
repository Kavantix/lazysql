package pgxdriver

import (
	"errors"
	"fmt"
	"strings"

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
	parts := strings.SplitN(string(table), ".", 2)
	prefix := strings.Builder{}
	if parts[0] != "public" {
		prefix.WriteByte('"')
		prefix.WriteString(parts[0])
		prefix.WriteByte('"')
		prefix.WriteByte('.')
	}
	return database.Query(fmt.Sprintf("SELECT *\nFROM %s\"%s\"\nLIMIT %d", prefix.String(), parts[1], limit))
}

func NewPgxDriver(dsn database.Dsn) (database.Driver, error) {
	if dsn.Port == 0 {
		dsn.Port = 5432
	}
	url := fmt.Sprintf("postgres://%s:%s@%s:%d",
		dsn.User, dsn.Password, dsn.Host, dsn.Port)
	config, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	// config.TLSConfig = nil

	driver := &pgxDriver{
		config: *config,
		BaseDriver: database.BaseDriver{
			Db: stdlib.OpenDB(*config),
		},
	}

	return driver, nil
}

func (m *pgxDriver) Databases() ([]database.Database, error) {
	databases := []database.Database{}
	rows, err := m.Db.Query("select datname from pg_catalog.pg_database where not datistemplate order by datname")
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
	whereClause := "where schemaname not in ('pg_catalog', 'information_schema')"
	if m.config.Database == "postgres" {
		whereClause = ""
	}
	rows, err := m.Db.Query(fmt.Sprintf("SELECT schemaname || '.' || tablename FROM pg_catalog.pg_tables %s order by schemaname, tablename", whereClause))
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
