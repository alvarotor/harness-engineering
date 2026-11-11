package check

import (
	"fmt"
	"path/filepath"
	"regexp"
)

// Registry mirrors docs/decisions/registry.yaml.
type Registry struct {
	Version           int          `yaml:"version"`
	Decisions         []Decision   `yaml:"decisions"`
	Preferences       []Preference `yaml:"preferences"`
	Assumptions       []Entry      `yaml:"assumptions"`
	OpenQuestions     []Entry      `yaml:"open_questions"`
	ResolvedQuestions []Entry      `yaml:"resolved_questions"`
	Changelog         []Changelog  `yaml:"changelog"`
}

// Decision is a D-NNNN entry.
type Decision struct {
	ID                string      `yaml:"id"`
	Title             string      `yaml:"title"`
	Status            string      `yaml:"status"`
	Source            string      `yaml:"source"`
	ADR               string      `yaml:"adr"`
	Scope             string      `yaml:"scope"`
	DefaultLanguage   string      `yaml:"default_language"`
	Exceptions        []Exception `yaml:"exceptions"`
	EnforcedBy        []string    `yaml:"enforced_by"`
	PlannedEnforcedBy []string    `yaml:"planned_enforced_by"`
	SupersededBy      string      `yaml:"superseded_by"`
}

// Exception is a scoped override of a decision.
type Exception struct {
	Path     string `yaml:"path"`
	Language string `yaml:"language"`
	Reason   string `yaml:"reason"`
}

// Preference is a P-NNNN entry. Status and superseded_by are optional until a preference changes.
type Preference struct {
	ID                string   `yaml:"id"`
	Statement         string   `yaml:"statement"`
	Source            string   `yaml:"source"`
	Rationale         string   `yaml:"rationale"`
	EnforcedBy        []string `yaml:"enforced_by"`
	PlannedEnforcedBy []string `yaml:"planned_enforced_by"`
	Status            string   `yaml:"status"`
	SupersededBy      string   `yaml:"superseded_by"`
}

// Entry is an assumption or question.
type Entry struct {
	ID         string `yaml:"id"`
	Statement  string `yaml:"statement"`
	Source     string `yaml:"source"`
	Question   string `yaml:"question"`
	NeededBy   string `yaml:"needed_by"`
	Resolution string `yaml:"resolution"`
}

// Changelog records corrections so old choices are not reintroduced.
type Changelog struct {
	Date    string   `yaml:"date"`
	Changes []string `yaml:"changes"`
}

var (
	validStatus = map[string]bool{"proposed": true, "accepted": true, "superseded": true, "deprecated": true}
	validSource = map[string]bool{"user": true, "default": true}
	adrStatusRe = regexp.MustCompile(`(?m)^- Status:\s*(\S+)`)
	// | [0001](0001-x.md) | Title | accepted | user |
	adrIndexRe = regexp.MustCompile(`(?m)^\|\s*\[\d{4}\]\(([^)]+)\)\s*\|[^|]*\|\s*(\w+)\s*\|`)
)

const regFix = "edit " + registryPath + " (see the header comment for the schema)"

func checkRegistrySchema(r *Repo) []Violation {
	if r.RegistryErr != nil {
		return []Violation{{"R-REG-001", registryPath, "cannot load registry: " + r.RegistryErr.Error(), regFix}}
	}
	var out []Violation
	bad := func(id, problem string) {
		out = append(out, Violation{"R-REG-001", registryPath, fmt.Sprintf("%s: %s", id, problem), regFix})
	}
	seen := map[string]bool{}
	checkID := func(id string, re *regexp.Regexp) {
		if !re.MatchString(id) {
			bad(id, "malformed id, want "+re.String())
		}
		if seen[id] {
			bad(id, "duplicate id")
		}
		seen[id] = true
	}
	for _, d := range r.Registry.Decisions {
		checkID(d.ID, regexp.MustCompile(`^D-\d{4}$`))
		if !validStatus[d.Status] {
			bad(d.ID, fmt.Sprintf("invalid status %q", d.Status))
		}
		if !validSource[d.Source] {
			bad(d.ID, fmt.Sprintf("invalid source %q, want user or default", d.Source))
		}
		if d.Title == "" || d.ADR == "" {
			bad(d.ID, "title and adr are required")
		}
	}
	for _, p := range r.Registry.Preferences {
		checkID(p.ID, regexp.MustCompile(`^P-\d{4}$`))
		if !validSource[p.Source] {
			bad(p.ID, fmt.Sprintf("invalid source %q, want user or default", p.Source))
		}
		if p.Status != "" && !validStatus[p.Status] {
			bad(p.ID, fmt.Sprintf("invalid status %q", p.Status))
		}
	}
	for _, a := range r.Registry.Assumptions {
		checkID(a.ID, regexp.MustCompile(`^A\d+$`))
	}
	for _, q := range append(append([]Entry{}, r.Registry.OpenQuestions...), r.Registry.ResolvedQuestions...) {
		checkID(q.ID, regexp.MustCompile(`^Q\d+$`))
	}
	return out
}

