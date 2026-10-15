// Package check implements the repository validation rules documented in
// docs/standards/validation.md.
package check

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Violation is one failed check with a remedy.
type Violation struct {
	Rule    string
	Path    string
	Problem string
	Fix     string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s  %s: %s\n    fix: %s", v.Rule, v.Path, v.Problem, v.Fix)
}

// Rule is a single validation rule.
type Rule struct {
	ID    string
	Check func(*Repo) []Violation
}

// Rules returns every rule, in execution order. Each ID must be documented in docs/standards/validation.md.
func Rules() []Rule {
	return []Rule{
		{"R-STRUCT-001", checkRequiredFiles},
		{"R-REG-001", checkRegistrySchema},
		{"R-REG-002", checkADRStatus},
		{"R-REG-003", checkADRIndex},
		{"R-REG-004", checkEnforcedBy},
		{"R-REG-005", checkSupersession},
		{"R-CAT-001", checkCatalogSchema},
		{"R-CAT-002", checkLanguages},
		{"R-CAT-003", checkCatalogDisk},
		{"R-CAT-004", checkDeployableFiles},
		{"R-CAT-005", checkDependsOn},
		{"R-AGENT-001", checkRoleSections},
		{"R-AGENT-002", checkRoleAdapters},
		{"R-SCOPE-001", checkScopedAgents},
		{"R-DOC-001", checkLinks},
		{"R-DOC-002", checkRulesDocumented},
		{"R-DOC-003", checkArchitectureMentions},
	}
}

// Repo is the loaded repository under validation.
type Repo struct {
	Root        string
	Registry    *Registry
	RegistryErr error
	Catalog     *Catalog
	CatalogErr  error
}

const (
	registryPath   = "docs/decisions/registry.yaml"
	catalogPath    = "docs/catalog/services.yaml"
	validationPath = "docs/standards/validation.md"
)

// Load reads and parses the registry and catalog. Parse errors are kept for the schema rules to report.
func Load(root string) *Repo {
	r := &Repo{Root: root}
	r.Registry, r.RegistryErr = loadYAML[Registry](r, registryPath)
	r.Catalog, r.CatalogErr = loadYAML[Catalog](r, catalogPath)
	return r
}

// Run executes all rules and returns the number of rules and the sorted violations.
func Run(root string) (int, []Violation) {
	r := Load(root)
	rules := Rules()
	var out []Violation
	for _, rule := range rules {
		out = append(out, rule.Check(r)...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Rule != out[j].Rule {
			return out[i].Rule < out[j].Rule
		}
		return out[i].Path < out[j].Path
	})
	return len(rules), out
}

func loadYAML[T any](r *Repo, rel string) (*T, error) {
	data, err := r.read(rel)
	if err != nil {
		return nil, err
	}
	var v T
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true) // unknown keys are typos, not extensions
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *Repo) abs(rel string) string { return filepath.Join(r.Root, filepath.FromSlash(rel)) }

func (r *Repo) read(rel string) ([]byte, error) { return os.ReadFile(r.abs(rel)) }

func (r *Repo) exists(rel string) bool {
	_, err := os.Stat(r.abs(rel))
	return !errors.Is(err, fs.ErrNotExist)
}
