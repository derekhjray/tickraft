// Copyright © 2026 Beijing Ruishuo Technology Co., Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Dual-licensed — see LICENSE for details.

package telemetry

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"

	"gorm.io/gorm"
)

// builtinTemplateFS embeds the built-in template JSON files from the
// templates/ directory so the binary is self-contained without external
// file dependencies at runtime.
//
//go:embed templates/*.json
var builtinTemplateFS embed.FS

// builtinTemplateFile is the on-disk JSON structure of a built-in template.
// The Config field is kept as raw JSON so it can be stored verbatim as a
// string in Template.Config without imposing a fixed schema.
type builtinTemplateFile struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Category     string          `json:"category"`
	ExecutorType string          `json:"executor_type"`
	Config       json.RawMessage `json:"config"`
}

// builtinTemplateNames lists the built-in template JSON files in a fixed
// order. The order determines the insertion sequence when seeding the
// database for the first time.
var builtinTemplateNames = []string{
	"icmp-ping.json",
	"http-homepage.json",
	"https-api.json",
	"tcp-database.json",
	"dns-resolution.json",
	"ssl-certificate.json",
	"redis-connect.json",
	"mysql-connect.json",
}

// readBuiltinTemplates reads and parses the embedded built-in template JSON
// files, returning them sorted by the predefined name order. A malformed
// file causes the function to return an error so startup fails loudly
// rather than silently seeding a partial template set.
func readBuiltinTemplates() ([]builtinTemplateFile, error) {
	templates := make([]builtinTemplateFile, 0, len(builtinTemplateNames))
	for _, name := range builtinTemplateNames {
		data, err := builtinTemplateFS.ReadFile("templates/" + name)
		if err != nil {
			return nil, fmt.Errorf("telemetry: read builtin template %q: %w", name, err)
		}
		var t builtinTemplateFile
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("telemetry: parse builtin template %q: %w", name, err)
		}
		templates = append(templates, t)
	}
	return templates, nil
}

// LoadBuiltinTemplates loads built-in templates into the database if they
// do not already exist. It is idempotent: templates that already exist (by
// name) are skipped, so it is safe to call on every startup. The function
// also runs AutoMigrate for the template table to ensure the schema exists
// before inserting.
func LoadBuiltinTemplates(dbc *gorm.DB) error {
	if dbc == nil {
		return fmt.Errorf("telemetry: load builtin templates: db is nil")
	}

	if err := dbc.AutoMigrate(&Template{}); err != nil {
		return fmt.Errorf("telemetry: migrate template table: %w", err)
	}

	templates, err := readBuiltinTemplates()
	if err != nil {
		return err
	}

	for _, t := range templates {
		// Check by name to avoid re-inserting on restarts.
		var count int64
		if err := dbc.Model(&Template{}).Where("name = ?", t.Name).Count(&count).Error; err != nil {
			return fmt.Errorf("telemetry: check builtin template %q: %w", t.Name, err)
		}
		if count > 0 {
			continue
		}

		model := Template{
			Name:         t.Name,
			Description:  t.Description,
			Category:     t.Category,
			ExecutorType: t.ExecutorType,
			Config:       string(t.Config),
			IsBuiltin:    true,
		}
		if err := dbc.Create(&model).Error; err != nil {
			return fmt.Errorf("telemetry: insert builtin template %q: %w", t.Name, err)
		}
	}

	return nil
}

// ListEmbeddedTemplateFiles returns the names of the embedded template JSON
// files, sorted alphabetically. It is primarily useful for diagnostics and
// tests.
func ListEmbeddedTemplateFiles() ([]string, error) {
	entries, err := fs.ReadDir(builtinTemplateFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("telemetry: read embedded template dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
