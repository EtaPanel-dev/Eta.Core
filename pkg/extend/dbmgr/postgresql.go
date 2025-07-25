package dbmgr

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLManager handles PostgreSQL database operations
type PostgreSQLManager struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
	conn     *sql.DB
}

// PostgreSQLConfig represents PostgreSQL connection configuration
type PostgreSQLConfig struct {
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Username string `json:"username" toml:"username"`
	Password string `json:"password" toml:"password"`
	Database string `json:"database" toml:"database"`
	SSLMode  string `json:"ssl_mode" toml:"ssl_mode"`
}

// PostgreSQLDatabaseInfo represents PostgreSQL database information
type PostgreSQLDatabaseInfo struct {
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	Encoding string `json:"encoding"`
	Collate  string `json:"collate"`
	CType    string `json:"ctype"`
	Size     string `json:"size"`
	Tables   int    `json:"tables"`
}

// PostgreSQLTableInfo represents PostgreSQL table information
type PostgreSQLTableInfo struct {
	Name      string `json:"name"`
	Schema    string `json:"schema"`
	Owner     string `json:"owner"`
	Rows      int64  `json:"rows"`
	Size      string `json:"size"`
	Comment   string `json:"comment"`
	TableType string `json:"table_type"`
}

// PostgreSQLUserInfo represents PostgreSQL user information
type PostgreSQLUserInfo struct {
	Username    string `json:"username"`
	CreateDB    bool   `json:"create_db"`
	CreateRole  bool   `json:"create_role"`
	Superuser   bool   `json:"superuser"`
	CanLogin    bool   `json:"can_login"`
	Replication bool   `json:"replication"`
}

// PostgreSQLSchemaInfo represents PostgreSQL schema information
type PostgreSQLSchemaInfo struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// NewPostgreSQLManager creates a new PostgreSQL manager instance
func NewPostgreSQLManager(config PostgreSQLConfig) *PostgreSQLManager {
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	return &PostgreSQLManager{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		Database: config.Database,
		SSLMode:  config.SSLMode,
	}
}

// Connect establishes connection to PostgreSQL server
func (p *PostgreSQLManager) Connect() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.Username, p.Password, p.Database, p.SSLMode)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL server: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	p.conn = conn
	log.Printf("成功连接到 PostgreSQL 服务器，地址 %s:%d", p.Host, p.Port)
	return nil
}

// Disconnect closes the PostgreSQL connection
func (p *PostgreSQLManager) Disconnect() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// TestConnection tests the PostgreSQL connection
func (p *PostgreSQLManager) TestConnection() error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}
	return p.conn.Ping()
}

