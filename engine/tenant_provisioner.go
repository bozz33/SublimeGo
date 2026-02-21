package engine

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bozz33/sublimego/internal/ent"
	_ "github.com/mattn/go-sqlite3"
)

// TenantProvisioner handles automatic database creation and migration for tenants.
type TenantProvisioner struct {
	// DBDir is the directory where tenant databases are stored (e.g. "./tenants")
	DBDir string
	// MigrationFunc is called after DB creation to run migrations/seeds
	MigrationFunc func(ctx context.Context, client *ent.Client) error
}

// NewTenantProvisioner creates a provisioner with default settings.
func NewTenantProvisioner(dbDir string) *TenantProvisioner {
	return &TenantProvisioner{
		DBDir: dbDir,
	}
}

// WithMigrations sets the migration function to run after DB creation.
func (p *TenantProvisioner) WithMigrations(fn func(ctx context.Context, client *ent.Client) error) *TenantProvisioner {
	p.MigrationFunc = fn
	return p
}

// ProvisionTenant creates a new database for the tenant if it doesn't exist.
// Returns the DSN to connect to the tenant's database.
func (p *TenantProvisioner) ProvisionTenant(ctx context.Context, tenant *Tenant) (string, error) {
	if err := os.MkdirAll(p.DBDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tenant DB directory: %w", err)
	}

	dbPath := filepath.Join(p.DBDir, fmt.Sprintf("%s.db", tenant.ID))
	dsn := fmt.Sprintf("file:%s?cache=shared&_fk=1", dbPath)

	// Check if DB already exists
	if _, err := os.Stat(dbPath); err == nil {
		// DB exists, return DSN
		return dsn, nil
	}

	// Create new database
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return "", fmt.Errorf("failed to ping tenant database: %w", err)
	}

	// Run migrations if provided
	if p.MigrationFunc != nil {
		client, err := ent.Open("sqlite3", dsn)
		if err != nil {
			return "", fmt.Errorf("failed to open ent client for migrations: %w", err)
		}
		defer client.Close()

		if err := p.MigrationFunc(ctx, client); err != nil {
			return "", fmt.Errorf("failed to run migrations for tenant %s: %w", tenant.ID, err)
		}
	}

	// Store DSN in tenant metadata
	if tenant.Meta == nil {
		tenant.Meta = make(map[string]any)
	}
	tenant.Meta["db_dsn"] = dsn

	return dsn, nil
}

// ProvisionAll provisions databases for all registered tenants.
func (p *TenantProvisioner) ProvisionAll(ctx context.Context, tenants []*Tenant) error {
	for _, t := range tenants {
		if _, err := p.ProvisionTenant(ctx, t); err != nil {
			return fmt.Errorf("failed to provision tenant %s: %w", t.ID, err)
		}
	}
	return nil
}

// DeleteTenant removes the tenant's database file.
// WARNING: This is destructive and cannot be undone.
func (p *TenantProvisioner) DeleteTenant(tenant *Tenant) error {
	dbPath := filepath.Join(p.DBDir, fmt.Sprintf("%s.db", tenant.ID))
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete tenant database: %w", err)
	}
	return nil
}

// GetTenantClient returns an ent.Client connected to the tenant's database.
func (p *TenantProvisioner) GetTenantClient(tenant *Tenant) (*ent.Client, error) {
	dsn, ok := tenant.Meta["db_dsn"].(string)
	if !ok || dsn == "" {
		return nil, fmt.Errorf("tenant %s has no database DSN", tenant.ID)
	}

	client, err := ent.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open tenant database: %w", err)
	}

	return client, nil
}
