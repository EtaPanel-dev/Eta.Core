package dbmgr

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisManager handles Redis operations
type RedisManager struct {
	Host     string
	Port     int
	Password string
	Database int
	client   *redis.Client
	ctx      context.Context
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Password string `json:"password" toml:"password"`
	Database int    `json:"database" toml:"database"`
}

// RedisInfo represents Redis server information
type RedisInfo struct {
	Version          string            `json:"version"`
	Mode             string            `json:"mode"`
	Role             string            `json:"role"`
	ConnectedClients int               `json:"connected_clients"`
	UsedMemory       string            `json:"used_memory"`
	UsedMemoryHuman  string            `json:"used_memory_human"`
	TotalKeys        int64             `json:"total_keys"`
	Uptime           int64             `json:"uptime"`
	Stats            map[string]string `json:"stats"`
}

// KeyInfo represents Redis key information
type KeyInfo struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	TTL   int64       `json:"ttl"`
	Size  int64       `json:"size"`
	Value interface{} `json:"value,omitempty"`
}

// RedisDatabaseInfo represents Redis database information
type RedisDatabaseInfo struct {
	Database int   `json:"database"`
	Keys     int64 `json:"keys"`
	Expires  int64 `json:"expires"`
}

// NewRedisManager creates a new Redis manager instance
func NewRedisManager(config RedisConfig) *RedisManager {
	return &RedisManager{
		Host:     config.Host,
		Port:     config.Port,
		Password: config.Password,
		Database: config.Database,
		ctx:      context.Background(),
	}
}

// Connect establishes connection to Redis server
func (r *RedisManager) Connect() error {
	r.client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", r.Host, r.Port),
		Password:     r.Password,
		DB:           r.Database,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test the connection
	_, err := r.client.Ping(r.ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Successfully connected to Redis server at %s:%d", r.Host, r.Port)
	return nil
}

// Disconnect closes the Redis connection
func (r *RedisManager) Disconnect() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// TestConnection tests the Redis connection
func (r *RedisManager) TestConnection() error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}
	_, err := r.client.Ping(r.ctx).Result()
	return err
}

// GetInfo returns Redis server information
func (r *RedisManager) GetInfo() (*RedisInfo, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	info, err := r.client.Info(r.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	redisInfo := &RedisInfo{
		Stats: make(map[string]string),
	}

	// Parse info string
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "redis_version":
					redisInfo.Version = value
				case "redis_mode":
					redisInfo.Mode = value
				case "role":
					redisInfo.Role = value
				case "connected_clients":
					if clients, err := strconv.Atoi(value); err == nil {
						redisInfo.ConnectedClients = clients
					}
				case "used_memory":
					redisInfo.UsedMemory = value
				case "used_memory_human":
					redisInfo.UsedMemoryHuman = value
				case "uptime_in_seconds":
					if uptime, err := strconv.ParseInt(value, 10, 64); err == nil {
						redisInfo.Uptime = uptime
					}
				default:
					redisInfo.Stats[key] = value
				}
			}
		}
	}

	// Get total keys count
	totalKeys, err := r.GetTotalKeys()
	if err == nil {
		redisInfo.TotalKeys = totalKeys
	}

	return redisInfo, nil
}

// GetTotalKeys returns total number of keys across all databases
func (r *RedisManager) GetTotalKeys() (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("no active connection")
	}

	var totalKeys int64
	for i := 0; i < 16; i++ { // Redis默认有16个数据库
		// Create a temporary client for each database
		tempClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
			Password: r.Password,
			DB:       i,
		})

		size, err := tempClient.DBSize(r.ctx).Result()
		err = tempClient.Close()
		if err != nil {
			return 0, err
		}

		totalKeys += size
	}

	return totalKeys, nil
}

// GetDatabases returns information about all Redis databases
func (r *RedisManager) GetDatabases() ([]RedisDatabaseInfo, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	var databases []RedisDatabaseInfo

	for i := 0; i < 16; i++ {
		// 临时切换到指定数据库
		tempClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
			Password: r.Password,
			DB:       i,
		})

		keys, err := tempClient.DBSize(r.ctx).Result()
		if err != nil {
			err := tempClient.Close()
			if err != nil {
				return nil, err
			}
			continue
		}

		// 获取有过期时间的key数量（这个操作比较重，实际使用时可能需要优化）
		var expires int64
		if keys > 0 && keys < 1000 { // 只在key数量不太多时统计expires
			allKeys, err := tempClient.Keys(r.ctx, "*").Result()
			if err == nil {
				for _, key := range allKeys {
					ttl, err := tempClient.TTL(r.ctx, key).Result()
					if err == nil && ttl > 0 {
						expires++
					}
				}
			}
		}

		databases = append(databases, RedisDatabaseInfo{
			Database: i,
			Keys:     keys,
			Expires:  expires,
		})

		err = tempClient.Close()
		if err != nil {
			return nil, err
		}
	}

	return databases, nil
}

