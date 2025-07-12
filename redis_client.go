package grc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SimpleRedisClient is a simple Redis client implementation without external dependencies
type SimpleRedisClient struct {
	addr     string
	password string
	db       int
	conn     net.Conn
	mu       sync.Mutex
}

// SimpleRedisConfig contains configuration for the simple Redis client
type SimpleRedisConfig struct {
	Addr     string // Redis server address (e.g., "localhost:6379")
	Password string // Redis password (optional)
	DB       int    // Redis database number (default: 0)
}

// NewSimpleRedisClient creates a new simple Redis client
func NewSimpleRedisClient(config SimpleRedisConfig) (*SimpleRedisClient, error) {
	client := &SimpleRedisClient{
		addr:     config.Addr,
		password: config.Password,
		db:       config.DB,
	}

	err := client.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// connect establishes a connection to Redis
func (r *SimpleRedisClient) connect() error {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return err
	}

	r.conn = conn

	// Authenticate if password is provided
	if r.password != "" {
		_, err = r.sendCommand("AUTH", r.password)
		if err != nil {
			r.conn.Close()
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Select database if not default
	if r.db != 0 {
		_, err = r.sendCommand("SELECT", strconv.Itoa(r.db))
		if err != nil {
			r.conn.Close()
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	return nil
}

// sendCommand sends a command to Redis and returns the response
func (r *SimpleRedisClient) sendCommand(cmd string, args ...string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Build Redis protocol command
	cmdArgs := []string{cmd}
	cmdArgs = append(cmdArgs, args...)

	// Format as Redis protocol
	command := fmt.Sprintf("*%d\r\n", len(cmdArgs))
	for _, arg := range cmdArgs {
		command += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}

	// Send command
	_, err := r.conn.Write([]byte(command))
	if err != nil {
		return "", err
	}

	// Read response
	reader := bufio.NewReader(r.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	response = strings.TrimSpace(response)

	// Handle different response types
	switch response[0] {
	case '+': // Simple string
		return response[1:], nil
	case '-': // Error
		return "", fmt.Errorf("Redis error: %s", response[1:])
	case ':': // Integer
		return response[1:], nil
	case '$': // Bulk string
		length, err := strconv.Atoi(response[1:])
		if err != nil {
			return "", err
		}
		if length == -1 {
			return "", ErrCacheMiss // NULL bulk string
		}
		
		data := make([]byte, length)
		_, err = reader.Read(data)
		if err != nil {
			return "", err
		}
		
		// Read the trailing \r\n
		reader.ReadString('\n')
		
		return string(data), nil
	default:
		return "", fmt.Errorf("unknown response type: %c", response[0])
	}
}

// Get retrieves a value from Redis
func (r *SimpleRedisClient) Get(ctx context.Context, key string) (interface{}, error) {
	value, err := r.sendCommand("GET", key)
	if err != nil {
		return nil, err
	}

	return []byte(value), nil
}

// Set stores a value in Redis with TTL
func (r *SimpleRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if ttl > 0 {
		ttlSeconds := int(ttl.Seconds())
		_, err = r.sendCommand("SETEX", key, strconv.Itoa(ttlSeconds), string(data))
	} else {
		_, err = r.sendCommand("SET", key, string(data))
	}

	return err
}

// Close closes the Redis connection
func (r *SimpleRedisClient) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}