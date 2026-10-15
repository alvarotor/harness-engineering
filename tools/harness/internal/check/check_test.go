package check

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureRegistry = `version: 1
decisions:
  - id: D-0001
    title: Go by default
    status: accepted
    source: user
    adr: docs/adr/0001-go.md
    scope: services/*
    default_language: go
    exceptions:
      - path: services/auth
        language: typescript
        reason: library is TS-only
    enforced_by: [AGENTS.md]
    planned_enforced_by: [templates]
  - id: D-0002
    title: TS apps
    status: accepted
    source: default
    adr: docs/adr/0002-ts.md
    scope: apps/*
    default_language: typescript
preferences:
  - id: P-0001
    statement: keep it simple
    source: user
assumptions:
  - id: A1
    statement: slug
open_questions:
  - id: Q1
    question: domain?
    needed_by: later
resolved_questions:
  - id: Q2
    question: repo?
    resolution: later
changelog:
  - date: 2026-10-15
    changes: [init]
`

const fixtureCatalog = `version: 1
deployables:
  - name: core
    kind: service
    path: services/core
    language: go
    status: active
    owner: team
    domain: orders
    personal_data: []
    depends_on: [auth]
  - name: auth
    kind: service
    path: services/auth
    language: typescript
    status: planned
    owner: team
    domain: identity
    personal_data: [email]
  - name: web
    kind: app
    path: apps/web
    language: typescript
    status: planned
    owner: team
    personal_data: []
    depends_on: [core]
`

// validRepo writes a minimal repository that passes every rule.
func validRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"AGENTS.md":       "See [arch](ARCHITECTURE.md#1-system-context), [web](https://example.com).\n```\n[ignored](missing.md)\n```\n",
		"ARCHITECTURE.md": "services/core services/auth apps/web\n",
		"CONTRIBUTING.md": "x", "Makefile": "x", ".tool-versions": "x",
		registryPath:                           fixtureRegistry,
		catalogPath:                            fixtureCatalog,
		"docs/adr/0001-go.md":                  "# 0001\n\n- Status: accepted\n",
		"docs/adr/0002-ts.md":                  "# 0002\n\n- Status: accepted\n",
		"docs/adr/README.md":                   "| [0001](0001-go.md) | Go | accepted | user |\n| [0002](0002-ts.md) | TS | accepted | default |\n",
		"docs/agents/README.md":                "roles",
		"docs/standards/engineering.md":        "x",
		"docs/standards/security.md":           "x",
		"docs/standards/testing.md":            "x",
		"docs/standards/definition-of-done.md": "x",
		"openspec/config.yaml":                 "schema: spec-driven\n",
		"services/core/Dockerfile":             "x",
		"services/core/AGENTS.md":              "x",
		"services/core/README.md":              "x",
		"services/core/go.mod":                 "x",
		"openspec/changes/archive/old/x.md":    "[stale](gone.md)",
	}
	var rows strings.Builder
	for _, r := range Rules() {
		rows.WriteString("| " + r.ID + " | x | x |\n")
	}
	files[validationPath] = rows.String()
	for _, f := range scopedAgents {
		files[f] = "x"
	}
	for _, role := range Roles {
		var b strings.Builder
		for _, s := range roleSections {
			b.WriteString("## " + s + "\ntext\n\n")
		}
		files["docs/agents/"+role+".md"] = b.String()
		files[".opencode/agents/"+role+".md"] = "read docs/agents/" + role + ".md"
	}
	for p, c := range files {
		write(t, root, p, c)
	}
	write(t, root, bootstrapPath, "#!/bin/sh\n")
	if err := os.Chmod(filepath.Join(root, bootstrapPath), 0o700); err != nil { //nolint:gosec // the fixture must be executable
		t.Fatal(err)
	}
	return root
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil { //nolint:gosec // test fixture inside t.TempDir()
		t.Fatal(err)
	}
}

func replace(t *testing.T, root, rel, old, repl string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel)) //nolint:gosec // test fixture inside t.TempDir()
	if err != nil || !bytes.Contains(data, []byte(old)) {
		t.Fatalf("fixture %s does not contain %q", rel, old)
	}
	write(t, root, rel, strings.Replace(string(data), old, repl, 1))
}

func remove(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(root, rel)); err != nil {
		t.Fatal(err)
	}
}