// SwitchDatabase switches to a different Redis database
func (r *RedisManager) SwitchDatabase(db int) error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	// 创建新的客户端连接到指定数据库
	newClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", r.Host, r.Port),
		Password:     r.Password,
		DB:           db,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// 测试新连接
	_, err := newClient.Ping(r.ctx).Result()
	if err != nil {
		err := newClient.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to switch to database %d: %w", db, err)
	}

	// 关闭旧连接
	err = r.client.Close()
	if err != nil {
		return err
	}

	// 更新连接
	r.client = newClient
	r.Database = db

	log.Printf("Switched to Redis database %d", db)
	return nil
}

// GetKeys returns keys matching the pattern with pagination
func (r *RedisManager) GetKeys(pattern string, cursor uint64, count int64) ([]string, uint64, error) {
	if r.client == nil {
		return nil, 0, fmt.Errorf("no active connection")
	}

	if pattern == "" {
		pattern = "*"
	}
	if count <= 0 {
		count = 100
	}

	keys, nextCursor, err := r.client.Scan(r.ctx, cursor, pattern, count).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to scan keys: %w", err)
	}

	return keys, nextCursor, nil
}

// GetKeyInfo returns detailed information about a key
func (r *RedisManager) GetKeyInfo(key string, includeValue bool) (*KeyInfo, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	// Check if key exists
	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check key existence: %w", err)
	}
	if exists == 0 {
		return nil, fmt.Errorf("key does not exist")
	}

	keyInfo := &KeyInfo{
		Key: key,
	}

	// Get key type
	keyType, err := r.client.Type(r.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get key type: %w", err)
	}
	keyInfo.Type = keyType

	// Get TTL
	ttl, err := r.client.TTL(r.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get key TTL: %w", err)
	}
	keyInfo.TTL = int64(ttl.Seconds())

	// Get memory usage (Redis 4.0+)
	memUsage, err := r.client.MemoryUsage(r.ctx, key).Result()
	if err == nil {
		keyInfo.Size = memUsage
	}

	// Get value if requested
	if includeValue {
		value, err := r.getKeyValue(key, keyType)
		if err == nil {
			keyInfo.Value = value
		}
	}

	return keyInfo, nil
}

// getKeyValue returns the value of a key based on its type
func (r *RedisManager) getKeyValue(key, keyType string) (interface{}, error) {
	switch keyType {
	case "string":
		return r.client.Get(r.ctx, key).Result()
	case "list":
		return r.client.LRange(r.ctx, key, 0, -1).Result()
	case "set":
		return r.client.SMembers(r.ctx, key).Result()
	case "zset":
		return r.client.ZRangeWithScores(r.ctx, key, 0, -1).Result()
	case "hash":
		return r.client.HGetAll(r.ctx, key).Result()
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// SetKey sets a key-value pair
func (r *RedisManager) SetKey(key, value string, expiration time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}

	log.Printf("Key '%s' set successfully", key)
	return nil
}

// DeleteKey deletes a key
func (r *RedisManager) DeleteKey(key string) error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	deleted, err := r.client.Del(r.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	if deleted == 0 {
		return fmt.Errorf("key does not exist")
	}

	log.Printf("Key '%s' deleted successfully", key)
	return nil
}

// DeleteKeys deletes multiple keys
func (r *RedisManager) DeleteKeys(keys []string) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("no active connection")
	}

	if len(keys) == 0 {
		return 0, nil
	}

	deleted, err := r.client.Del(r.ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to delete keys: %w", err)
	}

	log.Printf("Deleted %d keys", deleted)
	return deleted, nil
}

// SetExpire sets expiration time for a key
func (r *RedisManager) SetExpire(key string, expiration time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	success, err := r.client.Expire(r.ctx, key, expiration).Result()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	if !success {
		return fmt.Errorf("key does not exist")
	}

	log.Printf("Expiration set for key '%s'", key)
	return nil
}

// RemoveExpire removes expiration from a key
func (r *RedisManager) RemoveExpire(key string) error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	success, err := r.client.Persist(r.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to remove expiration: %w", err)
	}

	if !success {
		return fmt.Errorf("key does not exist or has no expiration")
	}

	log.Printf("Expiration removed from key '%s'", key)
	return nil
}

// FlushDatabase clears all keys in current database
func (r *RedisManager) FlushDatabase() error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	err := r.client.FlushDB(r.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush database: %w", err)
	}

	log.Printf("Database %d flushed successfully", r.Database)
	return nil
}

// FlushAll clears all keys in all databases
func (r *RedisManager) FlushAll() error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	err := r.client.FlushAll(r.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all databases: %w", err)
	}

	log.Println("All databases flushed successfully")
	return nil
}

