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
var _ database.Table = pgxTable{}

type pgxDriver struct {
	database.BaseDriver
	config pgx.ConnConfig
}

type pgxTable struct {
	schema        string
	name          string
	longestSchema string
}

// EqualsTable implements database.Table.
func (t pgxTable) EqualsTable(other database.Table) bool {
	otherTable, ok := other.(pgxTable)
	return ok && t == otherTable
}

func (t pgxTable) DisplayString() string {
	builder := strings.Builder{}
	if t.schema != "public" {
		builder.WriteString(t.schema)
		for i := len(t.schema); i < len(t.longestSchema); i++ {
			builder.WriteByte(' ')
		}
		builder.WriteString(" | ")
	}
	builder.WriteString(t.name)
	return builder.String()
}

// QueryForTable implements Driver.
func (m *pgxDriver) QueryForTable(dbTable database.Table, limit int) database.Query {
	table := dbTable.(pgxTable)
	prefix := strings.Builder{}
	if table.schema != "public" {
		prefix.WriteByte('"')
		prefix.WriteString(table.schema)
		prefix.WriteByte('"')
		prefix.WriteByte('.')
	}
	return database.Query(fmt.Sprintf("SELECT *\nFROM %s\"%s\"\nLIMIT %d", prefix.String(), table.name, limit))
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
	var result []database.Table
	tables := []pgxTable{}
	whereClause := "where schemaname not in ('pg_catalog', 'information_schema')"
	if m.config.Database == "postgres" {
		whereClause = ""
	}
	rows, err := m.Db.Query(fmt.Sprintf("SELECT schemaname, tablename FROM pg_catalog.pg_tables %s order by schemaname = 'public' desc, schemaname, tablename", whereClause))
	if err != nil {
		return result, err
	}
	index := 0
	longestSchema := ""
	for rows.Next() {
		table := pgxTable{}
		err := rows.Scan(&table.schema, &table.name)
		if err != nil {
			return result, err
		}
		if len(table.schema) > len(longestSchema) {
			longestSchema = table.schema
		}
		tables = append(tables, table)
		index += 1
	}
	result = make([]database.Table, len(tables))
	for i, table := range tables {
		table.longestSchema = longestSchema
		result[i] = table
	}
	return result, nil
}
