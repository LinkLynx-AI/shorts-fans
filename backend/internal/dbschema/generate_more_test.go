package dbschema

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateValidatesInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		want string
	}{
		{
			name: "missing dsn",
			opts: Options{},
			want: "POSTGRES_DSN is required",
		},
		{
			name: "production is rejected",
			opts: Options{
				AppEnv:      "production",
				PostgresDSN: "postgres://example",
			},
			want: "schema export is disabled in production",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Generate(context.Background(), tt.opts)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Generate() error got %v want substring %q", err, tt.want)
			}
		})
	}
}

func TestListMigrationFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for name := range map[string]struct{}{
		"000002_second.up.sql":  {},
		"000001_first.up.sql":   {},
		"000001_first.down.sql": {},
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("-- sql"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v, want nil", err)
		}
	}

	got, err := listMigrationFiles(dir)
	if err != nil {
		t.Fatalf("listMigrationFiles() error = %v, want nil", err)
	}
	want := []string{"000001_first.up.sql", "000002_second.up.sql"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("listMigrationFiles() got %#v want %#v", got, want)
	}
}

func TestListMigrationFilesRequiresUpMigrations(t *testing.T) {
	t.Parallel()

	_, err := listMigrationFiles(t.TempDir())
	if err == nil {
		t.Fatal("listMigrationFiles() error = nil, want error")
	}
}

func TestNewTempDatabaseName(t *testing.T) {
	t.Parallel()

	got, err := newTempDatabaseName("schema_export")
	if err != nil {
		t.Fatalf("newTempDatabaseName() error = %v, want nil", err)
	}
	if !strings.HasPrefix(got, "schema_export_") {
		t.Fatalf("newTempDatabaseName() got %q want prefix %q", got, "schema_export_")
	}
	if len(got) != len("schema_export_")+8 {
		t.Fatalf("newTempDatabaseName() got length %d want %d", len(got), len("schema_export_")+8)
	}
}

func TestUniqueStrings(t *testing.T) {
	t.Parallel()

	got := uniqueStrings("", " postgres ", "template1", "postgres", "template1", "  ")
	want := []string{"postgres", "template1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uniqueStrings() got %#v want %#v", got, want)
	}
}