// ExecuteCommand executes a Redis command
func (r *RedisManager) ExecuteCommand(cmd string, args ...interface{}) (interface{}, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	// Prepare arguments: cmd first, then the rest
	cmdArgs := make([]interface{}, 0, len(args)+1)
	cmdArgs = append(cmdArgs, cmd)
	cmdArgs = append(cmdArgs, args...)

	result, err := r.client.Do(r.ctx, cmdArgs...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return result, nil
}

// GetSlowLog returns Redis slow log entries
func (r *RedisManager) GetSlowLog(count int64) ([]map[string]interface{}, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	if count <= 0 {
		count = 10
	}

	slowLogs, err := r.client.SlowLogGet(r.ctx, count).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get slow log: %w", err)
	}

	var logs []map[string]interface{}
	for _, logText := range slowLogs {
		logEntry := map[string]interface{}{
			"id":        logText.ID,
			"timestamp": logText.Time.Unix(),
			"duration":  logText.Duration.Microseconds(),
			"command":   strings.Join(logText.Args, " "),
		}
		logs = append(logs, logEntry)
	}

	return logs, nil
}

// GetConfig returns Redis configuration
func (r *RedisManager) GetConfig(pattern string) (map[string]string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	if pattern == "" {
		pattern = "*"
	}

	config, err := r.client.ConfigGet(r.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// ConfigGet already returns a map[string]string, so we can return it directly
	return config, nil
}

// SetConfig sets Redis configuration
func (r *RedisManager) SetConfig(parameter, value string) error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	err := r.client.ConfigSet(r.ctx, parameter, value).Err()
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	log.Printf("Config parameter '%s' set to '%s'", parameter, value)
	return nil
}

// SaveConfig saves current configuration to disk
func (r *RedisManager) SaveConfig() error {
	if r.client == nil {
		return fmt.Errorf("no active connection")
	}

	err := r.client.ConfigRewrite(r.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Println("Configuration saved to disk")
	return nil
}

// GetMemoryStats returns Redis memory statistics
func (r *RedisManager) GetMemoryStats() (map[string]interface{}, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	// Use MEMORY STATS command (Redis 4.0+)
	result, err := r.client.Do(r.ctx, "MEMORY", "STATS").Result()
	if err != nil {
		// Fallback to INFO memory if MEMORY STATS is not available
		return r.getMemoryInfoFallback()
	}

	// Parse the result - MEMORY STATS returns an array of key-value pairs
	statsArray, ok := result.([]interface{})
	if !ok {
		return r.getMemoryInfoFallback()
	}

	stats := make(map[string]interface{})
	for i := 0; i < len(statsArray); i += 2 {
		if i+1 < len(statsArray) {
			key, keyOk := statsArray[i].(string)
			value := statsArray[i+1]
			if keyOk {
				stats[key] = value
			}
		}
	}

	return stats, nil
}

// getMemoryInfoFallback returns memory info from INFO command as fallback
func (r *RedisManager) getMemoryInfoFallback() (map[string]interface{}, error) {
	info, err := r.client.Info(r.ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	stats := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Try to convert numeric values
				if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats[key] = intVal
				} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					stats[key] = floatVal
				} else {
					stats[key] = value
				}
			}
		}
	}

	return stats, nil
}

// Monitor starts monitoring Redis commands (for debugging)
func (r *RedisManager) Monitor(ctx context.Context) (<-chan string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	// Create a separate connection for monitoring
	monitorClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
		Password: r.Password,
		DB:       r.Database,
	})

	stringCh := make(chan string, 100)

	go func() {
		defer close(stringCh)
		defer func(monitorClient *redis.Client) {
			err := monitorClient.Close()
			if err != nil {

			}
		}(monitorClient)

		// Start monitoring
		pubsub := monitorClient.Subscribe(ctx)
		defer func(pubsub *redis.PubSub) {
			err := pubsub.Close()
			if err != nil {

			}
		}(pubsub)

		// Execute MONITOR command
		_, err := monitorClient.Do(ctx, "MONITOR").Result()
		if err != nil {
			log.Printf("Failed to start monitoring: %v", err)
			return
		}

		// Read messages
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Use a simple approach - execute INFO command periodically
				// This is a simplified version since direct MONITOR access is complex
				info, err := r.client.Info(ctx, "commandstats").Result()
				if err == nil {
					stringCh <- info
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	return stringCh, nil
}

// GetCommandStats returns Redis command statistics (alternative to Monitor)
func (r *RedisManager) GetCommandStats() (map[string]interface{}, error) {
	if r.client == nil {
		return nil, fmt.Errorf("no active connection")
	}

	info, err := r.client.Info(r.ctx, "commandstats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get command stats: %w", err)
	}

	stats := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "cmdstat_") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cmdName := strings.TrimPrefix(parts[0], "cmdstat_")
				cmdInfo := parts[1]

				// Parse command info: calls=X,usec=Y,usec_per_call=Z
				infoMap := make(map[string]string)
				infoParts := strings.Split(cmdInfo, ",")
				for _, part := range infoParts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 {
						infoMap[kv[0]] = kv[1]
					}
				}
				stats[cmdName] = infoMap
			}
		}
	}

	return stats, nil
}
