package mysqldriver

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/Kavantix/lazysql/internal/database"
	"github.com/go-sql-driver/mysql"
)

var _ database.Driver = &mysqlDriver{}

type mysqlDriver struct {
	database.BaseDriver
	config *mysql.Config
}

// QueryForTable implements Driver.
func (m *mysqlDriver) QueryForTable(table database.Table, limit int) database.Query {
	return database.Query(fmt.Sprintf("SELECT *\nFROM `%s`\nLIMIT %d", table, limit))
}

func NewMysqlDriver(dsn database.Dsn) (database.Driver, error) {
	config := mysql.NewConfig()
	port := dsn.Port
	if port == 0 {
		port = 3306
	}
	config.Addr = dsn.Host + ":" + strconv.FormatInt(int64(dsn.Port), 10)
	config.User = dsn.User
	config.Passwd = dsn.Password
	config.TLSConfig = "skip-verify"

	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, err
	}

	driver := &mysqlDriver{
		config: config,
		BaseDriver: database.BaseDriver{
			Db: sql.OpenDB(connector),
		},
	}

	return driver, nil
}

func (m *mysqlDriver) Databases() ([]database.Database, error) {
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

func (m *mysqlDriver) SelectDatabase(db database.Database) error {
	if m.config.DBName == string(db) {
		return nil
	}
	if m.Db != nil {
		go m.Db.Close()
		m.config.DBName = string(db)
		connector, err := mysql.NewConnector(m.config)
		if err != nil {
			return err
		}
		m.Db = sql.OpenDB(connector)
	}
	return nil
}

func (m *mysqlDriver) Tables() ([]database.Table, error) {
	if m.config.DBName == "" {
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
