package dbmgr

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLManager handles MySQL database operations
type MySQLManager struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	conn     *sql.DB
}

// MySQLConfig represents MySQL connection configuration
type MySQLConfig struct {
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Username string `json:"username" toml:"username"`
	Password string `json:"password" toml:"password"`
	Database string `json:"database" toml:"database"`
}

// MySQLDatabaseInfo represents MySQL database information
type MySQLDatabaseInfo struct {
	Name      string `json:"name"`
	Size      string `json:"size"`
	Tables    int    `json:"tables"`
	Charset   string `json:"charset"`
	Collation string `json:"collation"`
}

// TableInfo represents table information
type TableInfo struct {
	Name    string `json:"name"`
	Engine  string `json:"engine"`
	Rows    int64  `json:"rows"`
	Size    string `json:"size"`
	Comment string `json:"comment"`
}

// UserInfo represents MySQL user information
type UserInfo struct {
	User string `json:"user"`
	Host string `json:"host"`
}

// NewMySQLManager creates a new MySQL manager instance
func NewMySQLManager(config MySQLConfig) *MySQLManager {
	return &MySQLManager{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		Database: config.Database,
	}
}

// Connect establishes connection to MySQL server
func (m *MySQLManager) Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.Username, m.Password, m.Host, m.Port, m.Database)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL server: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	m.conn = conn
	log.Printf("Successfully connected to MySQL server at %s:%d", m.Host, m.Port)
	return nil
}

// Disconnect closes the MySQL connection
func (m *MySQLManager) Disconnect() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// TestConnection tests the MySQL connection
func (m *MySQLManager) TestConnection() error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}
	return m.conn.Ping()
}

// GetDatabases returns list of all databases
func (m *MySQLManager) GetDatabases() ([]MySQLDatabaseInfo, error) {
	if m.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			SCHEMA_NAME as name,
			ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS size_mb,
			COUNT(TABLE_NAME) as table_count,
			DEFAULT_CHARACTER_SET_NAME as charset,
			DEFAULT_COLLATION_NAME as collation
		FROM information_schema.SCHEMATA s
		LEFT JOIN information_schema.TABLES t ON s.SCHEMA_NAME = t.TABLE_SCHEMA
		WHERE SCHEMA_NAME NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
		GROUP BY SCHEMA_NAME, DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME
		ORDER BY SCHEMA_NAME
	`

	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []MySQLDatabaseInfo
	for rows.Next() {
		var db MySQLDatabaseInfo
		var sizeMB sql.NullFloat64
		var tableCount sql.NullInt64

		err := rows.Scan(&db.Name, &sizeMB, &tableCount, &db.Charset, &db.Collation)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database row: %w", err)
		}

		if sizeMB.Valid {
			db.Size = fmt.Sprintf("%.2f MB", sizeMB.Float64)
		} else {
			db.Size = "0 MB"
		}

		if tableCount.Valid {
			db.Tables = int(tableCount.Int64)
		}

		databases = append(databases, db)
	}

	return databases, nil
}

// GetTables returns list of tables in the current database
func (m *MySQLManager) GetTables() ([]TableInfo, error) {
	if m.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			TABLE_NAME as name,
			ENGINE as engine,
			TABLE_ROWS as rows,
			ROUND((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) AS size_mb,
			TABLE_COMMENT as comment
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`

	rows, err := m.conn.Query(query, m.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		var sizeMB sql.NullFloat64
		var rowCount sql.NullInt64

		err := rows.Scan(&table.Name, &table.Engine, &rowCount, &sizeMB, &table.Comment)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		if rowCount.Valid {
			table.Rows = rowCount.Int64
		}

		if sizeMB.Valid {
			table.Size = fmt.Sprintf("%.2f MB", sizeMB.Float64)
		} else {
			table.Size = "0 MB"
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// CreateDatabase creates a new database
func (m *MySQLManager) CreateDatabase(name, charset, collation string) error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if charset == "" {
		charset = "utf8mb4"
	}
	if collation == "" {
		collation = "utf8mb4_unicode_ci"
	}

	query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s", name, charset, collation)
	_, err := m.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database '%s' created successfully", name)
	return nil
}

// DropDatabase drops a database
func (m *MySQLManager) DropDatabase(name string) error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}

	query := fmt.Sprintf("DROP DATABASE `%s`", name)
	_, err := m.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	log.Printf("Database '%s' dropped successfully", name)
	return nil
}

// CreateUser creates a new MySQL user
func (m *MySQLManager) CreateUser(username, password, host string) error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if host == "" {
		host = "%"
	}

	query := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err := m.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User '%s'@'%s' created successfully", username, host)
	return nil
}

// DropUser drops a MySQL user
func (m *MySQLManager) DropUser(username, host string) error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if host == "" {
		host = "%"
	}

	query := fmt.Sprintf("DROP USER '%s'@'%s'", username, host)
	_, err := m.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop user: %w", err)
	}

	log.Printf("User '%s'@'%s' dropped successfully", username, host)
	return nil
}

