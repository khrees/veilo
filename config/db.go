// Package config contains the configuration for the application.
package config

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE"`
	URL      string `env:"DATABASE_URL"`
}

func (c *DBConfig) DSN() string {
	if strings.TrimSpace(c.URL) != "" {
		return c.URL
	}

	sslMode := strings.TrimSpace(c.SSLMode)
	if sslMode == "" {
		sslMode = "require"
	}

	u := &url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   c.DBName,
	}

	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return u.String()
}

func (c *DBConfig) Connect() (*gorm.DB, error) {
	if err := c.preflight(); err != nil {
		return nil, err
	}

	dsn := c.DSN()
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Info("database connected")
	return db, nil
}

func (c *DBConfig) preflight() error {
	if strings.TrimSpace(c.URL) != "" {
		return nil
	}

	address := net.JoinHostPort(c.Host, c.Port)
	dialer := &net.Dialer{Timeout: 5 * time.Second}

	tcpConn, err := dialer.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf(
			"database preflight failed before login at %s: %w. Common causes: VPN/proxy, corporate firewall, DNS issues, or Supabase network restrictions",
			address, err,
		)
	}
	_ = tcpConn.Close()

	tlsConn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
		ServerName: c.Host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return fmt.Errorf(
			"database TLS preflight failed at %s: %w. If this only happens on a VPN, the VPN exit IP or TLS interception is likely the cause",
			address, err,
		)
	}
	_ = tlsConn.Close()

	return nil
}