func adrStatus(r *Repo, rel string) (string, bool) {
	data, err := r.read(rel)
	if err != nil {
		return "", false
	}
	m := adrStatusRe.FindSubmatch(data)
	if m == nil {
		return "", true
	}
	return string(m[1]), true
}

func checkADRStatus(r *Repo) []Violation {
	if r.Registry == nil {
		return nil
	}
	var out []Violation
	for _, d := range r.Registry.Decisions {
		status, ok := adrStatus(r, d.ADR)
		switch {
		case !ok:
			out = append(out, Violation{"R-REG-002", d.ADR, d.ID + " references a missing ADR", "create the ADR or fix the adr path in " + registryPath})
		case status != d.Status:
			out = append(out, Violation{"R-REG-002", d.ADR, fmt.Sprintf("ADR status %q differs from registry status %q for %s", status, d.Status, d.ID), "make the ADR '- Status:' line and the registry agree"})
		}
	}
	return out
}

func checkADRIndex(r *Repo) []Violation {
	if r.Registry == nil {
		return nil
	}
	files, _ := filepath.Glob(r.abs("docs/adr/[0-9][0-9][0-9][0-9]-*.md"))
	referenced := map[string]bool{}
	for _, d := range r.Registry.Decisions {
		referenced[filepath.ToSlash(filepath.Clean(d.ADR))] = true
	}
	index, _ := r.read("docs/adr/README.md")
	indexed := map[string]string{}
	for _, m := range adrIndexRe.FindAllSubmatch(index, -1) {
		indexed[string(m[1])] = string(m[2])
	}
	var out []Violation
	for _, f := range files {
		base := filepath.Base(f)
		rel := "docs/adr/" + base
		if !referenced[rel] {
			out = append(out, Violation{"R-REG-003", rel, "ADR has no registry decision", "add a decision with adr: " + rel + " to " + registryPath})
		}
		status, _ := adrStatus(r, rel)
		switch got, ok := indexed[base]; {
		case !ok:
			out = append(out, Violation{"R-REG-003", rel, "ADR missing from docs/adr/README.md", "add a row to the ADR index table"})
		case got != status:
			out = append(out, Violation{"R-REG-003", rel, fmt.Sprintf("index status %q differs from ADR status %q", got, status), "update the status column in docs/adr/README.md"})
		}
	}
	return out
}

func checkEnforcedBy(r *Repo) []Violation {
	if r.Registry == nil {
		return nil
	}
	var out []Violation
	for _, d := range r.Registry.Decisions {
		for _, p := range d.EnforcedBy {
			if !r.exists(p) {
				out = append(out, Violation{"R-REG-004", p, d.ID + " claims enforcement by a path that does not exist", "create it, or move it to planned_enforced_by"})
			}
		}
	}
	for _, pref := range r.Registry.Preferences {
		for _, p := range pref.EnforcedBy {
			if !r.exists(p) {
				out = append(out, Violation{"R-REG-004", p, pref.ID + " claims enforcement by a path that does not exist", "create it, or move the path to planned_enforced_by"})
			}
		}
	}
	return out
}

func checkSupersession(r *Repo) []Violation {
	if r.Registry == nil {
		return nil
	}
	decisionIDs := map[string]bool{}
	for _, d := range r.Registry.Decisions {
		decisionIDs[d.ID] = true
	}
	preferenceIDs := map[string]bool{}
	for _, p := range r.Registry.Preferences {
		preferenceIDs[p.ID] = true
	}
	var out []Violation
	check := func(id, status, supersededBy, targetType string, targetIDs map[string]bool) {
		switch {
		case status == "superseded" && supersededBy == "":
			out = append(out, Violation{"R-REG-005", registryPath, id + " is superseded but has no superseded_by", "name a replacement " + targetType})
		case status == "superseded" && !targetIDs[supersededBy]:
			out = append(out, Violation{"R-REG-005", registryPath, fmt.Sprintf("%s is superseded but superseded_by %q is not a known %s", id, supersededBy, targetType), "follow 'Changing an established decision' in CONTRIBUTING.md"})
		case status != "superseded" && supersededBy != "":
			out = append(out, Violation{"R-REG-005", registryPath, id + " has superseded_by but status " + status, "set status: superseded or remove superseded_by"})
		}
	}
	for _, d := range r.Registry.Decisions {
		check(d.ID, d.Status, d.SupersededBy, "decision", decisionIDs)
	}
	for _, p := range r.Registry.Preferences {
		check(p.ID, p.Status, p.SupersededBy, "preference", preferenceIDs)
	}
	return out
}

// scopeDecision returns the active decision governing a scope such as "services/*".
func scopeDecision(reg *Registry, scope string) *Decision {
	for i, d := range reg.Decisions {
		if d.Scope == scope && d.DefaultLanguage != "" && (d.Status == "accepted" || d.Status == "proposed") {
			return &reg.Decisions[i]
		}
	}
	return nil
}