// GrantPrivileges grants privileges to a user on a database
func (m *MySQLManager) GrantPrivileges(username, host, database, privileges string) error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if host == "" {
		host = "%"
	}
	if privileges == "" {
		privileges = "ALL PRIVILEGES"
	}

	query := fmt.Sprintf("GRANT %s ON `%s`.* TO '%s'@'%s'", privileges, database, username, host)
	_, err := m.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	// Flush privileges
	_, err = m.conn.Exec("FLUSH PRIVILEGES")
	if err != nil {
		return fmt.Errorf("failed to flush privileges: %w", err)
	}

	log.Printf("Granted %s on '%s' to '%s'@'%s'", privileges, database, username, host)
	return nil
}

// RevokePrivileges revokes privileges from a user on a database
func (m *MySQLManager) RevokePrivileges(username, host, database, privileges string) error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if host == "" {
		host = "%"
	}
	if privileges == "" {
		privileges = "ALL PRIVILEGES"
	}

	query := fmt.Sprintf("REVOKE %s ON `%s`.* FROM '%s'@'%s'", privileges, database, username, host)
	_, err := m.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to revoke privileges: %w", err)
	}

	// Flush privileges
	_, err = m.conn.Exec("FLUSH PRIVILEGES")
	if err != nil {
		return fmt.Errorf("failed to flush privileges: %w", err)
	}

	log.Printf("Revoked %s on '%s' from '%s'@'%s'", privileges, database, username, host)
	return nil
}

// GetUsers returns list of MySQL users
func (m *MySQLManager) GetUsers() ([]UserInfo, error) {
	if m.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := "SELECT User, Host FROM mysql.user WHERE User != '' ORDER BY User, Host"
	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var user UserInfo
		err := rows.Scan(&user.User, &user.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// ExecuteQuery executes a custom SQL query
func (m *MySQLManager) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if m.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, nil
}

// ExecuteNonQuery executes a non-query SQL statement (INSERT, UPDATE, DELETE)
func (m *MySQLManager) ExecuteNonQuery(query string) (int64, error) {
	if m.conn == nil {
		return 0, fmt.Errorf("no active connection")
	}

	result, err := m.conn.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute non-query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// BackupDatabase creates a backup of the specified database
func (m *MySQLManager) BackupDatabase(database, outputPath string) error {
	if m.conn == nil {
		return fmt.Errorf("no active connection")
	}

	// This is a simplified backup - in production, you'd want to use mysqldump
	// or implement a more sophisticated backup mechanism
	tables, err := m.GetTablesForDatabase(database)
	if err != nil {
		return fmt.Errorf("failed to get tables for backup: %w", err)
	}

	var backupSQL strings.Builder
	backupSQL.WriteString(fmt.Sprintf("-- Backup of database: %s\n", database))
	backupSQL.WriteString(fmt.Sprintf("-- Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	for _, table := range tables {
		// Get CREATE TABLE statement
		createQuery := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", database, table.Name)
		rows, err := m.conn.Query(createQuery)
		if err != nil {
			continue
		}

		if rows.Next() {
			var tableName, createStmt string
			rows.Scan(&tableName, &createStmt)
			backupSQL.WriteString(fmt.Sprintf("%s;\n\n", createStmt))
		}
		rows.Close()
	}

	// Write to file (simplified - you'd want better file handling)
	log.Printf("Backup SQL generated for database '%s'", database)
	return nil
}

// GetTablesForDatabase returns tables for a specific database
func (m *MySQLManager) GetTablesForDatabase(database string) ([]TableInfo, error) {
	if m.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			TABLE_NAME as name,
			ENGINE as engine,
			TABLE_ROWS as rows,
			ROUND((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) AS size_mb,
			TABLE_COMMENT as comment
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`

	rows, err := m.conn.Query(query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		var sizeMB sql.NullFloat64
		var rowCount sql.NullInt64

		err := rows.Scan(&table.Name, &table.Engine, &rowCount, &sizeMB, &table.Comment)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		if rowCount.Valid {
			table.Rows = rowCount.Int64
		}

		if sizeMB.Valid {
			table.Size = fmt.Sprintf("%.2f MB", sizeMB.Float64)
		} else {
			table.Size = "0 MB"
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// GetServerStatus returns MySQL server status information
func (m *MySQLManager) GetServerStatus() (map[string]interface{}, error) {
	if m.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	status := make(map[string]interface{})

	// Get version
	var version string
	err := m.conn.QueryRow("SELECT VERSION()").Scan(&version)
	if err == nil {
		status["version"] = version
	}

	// Get uptime
	var uptime int64
	err = m.conn.QueryRow("SHOW STATUS LIKE 'Uptime'").Scan(nil, &uptime)
	if err == nil {
		status["uptime"] = uptime
	}

	// Get connection count
	var connections int64
	err = m.conn.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(nil, &connections)
	if err == nil {
		status["connections"] = connections
	}

	return status, nil
}
