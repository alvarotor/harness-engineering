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
		{"R-REG-006", checkHistory},
		{"R-CAT-001", checkCatalogSchema},
		{"R-CAT-002", checkLanguages},
		{"R-CAT-003", checkCatalogDisk},
		{"R-CAT-004", checkDeployableFiles},
		{"R-CAT-005", checkDependsOn},
		{"R-COMPOSE-001", checkComposePorts},
		{"R-COMPOSE-002", checkDockerSocket},
		{"R-COMPOSE-003", checkComposeRoutes},
		{"R-COMPOSE-004", checkComposeImages},
		{"R-COMPOSE-005", checkHardening},
		{"R-IMG-001", checkDockerfiles},
		{"R-CI-001", checkActionPins},
		{"R-CI-002", checkWorkflowHardening},
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
	Registry    *Registry // active registry merged with the history file
	RegistryErr error
	History     *Registry // docs/decisions/history.yaml, nil when absent
	Catalog     *Catalog
	CatalogErr  error
}

const (
	registryPath   = "docs/decisions/registry.yaml"
	historyPath    = "docs/decisions/history.yaml"
	catalogPath    = "docs/catalog/services.yaml"
	validationPath = "docs/standards/validation.md"
)

// Load reads and parses the registry, its history and the catalog. Parse errors are kept for the schema rules to report.
func Load(root string) *Repo {
	r := &Repo{Root: root}
	r.Registry, r.RegistryErr = loadYAML[Registry](r, registryPath)
	if r.Registry != nil && r.exists(historyPath) {
		h, err := loadYAML[Registry](r, historyPath)
		if err != nil {
			r.Registry, r.RegistryErr = nil, fmt.Errorf("%s: %w", historyPath, err)
		} else {
			r.History = h
			r.Registry.Decisions = append(r.Registry.Decisions, h.Decisions...)
			r.Registry.Preferences = append(r.Registry.Preferences, h.Preferences...)
			r.Registry.Assumptions = append(r.Registry.Assumptions, h.Assumptions...)
			r.Registry.OpenQuestions = append(r.Registry.OpenQuestions, h.OpenQuestions...)
			r.Registry.ResolvedQuestions = append(r.Registry.ResolvedQuestions, h.ResolvedQuestions...)
			r.Registry.Changelog = append(r.Registry.Changelog, h.Changelog...)
		}
	}
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
