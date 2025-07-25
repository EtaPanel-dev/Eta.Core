package dbmgr

import (
	"fmt"
	"log"
	"time"
)

// Example demonstrates how to use MySQL and Redis managers
func Example() {
	// MySQL Manager Example
	fmt.Println("=== MySQL Manager Example ===")

	mysqlConfig := MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "test_db",
	}

	mysqlMgr := NewMySQLManager(mysqlConfig)

	// Connect to MySQL
	if err := mysqlMgr.Connect(); err != nil {
		log.Printf("Failed to connect to MySQL: %v", err)
	} else {
		defer mysqlMgr.Disconnect()

		// Get databases
		if databases, err := mysqlMgr.GetDatabases(); err == nil {
			fmt.Printf("Found %d databases\n", len(databases))
			for _, db := range databases {
				fmt.Printf("- %s (%s, %d tables)\n", db.Name, db.Size, db.Tables)
			}
		}

		// Get tables
		if tables, err := mysqlMgr.GetTables(); err == nil {
			fmt.Printf("Found %d tables in current database\n", len(tables))
			for _, table := range tables {
				fmt.Printf("- %s (%s, %d rows)\n", table.Name, table.Size, table.Rows)
			}
		}

		// Create a test database
		if err := mysqlMgr.CreateDatabase("test_new_db", "utf8mb4", "utf8mb4_unicode_ci"); err != nil {
			log.Printf("Failed to create database: %v", err)
		}

		// Create a test user
		if err := mysqlMgr.CreateUser("test_user", "test_password", "localhost"); err != nil {
			log.Printf("Failed to create user: %v", err)
		}

		// Grant privileges
		if err := mysqlMgr.GrantPrivileges("test_user", "localhost", "test_new_db", "ALL PRIVILEGES"); err != nil {
			log.Printf("Failed to grant privileges: %v", err)
		}

		// Execute custom query
		if results, err := mysqlMgr.ExecuteQuery("SHOW VARIABLES LIKE 'version%'"); err == nil {
			fmt.Printf("MySQL version info:\n")
			for _, row := range results {
				fmt.Printf("- %s: %s\n", row["Variable_name"], row["Value"])
			}
		}
	}

	fmt.Println("\n=== Redis Manager Example ===")

	// Redis Manager Example
	redisConfig := RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
	}

	redisMgr := NewRedisManager(redisConfig)

	// Connect to Redis
	if err := redisMgr.Connect(); err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		defer redisMgr.Disconnect()

		// Get Redis info
		if info, err := redisMgr.GetInfo(); err == nil {
			fmt.Printf("Redis Version: %s\n", info.Version)
			fmt.Printf("Connected Clients: %d\n", info.ConnectedClients)
			fmt.Printf("Used Memory: %s\n", info.UsedMemoryHuman)
			fmt.Printf("Total Keys: %d\n", info.TotalKeys)
			fmt.Printf("Uptime: %d seconds\n", info.Uptime)
		}

		// Get databases info
		if databases, err := redisMgr.GetDatabases(); err == nil {
			fmt.Printf("Redis Databases:\n")
			for _, db := range databases {
				if db.Keys > 0 {
					fmt.Printf("- DB%d: %d keys, %d with expiration\n", db.Database, db.Keys, db.Expires)
				}
			}
		}

		// Set some test keys
		redisMgr.SetKey("test:string", "Hello Redis!", 0)
		redisMgr.SetKey("test:expiring", "This will expire", 60*time.Second)

		// Get keys with pattern
		if keys, _, err := redisMgr.GetKeys("test:*", 0, 10); err == nil {
			fmt.Printf("Keys matching 'test:*':\n")
			for _, key := range keys {
				if keyInfo, err := redisMgr.GetKeyInfo(key, true); err == nil {
					fmt.Printf("- %s (%s): %v\n", keyInfo.Key, keyInfo.Type, keyInfo.Value)
					if keyInfo.TTL > 0 {
						fmt.Printf("  TTL: %d seconds\n", keyInfo.TTL)
					}
				}
			}
		}

		// Execute Redis command
		if result, err := redisMgr.ExecuteCommand("PING"); err == nil {
			fmt.Printf("PING result: %v\n", result)
		}

		// Get memory stats
		if memStats, err := redisMgr.GetMemoryStats(); err == nil {
			fmt.Printf("Memory stats available: %d entries\n", len(memStats))
		}

		// Get slow log
		if slowLogs, err := redisMgr.GetSlowLog(5); err == nil {
			if len(slowLogs) > 0 {
				fmt.Printf("Recent slow queries:\n")
				for _, log := range slowLogs {
					fmt.Printf("- ID: %v, Duration: %v μs, Command: %v\n",
						log["id"], log["duration"], log["command"])
				}
			} else {
				fmt.Println("No slow queries found")
			}
		}

		// Get command statistics (alternative to monitoring)
		if cmdStats, err := redisMgr.GetCommandStats(); err == nil {
			fmt.Printf("Command statistics available for %d commands\n", len(cmdStats))
			// Show top 3 most used commands
			count := 0
			for cmd, stats := range cmdStats {
				if count >= 3 {
					break
				}
				fmt.Printf("- %s: %v\n", cmd, stats)
				count++
			}
		}

		// Clean up test keys
		redisMgr.DeleteKeys([]string{"test:string", "test:expiring"})
	}
}

