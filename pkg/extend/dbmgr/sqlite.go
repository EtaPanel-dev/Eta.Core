package dbmgr

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
)

// SQLiteManager handles SQLite database operations
type SQLiteManager struct {
	DatabasePath string
	conn         *sql.DB
}

// SQLiteConfig represents SQLite connection configuration
type SQLiteConfig struct {
	DatabasePath string `json:"database_path" toml:"database_path"`
}

// SQLiteDatabaseInfo represents SQLite database information
type SQLiteDatabaseInfo struct {
	Name   string `json:"name"`
	Size   string `json:"size"`
	Tables int    `json:"tables"`
	Path   string `json:"path"`
}

// SQLiteTableInfo represents table information
type SQLiteTableInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Rows    int64  `json:"rows"`
	Size    string `json:"size"`
	Columns int    `json:"columns"`
}

// NewSQLiteManager creates a new SQLite manager instance
func NewSQLiteManager(config SQLiteConfig) *SQLiteManager {
	return &SQLiteManager{
		DatabasePath: config.DatabasePath,
	}
}

// Connect establishes connection to SQLite database
func (s *SQLiteManager) Connect() error {
	// Ensure directory exists
	dir := filepath.Dir(s.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", s.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Enable foreign keys and other optimizations
	_, err = conn.Exec(`
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA cache_size = 1000;
		PRAGMA temp_store = memory;
	`)
	if err != nil {
		return fmt.Errorf("failed to set SQLite pragmas: %w", err)
	}

	s.conn = conn
	log.Printf("Successfully connected to SQLite database at %s", s.DatabasePath)
	return nil
}

// Disconnect closes the SQLite connection
func (s *SQLiteManager) Disconnect() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// TestConnection tests the SQLite connection
func (s *SQLiteManager) TestConnection() error {
	if s.conn == nil {
		return fmt.Errorf("no active connection")
	}
	return s.conn.Ping()
}

// GetDatabaseInfo returns information about the current database
func (s *SQLiteManager) GetDatabaseInfo() (*SQLiteDatabaseInfo, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	info := &SQLiteDatabaseInfo{
		Name: filepath.Base(s.DatabasePath),
		Path: s.DatabasePath,
	}

	// Get file size
	if stat, err := os.Stat(s.DatabasePath); err == nil {
		info.Size = fmt.Sprintf("%.2f MB", float64(stat.Size())/1024/1024)
	}

	// Get table count
	query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table'"
	err := s.conn.QueryRow(query).Scan(&info.Tables)
	if err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}

	return info, nil
}

// GetTables returns list of tables in the database
func (s *SQLiteManager) GetTables() ([]SQLiteTableInfo, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := "SELECT name, type FROM sqlite_master WHERE type='table' ORDER BY name"
	rows, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var tables []SQLiteTableInfo
	for rows.Next() {
		var table SQLiteTableInfo
		var tableType string

		err := rows.Scan(&table.Name, &tableType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		table.Type = tableType

		// Get row count
		rowQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table.Name)
		err = s.conn.QueryRow(rowQuery).Scan(&table.Rows)
		if err != nil {
			table.Rows = 0 // If we can't get row count, set to 0
		}

		// Get column count
		colQuery := fmt.Sprintf("PRAGMA table_info(`%s`)", table.Name)
		colRows, err := s.conn.Query(colQuery)
		if err == nil {
			colCount := 0
			for colRows.Next() {
				colCount++
			}
			table.Columns = colCount
			if err := colRows.Close(); err != nil {
				log.Printf("Error closing column rows: %v", err)
			}
		}

		table.Size = "N/A" // SQLite doesn't provide per-table size easily

		tables = append(tables, table)
	}

	return tables, nil
}

// CreateTable creates a new table with the given SQL
func (s *SQLiteManager) CreateTable(sql string) error {
	if s.conn == nil {
		return fmt.Errorf("no active connection")
	}

	_, err := s.conn.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("Table created successfully")
	return nil
}

// DropTable drops a table
func (s *SQLiteManager) DropTable(tableName string) error {
	if s.conn == nil {
		return fmt.Errorf("no active connection")
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
	_, err := s.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	log.Printf("Table '%s' dropped successfully", tableName)
	return nil
}

// ExecuteQuery executes a custom SQL query
func (s *SQLiteManager) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	rows, err := s.conn.Query(query)
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
			val := valuePtrs[i].(*interface{})
			row[col] = *val
		}
		results = append(results, row)
	}

	return results, nil
}

// ExecuteStatement executes a statement (INSERT, UPDATE, DELETE)
func (s *SQLiteManager) ExecuteStatement(statement string) (int64, error) {
	if s.conn == nil {
		return 0, fmt.Errorf("no active connection")
	}

	result, err := s.conn.Exec(statement)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// BackupDatabase creates a backup of the database
func (s *SQLiteManager) BackupDatabase(backupPath string) error {
	if s.conn == nil {
		return fmt.Errorf("no active connection")
	}

	// Ensure backup directory exists
	dir := filepath.Dir(backupPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Simple file copy for SQLite backup
	sourceFile, err := os.Open(s.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			log.Printf("Error closing source file: %v", err)
		}
	}()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			log.Printf("Error closing destination file: %v", err)
		}
	}()

	// Copy file contents
	buffer := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := sourceFile.Read(buffer)
		if n == 0 || err != nil {
			break
		}
		if _, err := destFile.Write(buffer[:n]); err != nil {
			return fmt.Errorf("failed to write to backup file: %w", err)
		}
	}

	log.Printf("Database backed up to %s", backupPath)
	return nil
}

// GetTableSchema returns the schema for a specific table
func (s *SQLiteManager) GetTableSchema(tableName string) ([]map[string]interface{}, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := fmt.Sprintf("PRAGMA table_info(`%s`)", tableName)
	return s.ExecuteQuery(query)
}

// VacuumDatabase optimizes the database
func (s *SQLiteManager) VacuumDatabase() error {
	if s.conn == nil {
		return fmt.Errorf("no active connection")
	}

	_, err := s.conn.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	log.Printf("Database vacuumed successfully")
	return nil
}
