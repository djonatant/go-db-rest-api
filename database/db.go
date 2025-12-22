package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

type DBType string

const (
	MySQL     DBType = "mysql"
	Postgres  DBType = "postgres"
	SQLServer DBType = "sqlserver"
)

type DBConfig struct {
	Type     DBType `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type ConnectionManager struct {
	connections map[string]*sql.DB
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*sql.DB),
	}
}

func (m *ConnectionManager) Connect(id string, config DBConfig) error {
	db, err := OpenConnection(config)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[id] = db

	return nil
}

func OpenConnection(config DBConfig) (*sql.DB, error) {
	var dsn string
	var driver string

	switch config.Type {
	case MySQL:
		driver = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Password, config.Host, config.Port, config.DBName)
	case Postgres:
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.DBName)
	case SQLServer:
		driver = "sqlserver"
		dsn = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", config.Host, config.User, config.Password, config.Port, config.DBName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func (m *ConnectionManager) GetConnection(id string) (*sql.DB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	db, ok := m.connections[id]
	if !ok {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	return db, nil
}

func (m *ConnectionManager) CloseConnection(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	db, ok := m.connections[id]
	if !ok {
		return nil
	}

	delete(m.connections, id)
	return db.Close()
}