// HealthCheck performs health checks on both MySQL and Redis
func HealthCheck(mysqlConfig MySQLConfig, redisConfig RedisConfig) map[string]bool {
	status := make(map[string]bool)

	// Check MySQL
	mysqlMgr := NewMySQLManager(mysqlConfig)
	if err := mysqlMgr.Connect(); err == nil {
		status["mysql"] = mysqlMgr.TestConnection() == nil
		mysqlMgr.Disconnect()
	} else {
		status["mysql"] = false
	}

	// Check Redis
	redisMgr := NewRedisManager(redisConfig)
	if err := redisMgr.Connect(); err == nil {
		status["redis"] = redisMgr.TestConnection() == nil
		redisMgr.Disconnect()
	} else {
		status["redis"] = false
	}

	return status
}

// BatchOperations demonstrates batch operations
func BatchOperations() {
	fmt.Println("=== Batch Operations Example ===")

	// Redis batch operations
	redisConfig := RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
	}

	redisMgr := NewRedisManager(redisConfig)
	if err := redisMgr.Connect(); err == nil {
		defer redisMgr.Disconnect()

		// Set multiple keys
		keys := []string{"batch:1", "batch:2", "batch:3", "batch:4", "batch:5"}
		for i, key := range keys {
			redisMgr.SetKey(key, fmt.Sprintf("value_%d", i+1), 0)
		}

		// Get all batch keys
		if batchKeys, _, err := redisMgr.GetKeys("batch:*", 0, 100); err == nil {
			fmt.Printf("Created %d batch keys\n", len(batchKeys))
		}

		// Delete all batch keys
		if deleted, err := redisMgr.DeleteKeys(keys); err == nil {
			fmt.Printf("Deleted %d keys\n", deleted)
		}
	}

	// MySQL batch operations
	mysqlConfig := MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "test",
	}

	mysqlMgr := NewMySQLManager(mysqlConfig)
	if err := mysqlMgr.Connect(); err == nil {
		defer mysqlMgr.Disconnect()

		// Create multiple test tables
		tables := []string{"test_table_1", "test_table_2", "test_table_3"}
		for _, table := range tables {
			query := fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					id INT AUTO_INCREMENT PRIMARY KEY,
					name VARCHAR(100),
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`, table)

			if _, err := mysqlMgr.ExecuteNonQuery(query); err == nil {
				fmt.Printf("Created table: %s\n", table)
			}
		}

		// Drop test tables
		for _, table := range tables {
			query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
			if _, err := mysqlMgr.ExecuteNonQuery(query); err == nil {
				fmt.Printf("Dropped table: %s\n", table)
			}
		}
	}
}