// GetDatabases returns list of all databases
func (p *PostgreSQLManager) GetDatabases() ([]PostgreSQLDatabaseInfo, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			d.datname as name,
			pg_catalog.pg_get_userbyid(d.datdba) as owner,
			pg_catalog.pg_encoding_to_char(d.encoding) as encoding,
			d.datcollate as collate,
			d.datctype as ctype,
			pg_catalog.pg_size_pretty(pg_catalog.pg_database_size(d.datname)) as size,
			COALESCE(t.table_count, 0) as table_count
		FROM pg_catalog.pg_database d
		LEFT JOIN (
			SELECT 
				schemaname,
				COUNT(*) as table_count
			FROM pg_catalog.pg_tables 
			WHERE schemaname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
			GROUP BY schemaname
		) t ON d.datname = t.schemaname
		WHERE d.datistemplate = false
		ORDER BY d.datname
	`

	rows, err := p.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []PostgreSQLDatabaseInfo
	for rows.Next() {
		var db PostgreSQLDatabaseInfo
		var tableCount sql.NullInt64

		err := rows.Scan(&db.Name, &db.Owner, &db.Encoding, &db.Collate, &db.CType, &db.Size, &tableCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database row: %w", err)
		}

		if tableCount.Valid {
			db.Tables = int(tableCount.Int64)
		}

		databases = append(databases, db)
	}

	return databases, nil
}

// GetTables returns list of tables in the current database
func (p *PostgreSQLManager) GetTables() ([]PostgreSQLTableInfo, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			t.tablename as name,
			t.schemaname as schema,
			t.tableowner as owner,
			COALESCE(s.n_tup_ins - s.n_tup_del, 0) as rows,
			pg_size_pretty(pg_total_relation_size(c.oid)) as size,
			COALESCE(obj_description(c.oid), '') as comment,
			'TABLE' as table_type
		FROM pg_catalog.pg_tables t
		LEFT JOIN pg_catalog.pg_class c ON c.relname = t.tablename
		LEFT JOIN pg_catalog.pg_stat_user_tables s ON s.relname = t.tablename AND s.schemaname = t.schemaname
		WHERE t.schemaname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
		UNION ALL
		SELECT 
			v.viewname as name,
			v.schemaname as schema,
			v.viewowner as owner,
			0 as rows,
			'0 bytes' as size,
			'' as comment,
			'VIEW' as table_type
		FROM pg_catalog.pg_views v
		WHERE v.schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY name
	`

	rows, err := p.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []PostgreSQLTableInfo
	for rows.Next() {
		var table PostgreSQLTableInfo

		err := rows.Scan(&table.Name, &table.Schema, &table.Owner, &table.Rows, &table.Size, &table.Comment, &table.TableType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// GetSchemas returns list of schemas in the current database
func (p *PostgreSQLManager) GetSchemas() ([]PostgreSQLSchemaInfo, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			schema_name as name,
			schema_owner as owner
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast', 'pg_temp_1', 'pg_toast_temp_1')
		ORDER BY schema_name
	`

	rows, err := p.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []PostgreSQLSchemaInfo
	for rows.Next() {
		var schema PostgreSQLSchemaInfo
		err := rows.Scan(&schema.Name, &schema.Owner)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema row: %w", err)
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// CreateDatabase creates a new database
func (p *PostgreSQLManager) CreateDatabase(name, owner, encoding, collate, ctype string) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if encoding == "" {
		encoding = "UTF8"
	}
	if collate == "" {
		collate = "en_US.UTF-8"
	}
	if ctype == "" {
		ctype = "en_US.UTF-8"
	}

	var query string
	if owner != "" {
		query = fmt.Sprintf(`CREATE DATABASE "%s" WITH OWNER = "%s" ENCODING = '%s' LC_COLLATE = '%s' LC_CTYPE = '%s'`,
			name, owner, encoding, collate, ctype)
	} else {
		query = fmt.Sprintf(`CREATE DATABASE "%s" WITH ENCODING = '%s' LC_COLLATE = '%s' LC_CTYPE = '%s'`,
			name, encoding, collate, ctype)
	}

	_, err := p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("数据库 '%s' 创建成功", name)
	return nil
}

// DropDatabase drops a database
func (p *PostgreSQLManager) DropDatabase(name string) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	// Terminate connections to the database first
	terminateQuery := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`
	_, err := p.conn.Exec(terminateQuery, name)
	if err != nil {
		log.Printf("警告: 无法终止对数据库 '%s' 的连接: %v", name, err)
	}

	query := fmt.Sprintf(`DROP DATABASE "%s"`, name)
	_, err = p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	log.Printf("数据库 '%s' 删除成功", name)
	return nil
}

// CreateUser creates a new PostgreSQL user/role
func (p *PostgreSQLManager) CreateUser(username, password string, options map[string]bool) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	var optionStrings []string

	if options["superuser"] {
		optionStrings = append(optionStrings, "SUPERUSER")
	} else {
		optionStrings = append(optionStrings, "NOSUPERUSER")
	}

	if options["createdb"] {
		optionStrings = append(optionStrings, "CREATEDB")
	} else {
		optionStrings = append(optionStrings, "NOCREATEDB")
	}

	if options["createrole"] {
		optionStrings = append(optionStrings, "CREATEROLE")
	} else {
		optionStrings = append(optionStrings, "NOCREATEROLE")
	}

	if options["login"] {
		optionStrings = append(optionStrings, "LOGIN")
	} else {
		optionStrings = append(optionStrings, "NOLOGIN")
	}

	if options["replication"] {
		optionStrings = append(optionStrings, "REPLICATION")
	} else {
		optionStrings = append(optionStrings, "NOREPLICATION")
	}

	query := fmt.Sprintf(`CREATE ROLE "%s" WITH %s PASSWORD '%s'`, username, strings.Join(optionStrings, " "), password)
	_, err := p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("用户 '%s' 创建成功", username)
	return nil
}

// DropUser drops a PostgreSQL user/role
func (p *PostgreSQLManager) DropUser(username string) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	query := fmt.Sprintf(`DROP ROLE "%s"`, username)
	_, err := p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop user: %w", err)
	}

	log.Printf("用户 '%s' 删除成功", username)
	return nil
}

