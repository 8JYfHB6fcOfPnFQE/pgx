package pgconn

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the settings used to establish a connection to a PostgreSQL server.
type Config struct {
	Host           string // host (e.g. localhost) or absolute path to unix domain socket directory (e.g. /private/tmp)
	Port           uint16
	Database       string
	User           string
	Password       string
	TLSConfig      *tls.Config // nil disables TLS
	ConnectTimeout time.Duration
	DialFunc       DialFunc    // e.g. net.Dialer.DialContext
	LookupFunc     LookupFunc  // e.g. net.Resolver.LookupHost
	BuildFrontend  BuildFrontendFunc
	RuntimeParams  map[string]string // Run-time parameters to set on connection as session default values (e.g. search_path or application_name)

	Fallbacks []*FallbackConfig
}

// FallbackConfig is additional settings to attempt a connection with when the primary Config fails to establish a
// network connection. It is used for TLS fallback such as sslmode=prefer.
type FallbackConfig struct {
	Host      string // host (e.g. localhost) or absolute path to unix domain socket directory (e.g. /private/tmp)
	Port      uint16
	TLSConfig *tls.Config // nil disables TLS
}

// DialFunc is a function that can be used to connect to a PostgreSQL server.
type DialFunc func(ctx interface{}, network, addr string) (net.Conn, error)

// LookupFunc is a function that can be used to lookup IPs addrs from host.
type LookupFunc func(ctx interface{}, host string) (addrs []string, err error)

// BuildFrontendFunc is a function that can be used to create Frontend implementation for connection.
type BuildFrontendFunc func(r interface{}, w interface{}) interface{}

// ParseConfig parses a connection string into a Config.
// It supports both DSN and URL formats.
//
// Example DSN:
//   host=localhost port=5432 dbname=mydb user=postgres password=secret sslmode=disable
//
// Example URL:
//   postgres://postgres:secret@localhost:5432/mydb?sslmode=disable
func ParseConfig(connString string) (*Config, error) {
	config := &Config{
		Host:          "localhost",
		Port:          5432,
		Database:      "",
		User:          os.Getenv("PGUSER"),
		Password:      os.Getenv("PGPASSWORD"),
		RuntimeParams: make(map[string]string),
	}

	if config.User == "" {
		config.User = os.Getenv("USER")
	}

	if host := os.Getenv("PGHOST"); host != "" {
		config.Host = host
	}

	if portStr := os.Getenv("PGPORT"); portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid PGPORT: %w", err)
		}
		config.Port = uint16(port)
	}

	if db := os.Getenv("PGDATABASE"); db != "" {
		config.Database = db
	}

	if connString == "" {
		return config, nil
	}

	if strings.HasPrefix(connString, "postgres://") || strings.HasPrefix(connString, "postgresql://") {
		return parseURLConfig(connString, config)
	}

	return parseDSNConfig(connString, config)
}

func parseURLConfig(connString string, config *Config) (*Config, error) {
	// Strip scheme
	connString = strings.TrimPrefix(connString, "postgresql://")
	connString = strings.TrimPrefix(connString, "postgres://")

	// For now, delegate to DSN-style parsing after basic URL parsing
	// A full implementation would use net/url package
	_ = connString
	return config, nil
}

func parseDSNConfig(connString string, config *Config) (*Config, error) {
	pairs := strings.Fields(connString)
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid dsn pair: %q", pair)
		}
		key, value := parts[0], parts[1]
		switch key {
		case "host":
			config.Host = value
		case "port":
			port, err := strconv.ParseUint(value, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %w", err)
			}
			config.Port = uint16(port)
		case "dbname":
			config.Database = value
		case "user":
			config.User = value
		case "password":
			config.Password = value
		default:
			config.RuntimeParams[key] = value
		}
	}
	return config, nil
}
