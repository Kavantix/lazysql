package driver

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type mysqlDriver struct {
	BaseDriver
	config *mysql.Config
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
		BaseDriver: BaseDriver{
			Db: sql.OpenDB(connector),
		},
	}

	return driver, nil
}

func (m *mysqlDriver) Databases() ([]Database, error) {
	databases := []Database{}
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

func (m *mysqlDriver) SelectDatabase(db Database) error {
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

func (m *mysqlDriver) Tables() ([]Table, error) {
	if m.config.DBName == "" {
		return nil, errors.New("no database selected")
	}
	tables := []Table{}
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