// GrantPrivileges grants privileges to a user on a database
func (p *PostgreSQLManager) GrantPrivileges(username, database, privileges string) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if privileges == "" {
		privileges = "ALL PRIVILEGES"
	}

	query := fmt.Sprintf(`GRANT %s ON DATABASE "%s" TO "%s"`, privileges, database, username)
	_, err := p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	log.Printf("授予 %s 在数据库 '%s' 上的权限给用户 '%s'", privileges, database, username)
	return nil
}

// RevokePrivileges revokes privileges from a user on a database
func (p *PostgreSQLManager) RevokePrivileges(username, database, privileges string) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	if privileges == "" {
		privileges = "ALL PRIVILEGES"
	}

	query := fmt.Sprintf(`REVOKE %s ON DATABASE "%s" FROM "%s"`, privileges, database, username)
	_, err := p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to revoke privileges: %w", err)
	}

	log.Printf("撤销用户 '%s' 在数据库 '%s' 上的 %s 权限", username, database, privileges)
	return nil
}

// GetUsers returns list of PostgreSQL users/roles
func (p *PostgreSQLManager) GetUsers() ([]PostgreSQLUserInfo, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			rolname as username,
			rolcreatedb as create_db,
			rolcreaterole as create_role,
			rolsuper as superuser,
			rolcanlogin as can_login,
			rolreplication as replication
		FROM pg_catalog.pg_roles
		WHERE rolname NOT LIKE 'pg_%'
		ORDER BY rolname
	`

	rows, err := p.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []PostgreSQLUserInfo
	for rows.Next() {
		var user PostgreSQLUserInfo
		err := rows.Scan(&user.Username, &user.CreateDB, &user.CreateRole, &user.Superuser, &user.CanLogin, &user.Replication)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// ExecuteQuery executes a custom SQL query
func (p *PostgreSQLManager) ExecuteQuery(query string) ([]map[string]interface{}, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	rows, err := p.conn.Query(query)
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
func (p *PostgreSQLManager) ExecuteNonQuery(query string) (int64, error) {
	if p.conn == nil {
		return 0, fmt.Errorf("no active connection")
	}

	result, err := p.conn.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to execute non-query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetServerStatus returns PostgreSQL server status information
func (p *PostgreSQLManager) GetServerStatus() (map[string]interface{}, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	status := make(map[string]interface{})

	// Get version
	var version string
	err := p.conn.QueryRow("SELECT version()").Scan(&version)
	if err == nil {
		status["version"] = version
	}

	// Get current database
	var currentDB string
	err = p.conn.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err == nil {
		status["current_database"] = currentDB
	}

	// Get server uptime (start time)
	var startTime time.Time
	err = p.conn.QueryRow("SELECT pg_postmaster_start_time()").Scan(&startTime)
	if err == nil {
		status["start_time"] = startTime
		status["uptime_seconds"] = time.Since(startTime).Seconds()
	}

	// Get active connections
	var activeConnections int
	err = p.conn.QueryRow("SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeConnections)
	if err == nil {
		status["active_connections"] = activeConnections
	}

	// Get total connections
	var totalConnections int
	err = p.conn.QueryRow("SELECT count(*) FROM pg_stat_activity").Scan(&totalConnections)
	if err == nil {
		status["total_connections"] = totalConnections
	}

	// Get database size
	var dbSize string
	err = p.conn.QueryRow("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize)
	if err == nil {
		status["database_size"] = dbSize
	}

	return status, nil
}

// GetConnections returns current database connections
func (p *PostgreSQLManager) GetConnections() ([]map[string]interface{}, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			pid,
			usename,
			application_name,
			client_addr,
			client_port,
			backend_start,
			state,
			query
		FROM pg_stat_activity
		WHERE datname = current_database()
		ORDER BY backend_start DESC
	`

	return p.ExecuteQuery(query)
}