func TestValidRepoHasNoViolations(t *testing.T) {
	n, vs := Run(validRepo(t))
	if n != len(Rules()) {
		t.Errorf("rules run = %d, want %d", n, len(Rules()))
	}
	for _, v := range vs {
		t.Errorf("unexpected violation: %s", v)
	}
}

func TestEachRuleDetectsItsViolation(t *testing.T) {
	cases := []struct {
		name   string
		rule   string
		mutate func(t *testing.T, root string)
	}{
		{"missing required file", "R-STRUCT-001", func(t *testing.T, r string) { remove(t, r, "CONTRIBUTING.md") }},
		{"bootstrap not executable", "R-STRUCT-001", func(t *testing.T, r string) {
			if err := os.Chmod(filepath.Join(r, bootstrapPath), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"registry invalid yaml", "R-REG-001", func(t *testing.T, r string) { write(t, r, registryPath, "decisions: [") }},
		{"registry unknown field", "R-REG-001", func(t *testing.T, r string) { replace(t, r, registryPath, "title: TS apps", "titel: TS apps") }},
		{"malformed id", "R-REG-001", func(t *testing.T, r string) { replace(t, r, registryPath, "id: P-0001", "id: P-1") }},
		{"duplicate id", "R-REG-001", func(t *testing.T, r string) { replace(t, r, registryPath, "id: D-0002", "id: D-0001") }},
		{"bad status", "R-REG-001", func(t *testing.T, r string) { replace(t, r, registryPath, "status: accepted", "status: maybe") }},
		{"bad source", "R-REG-001", func(t *testing.T, r string) { replace(t, r, registryPath, "source: default", "source: chat") }},
		{"bad preference source", "R-REG-001", func(t *testing.T, r string) {
			replace(t, r, registryPath, "source: user\nassumptions", "source: x\nassumptions")
		}},
		{"ADR missing", "R-REG-002", func(t *testing.T, r string) { remove(t, r, "docs/adr/0002-ts.md") }},
		{"ADR status mismatch", "R-REG-002", func(t *testing.T, r string) {
			replace(t, r, "docs/adr/0001-go.md", "accepted", "proposed")
		}},
		{"ADR not in registry", "R-REG-003", func(t *testing.T, r string) { write(t, r, "docs/adr/0003-x.md", "- Status: accepted\n") }},
		{"index status mismatch", "R-REG-003", func(t *testing.T, r string) {
			replace(t, r, "docs/adr/README.md", "| Go | accepted", "| Go | proposed")
		}},
		{"enforced_by missing", "R-REG-004", func(t *testing.T, r string) {
			replace(t, r, registryPath, "enforced_by: [AGENTS.md]", "enforced_by: [nope.md]")
		}},
		{"superseded without replacement", "R-REG-005", func(t *testing.T, r string) {
			replace(t, r, registryPath, "status: accepted\n    source: default", "status: superseded\n    source: default")
		}},
		{"superseded_by on accepted", "R-REG-005", func(t *testing.T, r string) {
			replace(t, r, registryPath, "title: TS apps", "title: TS apps\n    superseded_by: D-0001")
		}},
		{"catalog invalid yaml", "R-CAT-001", func(t *testing.T, r string) { write(t, r, catalogPath, "deployables: {") }},
		{"empty owner", "R-CAT-001", func(t *testing.T, r string) { replace(t, r, catalogPath, "owner: team", `owner: ""`) }},
		{"bad kind", "R-CAT-001", func(t *testing.T, r string) { replace(t, r, catalogPath, "kind: app", "kind: lambda") }},
		{"path mismatch", "R-CAT-001", func(t *testing.T, r string) { replace(t, r, catalogPath, "path: apps/web", "path: apps/webby") }},
		{"bad name and status", "R-CAT-001", func(t *testing.T, r string) {
			replace(t, r, catalogPath, "name: web", "name: Web")
			replace(t, r, catalogPath, "status: planned\n    owner: team\n    personal_data: []", "status: draft\n    owner: team\n    personal_data: []")
		}},
		{"duplicate name", "R-CAT-001", func(t *testing.T, r string) { replace(t, r, catalogPath, "name: auth", "name: core") }},
		{"service without domain", "R-CAT-001", func(t *testing.T, r string) { replace(t, r, catalogPath, "domain: orders", `domain: ""`) }},
		{"missing personal_data", "R-CAT-001", func(t *testing.T, r string) {
			replace(t, r, catalogPath, "personal_data: [email]\n", "")
		}},
		{"undeclared backend language", "R-CAT-002", func(t *testing.T, r string) { replace(t, r, catalogPath, "language: go", "language: python") }},
		{"wrong app language", "R-CAT-002", func(t *testing.T, r string) {
			replace(t, r, catalogPath, "path: apps/web\n    language: typescript", "path: apps/web\n    language: go")
		}},
		{"orphan folder", "R-CAT-003", func(t *testing.T, r string) { write(t, r, "services/orphan/x.go", "x") }},
		{"planned but on disk", "R-CAT-003", func(t *testing.T, r string) { write(t, r, "apps/web/index.html", "x") }},
		{"active but missing", "R-CAT-003", func(t *testing.T, r string) { remove(t, r, "services/core") }},
		{"missing Dockerfile", "R-CAT-004", func(t *testing.T, r string) { remove(t, r, "services/core/Dockerfile") }},
		{"unknown dependency", "R-CAT-005", func(t *testing.T, r string) { replace(t, r, catalogPath, "depends_on: [auth]", "depends_on: [ghost]") }},
		{"role missing section", "R-AGENT-001", func(t *testing.T, r string) {
			replace(t, r, "docs/agents/reviewer.md", "## Stop conditions", "## Stop")
		}},
		{"role file missing", "R-AGENT-001", func(t *testing.T, r string) { remove(t, r, "docs/agents/devops.md") }},
		{"extra role", "R-AGENT-001", func(t *testing.T, r string) { write(t, r, "docs/agents/qa-lead.md", "x") }},
		{"adapter missing", "R-AGENT-002", func(t *testing.T, r string) { remove(t, r, ".opencode/agents/planner.md") }},
		{"adapter not pointing to canon", "R-AGENT-002", func(t *testing.T, r string) { write(t, r, ".opencode/agents/planner.md", "you plan") }},
		{"scoped AGENTS missing", "R-SCOPE-001", func(t *testing.T, r string) { remove(t, r, "infra/AGENTS.md") }},
		{"broken relative link", "R-DOC-001", func(t *testing.T, r string) { write(t, r, "docs/x.md", "[a](../nope.md)") }},
		{"broken root link", "R-DOC-001", func(t *testing.T, r string) { write(t, r, "docs/x.md", `[a](/nope.md "t")`) }},
		{"undocumented rule", "R-DOC-002", func(t *testing.T, r string) { replace(t, r, validationPath, "| R-DOC-003 | x | x |\n", "") }},
		{"unimplemented rule", "R-DOC-002", func(t *testing.T, r string) {
			replace(t, r, validationPath, "| R-DOC-003", "| R-NEW-001 | x | x |\n| R-DOC-003")
		}},
		{"catalog path not in architecture", "R-DOC-003", func(t *testing.T, r string) { write(t, r, "ARCHITECTURE.md", "services/core apps/web") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := validRepo(t)
			tc.mutate(t, root)
			_, vs := Run(root)
			found := false
			for _, v := range vs {
				found = found || v.Rule == tc.rule
				if v.Fix == "" || v.Problem == "" {
					t.Errorf("violation without problem/fix: %+v", v)
				}
			}
			if !found {
				t.Errorf("want a %s violation, got %v", tc.rule, vs)
			}
		})
	}
}

func TestBrokenInputsDoNotCascade(t *testing.T) {
	root := validRepo(t)
	remove(t, root, registryPath)
	remove(t, root, catalogPath)
	remove(t, root, validationPath)
	remove(t, root, "ARCHITECTURE.md")
	_, vs := Run(root)
	for _, v := range vs {
		switch v.Rule {
		case "R-STRUCT-001", "R-REG-001", "R-CAT-001", "R-DOC-001": // DOC-001: AGENTS.md links to the removed ARCHITECTURE.md
		default:
			t.Errorf("unexpected cascading violation: %s", v)
		}
	}
}

func TestMarkdownLinks(t *testing.T) {
	got := markdownLinks([]byte("[a](x.md#h) [b](#top) [c](mailto:a@b) [d](http://x)\n```\n[e](y.md)\n```\n[f](z.md \"title\")"))
	want := []string{"x.md", "z.md"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", got, want)
	}
}
