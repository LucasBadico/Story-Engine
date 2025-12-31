package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/config"
	_ "github.com/lib/pq"
)

const (
	pgIdentMax = 63
)

// DatabaseManager handles database operations for tests
type DatabaseManager struct {
	db         *sql.DB
	connString string
	port       int
	mu         sync.Mutex
	user       string
	password   string
	host       string
	dbname     string
}

var reSanitize = regexp.MustCompile(`[^a-z0-9_]+`)

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(cfg config.DatabaseConfig) (*DatabaseManager, error) {
	// Connect to the postgres database for admin operations
	adminDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)

	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Verify the connection works
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &DatabaseManager{
		db:       db,
		port:     cfg.Port,
		user:     cfg.User,
		password: cfg.Password,
		host:     cfg.Host,
		dbname:   cfg.Database,
	}, nil
}

// MakeDBName generates a valid and unique PostgreSQL database name (<=63 chars).
// Format: <prefix>_<YYMMDDhhmmss>_<uuid8>[_<slug>]
func MakeDBName(prefix, testName string) string {
	if prefix == "" {
		prefix = "test"
	}

	// Unique core part: prefix + UTC timestamp + short UUID
	ts := time.Now().UTC().Format("060102150405")           // YYMMDDhhmmss (12 chars)
	u8 := strings.ReplaceAll(uuid.NewString(), "-", "")[:8] // 8-character hex

	core := fmt.Sprintf("%s_%s_%s", strings.ToLower(prefix), ts, u8)

	// Sanitize testName -> lowercase, only [a-z0-9_]
	slug := strings.ToLower(testName)
	slug = strings.ReplaceAll(slug, "-", "_")
	slug = reSanitize.ReplaceAllString(slug, "_")
	slug = strings.Trim(slug, "_")

	// Calculate remaining space for "_<slug>"
	remain := pgIdentMax - len(core) - 1

	if remain > 0 && len(slug) > 0 {
		if len(slug) > remain {
			slug = slug[:remain]
		}
		return fmt.Sprintf("%s_%s", core, slug)
	}

	if len(core) > pgIdentMax {
		return core[:pgIdentMax]
	}
	return core
}

// GetTemplateDBName returns the name of the template database
func GetTemplateDBName(cfg config.DatabaseConfig) string {
	// Use the main database as template (it already has migrations applied)
	return cfg.Database
}

// CreateTemplateDatabase creates a template database with the schema
func (m *DatabaseManager) CreateTemplateDatabase(templateName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Drop the template database if it exists
	_, err := m.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", safeIdent(templateName)))
	if err != nil {
		return fmt.Errorf("failed to drop template database: %w", err)
	}

	// Create the template database
	_, err = m.db.Exec(fmt.Sprintf("CREATE DATABASE %s", safeIdent(templateName)))
	if err != nil {
		return fmt.Errorf("failed to create template database: %w", err)
	}

	return nil
}

// ApplyMigrations applies golang-migrate migrations to the template database
func (m *DatabaseManager) ApplyMigrations(templateName, migrationsPath string) error {
	// Build connection string for template database
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		m.user, m.password, m.host, m.port, templateName)

	// Use golang-migrate CLI
	cmd := exec.Command("migrate",
		"-path", migrationsPath,
		"-database", dbURL,
		"up")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run migrations: %w\nStderr: %s", err, stderr.String())
	}

	return nil
}

// CloneDatabase creates a new test database by cloning the template
func (m *DatabaseManager) CloneDatabase(templateName, targetName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if target exists
	targetExists, err := m.VerifyDatabaseExists(targetName)
	if err != nil {
		return fmt.Errorf("failed to check if target exists: %w", err)
	}

	if targetExists {
		// Drop existing database
		_, _ = m.db.Exec(fmt.Sprintf(
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = %s",
			safeLiteral(targetName)))

		_, err = m.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", safeIdent(targetName)))
		if err != nil {
			return fmt.Errorf("failed to drop existing database: %w", err)
		}
	}

	// Clone from template using pg_dump for schema only
	if err := m.cloneSchemaWithPgDump(templateName, targetName); err != nil {
		return fmt.Errorf("failed to clone schema: %w", err)
	}

	return nil
}

// cloneSchemaWithPgDump runs: CREATE DATABASE + pg_dump --schema-only | psql
func (m *DatabaseManager) cloneSchemaWithPgDump(templateDB, targetDB string) error {
	// First create empty database
	_, err := m.db.Exec(fmt.Sprintf("CREATE DATABASE %s", safeIdent(targetDB)))
	if err != nil {
		return fmt.Errorf("failed to create empty database: %w", err)
	}

	// Build environment with PGPASSWORD
	env := os.Environ()
	if m.password != "" {
		env = append(env, "PGPASSWORD="+m.password)
	}

	pgDumpArgs := []string{
		"-h", m.host,
		"-p", fmt.Sprintf("%d", m.port),
		"-U", m.user,
		"--schema-only",
		"--no-owner",
		"--no-privileges",
		templateDB,
	}
	psqlArgs := []string{
		"-h", m.host,
		"-p", fmt.Sprintf("%d", m.port),
		"-U", m.user,
		targetDB,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pgDump := exec.CommandContext(ctx, "pg_dump", pgDumpArgs...)
	pgDump.Env = env

	psql := exec.CommandContext(ctx, "psql", psqlArgs...)
	psql.Env = env

	// Pipe pg_dump stdout -> psql stdin
	stdoutPipe, err := pgDump.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pg_dump StdoutPipe: %w", err)
	}
	defer stdoutPipe.Close()

	psql.Stdin = stdoutPipe

	var dumpErr, psqlErr bytes.Buffer
	pgDump.Stderr = &dumpErr
	psql.Stderr = &psqlErr

	if err := pgDump.Start(); err != nil {
		return fmt.Errorf("pg_dump start: %w (%s)", err, dumpErr.String())
	}
	if err := psql.Start(); err != nil {
		_ = pgDump.Process.Kill()
		return fmt.Errorf("psql start: %w (%s)", err, psqlErr.String())
	}

	if err := pgDump.Wait(); err != nil {
		_ = psql.Process.Kill()
		return fmt.Errorf("pg_dump wait: %w (%s)", err, dumpErr.String())
	}
	if err := psql.Wait(); err != nil {
		return fmt.Errorf("psql wait: %w (%s)", err, psqlErr.String())
	}

	return nil
}

// VerifyDatabaseExists checks if a database exists
func (m *DatabaseManager) VerifyDatabaseExists(dbName string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err := m.db.QueryRow(query, dbName).Scan(&exists)
	return exists, err
}

// GetConnectionString returns a connection string for the specified database
func (m *DatabaseManager) GetConnectionString(dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		m.user, m.password, m.host, m.port, dbName)
}

// Close closes the database connection
func (m *DatabaseManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// DropDatabase drops a database if it exists
func (m *DatabaseManager) DropDatabase(dbName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Terminate connections
	_, _ = m.db.Exec(fmt.Sprintf(
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = %s",
		safeLiteral(dbName)))

	// Drop database
	_, err := m.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", safeIdent(dbName)))
	return err
}

// safeIdent safely quotes an SQL identifier
func safeIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

// safeLiteral safely quotes a string literal for SQL
func safeLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

