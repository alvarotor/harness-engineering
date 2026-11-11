package check

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var requiredFiles = []string{
	"AGENTS.md", "ARCHITECTURE.md", "CONTRIBUTING.md", "Makefile",
	registryPath, catalogPath, "docs/adr/README.md", "docs/agents/README.md",
	"docs/standards/engineering.md", "docs/standards/security.md", "docs/standards/testing.md",
	"docs/standards/definition-of-done.md", validationPath, "openspec/config.yaml", bootstrapPath,
}

const bootstrapPath = "scripts/bootstrap.sh"

// Roles is the fixed role set (spec agent-roles). Adding one is a decision change.
var Roles = []string{"orchestrator", "architect", "planner", "implementer", "test-engineer", "security-reviewer", "reviewer", "documentation", "devops"}

var roleSections = []string{
	"Purpose", "Responsibilities", "Allowed actions", "Required inputs", "Expected outputs",
	"Validation", "Handoff", "Stop conditions", "Human approval",
}

var scopedAgents = []string{"apps/AGENTS.md", "services/AGENTS.md", "contracts/AGENTS.md", "infra/AGENTS.md"}

// Directories whose Markdown is generated, vendored or frozen.
var skipDirs = []string{".git", "node_modules", "vendor", "dist", "coverage", "openspec/changes/archive", ".opencode/skills", ".opencode/commands"}

var (
	linkRe    = regexp.MustCompile(`\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	ruleRowRe = regexp.MustCompile(`(?m)^\|\s*(R-[A-Z]+-\d{3})\s*\|`)
)

func checkRequiredFiles(r *Repo) []Violation {
	var out []Violation
	for _, f := range requiredFiles {
		if !r.exists(f) {
			out = append(out, Violation{"R-STRUCT-001", f, "required harness file is missing", "restore it; removing harness files needs a decision change"})
		}
	}
	if info, err := os.Stat(r.abs(bootstrapPath)); err == nil && info.Mode()&0o111 == 0 {
		out = append(out, Violation{"R-STRUCT-001", bootstrapPath, "bootstrap script is not executable, so a fresh clone cannot run it", "chmod +x " + bootstrapPath + " (git keeps the bit)"})
	}
	return out
}

func checkRoleSections(r *Repo) []Violation {
	var out []Violation
	known := map[string]bool{"README.md": true}
	for _, role := range Roles {
		rel := "docs/agents/" + role + ".md"
		known[role+".md"] = true
		data, err := r.read(rel)
		if err != nil {
			out = append(out, Violation{"R-AGENT-001", rel, "role definition missing", "write it with all required sections (see docs/agents/README.md)"})
			continue
		}
		for _, s := range roleSections {
			if !regexp.MustCompile(`(?m)^## ` + regexp.QuoteMeta(s) + `\s*$`).Match(data) {
				out = append(out, Violation{"R-AGENT-001", rel, "missing section '## " + s + "'", "add the section"})
			}
		}
	}
	entries, _ := os.ReadDir(r.abs("docs/agents"))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !known[e.Name()] {
			out = append(out, Violation{"R-AGENT-001", "docs/agents/" + e.Name(), "unknown role; the role set is fixed", "merge it into an existing role, or add it through an architect decision change (and update tools/harness Roles)"})
		}
	}
	return out
}

func checkRoleAdapters(r *Repo) []Violation {
	var out []Violation
	for _, role := range Roles {
		rel := ".opencode/agents/" + role + ".md"
		data, err := r.read(rel)
		if err != nil || !bytes.Contains(data, []byte("docs/agents/"+role+".md")) {
			out = append(out, Violation{"R-AGENT-002", rel, "OpenCode adapter missing or not pointing to docs/agents/" + role + ".md", "create the adapter; its body must reference the canonical role file"})
		}
	}
	return out
}

func checkScopedAgents(r *Repo) []Violation {
	var out []Violation
	for _, f := range scopedAgents {
		if !r.exists(f) {
			out = append(out, Violation{"R-SCOPE-001", f, "scoped agent instructions missing", "restore the file"})
		}
	}
	return out
}

func checkLinks(r *Repo) []Violation {
	var out []Violation
	_ = filepath.WalkDir(r.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(r.Root, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			for _, s := range skipDirs {
				if rel == s {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(rel, ".md") {
			return nil
		}
		data, err := os.ReadFile(p) //nolint:gosec // reading repo files by walked path is the purpose of this rule
		if err != nil {
			return nil
		}
		for _, target := range markdownLinks(data) {
			resolved := filepath.Join(filepath.Dir(p), filepath.FromSlash(target))
			if strings.HasPrefix(target, "/") {
				resolved = r.abs(target)
			}
			if _, err := os.Stat(resolved); err != nil {
				out = append(out, Violation{"R-DOC-001", rel, "broken link: " + target, "fix the link or restore the target"})
			}
		}
		return nil
	})
	return out
}

// markdownLinks returns local link targets outside fenced code blocks, without anchors.
func markdownLinks(data []byte) []string {
	var links []string
	inFence := false
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		for _, m := range linkRe.FindAllStringSubmatch(line, -1) {
			target, _, _ := strings.Cut(m[1], "#")
			if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			links = append(links, target)
		}
	}
	return links
}

func checkRulesDocumented(r *Repo) []Violation {
	data, err := r.read(validationPath)
	if err != nil {
		return nil // R-STRUCT-001 reports the missing file
	}
	documented := map[string]bool{}
	for _, m := range ruleRowRe.FindAllSubmatch(data, -1) {
		documented[string(m[1])] = true
	}
	var out []Violation
	for _, rule := range Rules() {
		if !documented[rule.ID] {
			out = append(out, Violation{"R-DOC-002", validationPath, rule.ID + " is implemented but not documented", "add a table row for it"})
		}
		delete(documented, rule.ID)
	}
	for id := range documented {
		out = append(out, Violation{"R-DOC-002", validationPath, id + " is documented but not implemented", "remove the row or implement the rule in tools/harness"})
	}
	return out
}

func checkArchitectureMentions(r *Repo) []Violation {
	if r.Catalog == nil {
		return nil
	}
	arch, err := r.read("ARCHITECTURE.md")
	if err != nil {
		return nil
	}
	var out []Violation
	for _, d := range r.Catalog.Deployables {
		if d.Status != "deprecated" && !bytes.Contains(arch, []byte(d.Path)) {
			out = append(out, Violation{"R-DOC-003", "ARCHITECTURE.md", d.Path + " is in the catalog but not in ARCHITECTURE.md", "add it to the service boundaries table (§2)"})
		}
	}
	return out
}
