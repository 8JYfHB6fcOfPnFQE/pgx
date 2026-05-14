// Package pgconn is a low-level PostgreSQL database driver.
//
// It provides a low-level interface to PostgreSQL that is suitable for
// implementing higher-level drivers such as the standard database/sql package.
package pgconn

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
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
	DialFunc       DialFunc // e.g. net.Dialer.DialContext
	LookupFunc     LookupFunc
	BuildFrontend  BuildFrontendFunc
	RuntimeParams  map[string]string // Run-time parameters to set on connection as session default values (e.g. search_path or application_name)

	FallbackConfigs []*FallbackConfig

	ValidateConnect ValidateConnectFunc
	AfterConnect    AfterConnectFunc

	OnNotice        NoticeHandler
	OnNotification  NotificationHandler
	OnPgError       PgErrorHandler
}

// DefaultPort is the default PostgreSQL port. Defined here for easy reference
// when constructing Config values without going through the DSN parser.
const DefaultPort uint16 = 5432

// FallbackConfig is used to attempt a connection with a different set of connection parameters if the primary attempt
// fails. It is used for TLS fallback such as sslmode=prefer.
type FallbackConfig struct {
	Host      string
	Port      uint16
	TLSConfig *tls.Config
}

// PgConn is a low-level PostgreSQL connection handle. It is not safe for concurrent usage.
type PgConn struct {
	conn              net.Conn
	pid               uint32 // backend pid
	secretKey         uint32 // key to use to send a cancel query message to the server
	parameterStatuses map[string]string // parameters that have been reported by the server
	txStatus          byte
	config            *Config

	closed bool
}

// DialFunc is a function that can be used to connect to a PostgreSQL server.
type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// LookupFunc is a function that can be used to look up the IP addresses for a given host.
type LookupFunc func(ctx context.Context, host string) (addrs []string, err error)

// BuildFrontendFunc is a function that can be used to create a Frontend.
type BuildFrontendFunc func(r interface{}, w interface{}) interface{}

// ValidateConnectFunc is called at the end of a successful connection. It can be used to validate that the connection
// is usable (e.g. the correct PostgreSQL version is running).
type ValidateConnectFunc func(ctx context.Context, pgconn *PgConn) error

// AfterConnectFunc is called after a connection is established and validated. It can be used to set session-level
// state such as search_path.
type AfterConnectFunc func(ctx context.Context, pgconn *PgConn) error

// NoticeHandler is a function that can handle a notice response.
type NoticeHandler func(c *PgConn, n *Notice)

// NotificationHandler is a function that can handle a notification from the LISTEN/NOTI
