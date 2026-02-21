package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bozz33/sublimego/engine"
)

func TestTenantProvisioner(t *testing.T) {
	// Use temp directory for test databases
	tmpDir := t.TempDir()
	provisioner := engine.NewTenantProvisioner(tmpDir)

	tenant := &engine.Tenant{
		ID:        "acme",
		Name:      "Acme Corp",
		Subdomain: "acme",
	}

	ctx := context.Background()

	// Test provisioning
	dsn, err := provisioner.ProvisionTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("ProvisionTenant failed: %v", err)
	}

	if dsn == "" {
		t.Error("expected non-empty DSN")
	}

	// Verify DB file was created
	dbPath := filepath.Join(tmpDir, "acme.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file not created at %s", dbPath)
	}

	// Test DSN stored in metadata
	if tenant.Meta["db_dsn"] != dsn {
		t.Errorf("DSN not stored in tenant metadata")
	}

	// Test re-provisioning returns same DSN
	dsn2, err := provisioner.ProvisionTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("Re-provisioning failed: %v", err)
	}
	if dsn != dsn2 {
		t.Errorf("expected same DSN on re-provision, got %s vs %s", dsn, dsn2)
	}
}

func TestTenantProvisionerDelete(t *testing.T) {
	tmpDir := t.TempDir()
	provisioner := engine.NewTenantProvisioner(tmpDir)

	tenant := &engine.Tenant{
		ID:        "test",
		Name:      "Test Tenant",
		Subdomain: "test",
	}

	ctx := context.Background()

	// Create tenant DB
	_, err := provisioner.ProvisionTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("ProvisionTenant failed: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file not created")
	}

	// Delete tenant DB
	if err := provisioner.DeleteTenant(tenant); err != nil {
		t.Fatalf("DeleteTenant failed: %v", err)
	}

	// Verify DB file was deleted
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Error("database file should have been deleted")
	}
}