// KillConnection terminates a database connection by PID
func (p *PostgreSQLManager) KillConnection(pid int) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	query := "SELECT pg_terminate_backend($1)"
	var result bool
	err := p.conn.QueryRow(query, pid).Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to kill connection: %w", err)
	}

	if !result {
		return fmt.Errorf("failed to terminate connection with PID %d", pid)
	}

	log.Printf("成功终止 PID 为 %d 的连接", pid)
	return nil
}

// GetTableSize returns size information for tables
func (p *PostgreSQLManager) GetTableSize() ([]map[string]interface{}, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			schemaname,
			tablename,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
			pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_size
		FROM pg_tables
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
	`

	return p.ExecuteQuery(query)
}

// BackupDatabase creates a logical backup using pg_dump command
func (p *PostgreSQLManager) BackupDatabase(database, outputPath string, options map[string]interface{}) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	// This would typically use pg_dump command
	// For now, return a placeholder implementation
	log.Printf("备份已启动，数据库: '%s'，输出路径: '%s'", database, outputPath)

	// In a real implementation, you would:
	// 1. Construct pg_dump command with appropriate options
	// 2. Execute the command using os/exec
	// 3. Handle the output and errors

	return fmt.Errorf("backup functionality requires pg_dump command integration")
}

// RestoreDatabase restores a database from backup using pg_restore command
func (p *PostgreSQLManager) RestoreDatabase(database, backupPath string, options map[string]interface{}) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	// This would typically use pg_restore command
	// For now, return a placeholder implementation
	log.Printf("恢复已启动，数据库: '%s'，备份路径: '%s'", database, backupPath)

	// In a real implementation, you would:
	// 1. Construct pg_restore command with appropriate options
	// 2. Execute the command using os/exec
	// 3. Handle the output and errors

	return fmt.Errorf("restore functionality requires pg_restore command integration")
}

// GetExtensions returns list of installed PostgreSQL extensions
func (p *PostgreSQLManager) GetExtensions() ([]map[string]interface{}, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}

	query := `
		SELECT 
			extname as name,
			extversion as version,
			nspname as schema
		FROM pg_extension e
		JOIN pg_namespace n ON e.extnamespace = n.oid
		ORDER BY extname
	`

	return p.ExecuteQuery(query)
}

// CreateExtension creates/installs a PostgreSQL extension
func (p *PostgreSQLManager) CreateExtension(name, schema string) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	var query string
	if schema != "" {
		query = fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s" WITH SCHEMA "%s"`, name, schema)
	} else {
		query = fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS "%s"`, name)
	}

	_, err := p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create extension: %w", err)
	}

	log.Printf("扩展 '%s' 创建成功", name)
	return nil
}

// DropExtension drops a PostgreSQL extension
func (p *PostgreSQLManager) DropExtension(name string, cascade bool) error {
	if p.conn == nil {
		return fmt.Errorf("no active connection")
	}

	var query string
	if cascade {
		query = fmt.Sprintf(`DROP EXTENSION IF EXISTS "%s" CASCADE`, name)
	} else {
		query = fmt.Sprintf(`DROP EXTENSION IF EXISTS "%s"`, name)
	}

	_, err := p.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop extension: %w", err)
	}

	log.Printf("扩展 '%s' 删除成功", name)
	return nil
}
