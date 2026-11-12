package check

import (
	"fmt"
	"os"
	"regexp"
)

// Catalog mirrors docs/catalog/services.yaml.
type Catalog struct {
	Version     int          `yaml:"version"`
	Deployables []Deployable `yaml:"deployables"`
}

// Deployable is one app or service.
type Deployable struct {
	Name         string    `yaml:"name"`
	Kind         string    `yaml:"kind"`
	Path         string    `yaml:"path"`
	Language     string    `yaml:"language"`
	Status       string    `yaml:"status"`
	Owner        string    `yaml:"owner"`
	Domain       string    `yaml:"domain"`
	Description  string    `yaml:"description"`
	Exposed      string    `yaml:"exposed"`
	Database     string    `yaml:"database"`
	PersonalData *[]string `yaml:"personal_data"`
	DependsOn    []string  `yaml:"depends_on"`
	Decisions    []string  `yaml:"decisions"`
}

var (
	nameRe        = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	kindDir       = map[string]string{"app": "apps", "service": "services"}
	catalogStatus = map[string]bool{"planned": true, "active": true, "deprecated": true}
	manifest      = map[string]string{"go": "go.mod", "typescript": "package.json"}
)

const catFix = "edit " + catalogPath + " (see docs/catalog/README.md)"

func checkCatalogSchema(r *Repo) []Violation {
	if r.CatalogErr != nil {
		return []Violation{{"R-CAT-001", catalogPath, "cannot load catalog: " + r.CatalogErr.Error(), catFix}}
	}
	var out []Violation
	bad := func(name, problem string) {
		out = append(out, Violation{"R-CAT-001", catalogPath, fmt.Sprintf("%s: %s", name, problem), catFix})
	}
	seen := map[string]bool{}
	for _, d := range r.Catalog.Deployables {
		if !nameRe.MatchString(d.Name) {
			bad(d.Name, "name must be kebab-case")
		}
		if seen[d.Name] {
			bad(d.Name, "duplicate name")
		}
		seen[d.Name] = true
		if dir, ok := kindDir[d.Kind]; !ok {
			bad(d.Name, fmt.Sprintf("invalid kind %q, want app or service", d.Kind))
		} else if d.Path != dir+"/"+d.Name {
			bad(d.Name, fmt.Sprintf("path %q, want %s/%s", d.Path, dir, d.Name))
		}
		if !catalogStatus[d.Status] {
			bad(d.Name, fmt.Sprintf("invalid status %q", d.Status))
		}
		if d.Owner == "" || d.Language == "" {
			bad(d.Name, "owner and language are required")
		}
		if d.Kind == "service" && d.Domain == "" {
			bad(d.Name, "services must name their business domain")
		}
		if d.PersonalData == nil {
			bad(d.Name, "personal_data is required (use [] if none)")
		}
	}
	return out
}

// builtinLanguage is the default language per scope when no registry decision with that scope exists.
var builtinLanguage = map[string]string{"services/*": "go", "apps/*": "typescript"}

func checkLanguages(r *Repo) []Violation {
	if r.Catalog == nil || r.Registry == nil {
		return nil
	}
	var out []Violation
	for _, d := range r.Catalog.Deployables {
		scope := kindDir[d.Kind] + "/*"
		def, source, fix, exceptions := builtinLanguage[scope], "the built-in default", "record a decision with scope "+scope+" and an exception (path, language, reason)", []Exception(nil)
		if dec := scopeDecision(r.Registry, scope); dec != nil {
			def, source, fix, exceptions = dec.DefaultLanguage, dec.ID, "add an exception (path, language, reason) to "+dec.ID, dec.Exceptions
		}
		if def == "" || d.Language == def {
			continue
		}
		allowed := false
		for _, e := range exceptions {
			allowed = allowed || (e.Path == d.Path && e.Language == d.Language)
		}
		if !allowed {
			out = append(out, Violation{"R-CAT-002", d.Path,
				fmt.Sprintf("language %q is not the default %q of %s and has no exception", d.Language, def, source),
				fmt.Sprintf("use %s, or %s through an approved change", def, fix)})
		}
	}
	return out
}

func checkCatalogDisk(r *Repo) []Violation {
	if r.Catalog == nil {
		return nil
	}
	var out []Violation
	known := map[string]bool{}
	for _, d := range r.Catalog.Deployables {
		known[d.Path] = true
		onDisk := r.exists(d.Path)
		switch {
		case d.Status == "planned" && onDisk:
			out = append(out, Violation{"R-CAT-003", d.Path, "code exists but catalog status is planned", "set status: active in " + catalogPath})
		case d.Status != "planned" && !onDisk:
			out = append(out, Violation{"R-CAT-003", d.Path, "catalog status is " + d.Status + " but the folder is missing", "add the code, or set status: planned"})
		}
	}
	for _, dir := range []string{"apps", "services"} {
		entries, _ := os.ReadDir(r.abs(dir))
		for _, e := range entries {
			if p := dir + "/" + e.Name(); e.IsDir() && !known[p] {
				out = append(out, Violation{"R-CAT-003", p, "deployable folder has no catalog entry", "add it to " + catalogPath + " or remove the folder"})
			}
		}
	}
	return out
}

func checkDeployableFiles(r *Repo) []Violation {
	if r.Catalog == nil {
		return nil
	}
	var out []Violation
	for _, d := range r.Catalog.Deployables {
		if d.Status != "active" {
			continue
		}
		required := []string{"Dockerfile", "AGENTS.md", "README.md", "compose.yaml"}
		if m, ok := manifest[d.Language]; ok {
			required = append(required, m)
		}
		for _, f := range required {
			if p := d.Path + "/" + f; !r.exists(p) {
				out = append(out, Violation{"R-CAT-004", p, "required file missing for an active deployable", "add it (service templates generate it)"})
			}
		}
	}
	return out
}

func checkDependsOn(r *Repo) []Violation {
	if r.Catalog == nil {
		return nil
	}
	names := map[string]bool{}
	for _, d := range r.Catalog.Deployables {
		names[d.Name] = true
	}
	var out []Violation
	for _, d := range r.Catalog.Deployables {
		for _, dep := range d.DependsOn {
			if !names[dep] || dep == d.Name {
				out = append(out, Violation{"R-CAT-005", catalogPath, fmt.Sprintf("%s depends on unknown or self %q", d.Name, dep), "fix depends_on or add the catalog entry"})
			}
		}
	}
	return out
}
