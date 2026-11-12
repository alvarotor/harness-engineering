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

// validDockerfile passes R-IMG-001: pinned bases, a stage reference, a non-root USER and a HEALTHCHECK.
const validDockerfile = `FROM golang:1.26-alpine@sha256:abc AS build
RUN go build -o /out/service ./cmd/service
FROM build AS lint
RUN go vet ./...
FROM scratch
COPY --from=build /out/service /service
USER 65532:65532
HEALTHCHECK CMD ["/service","healthcheck"]
`

// validRepo writes a minimal repository that passes every rule.
func validRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"AGENTS.md":       "See [arch](ARCHITECTURE.md#1-system-context), [web](https://example.com).\n```\n[ignored](missing.md)\n```\n",
		"ARCHITECTURE.md": "services/core services/auth apps/web\n",
		"CONTRIBUTING.md": "x", "Makefile": "x",
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
		"services/core/Dockerfile":             validDockerfile,
		"services/core/compose.yaml":           "services: {}\n",
		"docs/standards/new-product.md":        "x",
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

func TestSupersededPreference(t *testing.T) {
	root := validRepo(t)
	replace(t, root, registryPath,
		"  - id: P-0001\n    statement: keep it simple\n    source: user\n",
		"  - id: P-0001\n    statement: keep it simple\n    source: user\n    status: superseded\n    superseded_by: P-0002\n  - id: P-0002\n    statement: revised preference\n    source: user\n")
	_, violations := Run(root)
	if len(violations) != 0 {
		t.Fatalf("superseded preference should be valid, got %v", violations)
	}
}

func TestPreferenceEnforcementMetadata(t *testing.T) {
	root := validRepo(t)
	replace(t, root, registryPath,
		"source: user\nassumptions",
		"source: user\n    rationale: keep choices easy to revisit\n    enforced_by: [AGENTS.md]\n    planned_enforced_by: [future/preference-check.md]\nassumptions")
	_, violations := Run(root)
	if len(violations) != 0 {
		t.Fatalf("valid preference enforcement metadata should pass, got %v", violations)
	}
}

func TestPreferenceEnforcedByMissingPath(t *testing.T) {
	root := validRepo(t)
	replace(t, root, registryPath,
		"source: user\nassumptions",
		"source: user\n    enforced_by: [missing/preference-check.md]\nassumptions")
	_, violations := Run(root)
	for _, violation := range violations {
		if violation.Rule == "R-REG-004" && violation.Path == "missing/preference-check.md" && strings.Contains(violation.Problem, "P-0001") {
			return
		}
	}
	t.Fatalf("want R-REG-004 violation for missing preference path, got %v", violations)
}

func TestHistoryFileIsMerged(t *testing.T) {
	root := validRepo(t)
	replace(t, root, registryPath, "  - id: P-0001\n", "  - id: P-0002\n    statement: newer\n    source: user\n  - id: P-0001\n")
	replace(t, root, registryPath, "changelog:\n  - date: 2026-10-15\n    changes: [init]\n", "")
	write(t, root, historyPath, "version: 1\npreferences:\n  - id: P-0003\n    statement: old\n    source: user\n    status: superseded\n    superseded_by: P-0002\nchangelog:\n  - date: 2026-10-15\n    changes: [init]\n")
	_, violations := Run(root)
	if len(violations) != 0 {
		t.Fatalf("superseded entry in history should be valid, got %v", violations)
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
		{"bad preference status", "R-REG-001", func(t *testing.T, r string) {
			replace(t, r, registryPath, "source: user\nassumptions", "source: user\n    status: maybe\nassumptions")
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
		{"superseded preference without replacement", "R-REG-005", func(t *testing.T, r string) {
			replace(t, r, registryPath, "source: user\nassumptions", "source: user\n    status: superseded\nassumptions")
		}},
		{"superseded preference with unknown replacement", "R-REG-005", func(t *testing.T, r string) {
			replace(t, r, registryPath, "source: user\nassumptions", "source: user\n    status: superseded\n    superseded_by: P-9999\nassumptions")
		}},
		{"preference cannot supersede a decision", "R-REG-005", func(t *testing.T, r string) {
			replace(t, r, registryPath, "source: user\nassumptions", "source: user\n    status: superseded\n    superseded_by: D-0001\nassumptions")
		}},
		{"decision cannot be superseded by a preference", "R-REG-005", func(t *testing.T, r string) {
			replace(t, r, registryPath, "title: TS apps\n    status: accepted\n    source: default", "title: TS apps\n    status: superseded\n    source: default\n    superseded_by: P-0001")
		}},
		{"superseded_by on accepted preference", "R-REG-005", func(t *testing.T, r string) {
			replace(t, r, registryPath, "source: user\nassumptions", "source: user\n    status: accepted\n    superseded_by: P-0002\nassumptions")
		}},
		{"catalog invalid yaml", "R-CAT-001", func(t *testing.T, r string) { write(t, r, catalogPath, "deployables: {") }},
		{"history invalid yaml", "R-REG-001", func(t *testing.T, r string) { write(t, r, historyPath, "decisions: [") }},
		{"current entry in history", "R-REG-006", func(t *testing.T, r string) {
			write(t, r, historyPath, "version: 1\npreferences:\n  - id: P-0009\n    statement: still current\n    source: user\nassumptions:\n  - id: A9\n    statement: x\n")
		}},
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
		{"missing compose file", "R-CAT-004", func(t *testing.T, r string) { remove(t, r, "services/core/compose.yaml") }},
		{"missing new-product standard", "R-STRUCT-001", func(t *testing.T, r string) { remove(t, r, "docs/standards/new-product.md") }},
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
		{"compose port on non-edge service", "R-COMPOSE-001", func(t *testing.T, r string) {
			write(t, r, "infra/compose/base.yaml", "services:\n  api:\n    ports: [\"8080:8080\"]\n")
		}},
		{"compose file does not parse", "R-COMPOSE-001", func(t *testing.T, r string) { write(t, r, "infra/compose/base.yaml", "services: [") }},
		{"docker socket mounted by another service", "R-COMPOSE-002", func(t *testing.T, r string) {
			write(t, r, "infra/compose/base.yaml", "services:\n  api:\n    volumes: [\"/var/run/docker.sock:/var/run/docker.sock:ro\"]\n")
		}},
		{"socket-proxy mounts socket read-write", "R-COMPOSE-002", func(t *testing.T, r string) {
			write(t, r, "infra/compose/base.yaml", "services:\n  socket-proxy:\n    volumes:\n      - type: bind\n        source: /var/run/docker.sock\n        target: /var/run/docker.sock\n")
		}},
		{"container_name set", "R-COMPOSE-003", func(t *testing.T, r string) {
			write(t, r, "services/core/compose.yaml", "services:\n  core:\n    container_name: core\n")
		}},
		{"links used", "R-COMPOSE-003", func(t *testing.T, r string) {
			write(t, r, "services/core/compose.yaml", "services:\n  core:\n    links: [db]\n")
		}},
		{"label routes ops port", "R-COMPOSE-003", func(t *testing.T, r string) {
			write(t, r, "services/core/compose.yaml", "services:\n  core:\n    labels:\n      traefik.http.services.core.loadbalancer.server.port: 8081\n")
		}},
		{"label list routes ops port", "R-COMPOSE-003", func(t *testing.T, r string) {
			write(t, r, "services/core/compose.yaml", "services:\n  core:\n    labels: [\"traefik.http.services.core.loadbalancer.server.port=8081\"]\n")
		}},
		{"third-party image without digest", "R-COMPOSE-004", func(t *testing.T, r string) {
			write(t, r, "infra/compose/base.yaml", "services:\n  db:\n    image: postgres:17\n")
		}},
		{"base image without digest", "R-IMG-001", func(t *testing.T, r string) {
			replace(t, r, "services/core/Dockerfile", "golang:1.26-alpine@sha256:abc", "golang:1.26-alpine")
		}},
		{"template base without digest", "R-IMG-001", func(t *testing.T, r string) {
			write(t, r, "templates/go-service/Dockerfile", "FROM alpine:3.20\nUSER 65532\nHEALTHCHECK CMD [\"/x\"]\n")
		}},
		{"runs as root", "R-IMG-001", func(t *testing.T, r string) {
			replace(t, r, "services/core/Dockerfile", "USER 65532:65532", "USER root")
		}},
		{"no USER", "R-IMG-001", func(t *testing.T, r string) {
			replace(t, r, "services/core/Dockerfile", "USER 65532:65532\n", "")
		}},
		{"no HEALTHCHECK", "R-IMG-001", func(t *testing.T, r string) {
			replace(t, r, "services/core/Dockerfile", "HEALTHCHECK CMD [\"/service\",\"healthcheck\"]\n", "")
		}},
		{"USER only in builder stage", "R-IMG-001", func(t *testing.T, r string) {
			write(t, r, "services/core/Dockerfile", "FROM golang@sha256:abc AS build\nUSER 65532\nHEALTHCHECK CMD [\"/x\"]\nFROM scratch\nCOPY --from=build /x /x\n")
		}},
		{"HEALTHCHECK only in builder stage", "R-IMG-001", func(t *testing.T, r string) {
			write(t, r, "services/core/Dockerfile", "FROM golang@sha256:abc AS build\nHEALTHCHECK CMD [\"/x\"]\nFROM scratch\nUSER 65532\nCOPY --from=build /x /x\n")
		}},
		{"template fragment publishes a port", "R-COMPOSE-001", func(t *testing.T, r string) {
			write(t, r, "templates/go-service/compose.yaml", "services:\n  __NAME__:\n    ports: [\"8080:8080\"]\n")
		}},
		{"template fragment mounts docker socket", "R-COMPOSE-002", func(t *testing.T, r string) {
			write(t, r, "templates/go-service/compose.yaml", "services:\n  __NAME__:\n    volumes: [\"/var/run/docker.sock:/var/run/docker.sock\"]\n")
		}},
		{"template fragment has a 8081 route", "R-COMPOSE-003", func(t *testing.T, r string) {
			write(t, r, "templates/go-service/compose.yaml", "services:\n  __NAME__:\n    labels: [\"x.loadbalancer.server.port=8081\"]\n")
		}},
		{"run socket alias mounted", "R-COMPOSE-002", func(t *testing.T, r string) {
			write(t, r, "infra/compose/base.yaml", "services:\n  api:\n    volumes: [\"/run/docker.sock:/run/docker.sock\"]\n")
		}},
		{"socket-proxy alias mounted read-write", "R-COMPOSE-002", func(t *testing.T, r string) {
			write(t, r, "infra/compose/base.yaml", "services:\n  socket-proxy:\n    volumes: [\"/var/run/docker.sock:/var/run/docker.sock:ro\", \"/run/docker.sock:/run/docker.sock\"]\n")
		}},
		{"deployable without cap_drop", "R-COMPOSE-005", func(t *testing.T, r string) {
			write(t, r, "services/core/compose.yaml", "services:\n  core:\n    read_only: true\n    security_opt: [\"no-new-privileges:true\"]\n")
		}},
		{"deployable without no-new-privileges", "R-COMPOSE-005", func(t *testing.T, r string) {
			write(t, r, "services/core/compose.yaml", "services:\n  core:\n    read_only: true\n    cap_drop: [ALL]\n")
		}},
		{"deployable with writable root", "R-COMPOSE-005", func(t *testing.T, r string) {
			write(t, r, "services/core/compose.yaml", "services:\n  core:\n    cap_drop: [ALL]\n    security_opt: [\"no-new-privileges:true\"]\n    read_only: false\n")
		}},
		{"template fragment not hardened", "R-COMPOSE-005", func(t *testing.T, r string) {
			write(t, r, "templates/go-service/compose.yaml", "services:\n  __NAME__:\n    image: x@sha256:abc\n")
		}},
		{"workflow without permissions", "R-CI-002", func(t *testing.T, r string) {
			write(t, r, ".github/workflows/ci.yml", "jobs:\n  test:\n    steps:\n      - run: make check\n")
		}},
		{"run interpolates expression", "R-CI-002", func(t *testing.T, r string) {
			write(t, r, ".github/workflows/ci.yml", "permissions:\n  contents: read\njobs:\n  test:\n    steps:\n      - run: echo ${{ github.head_ref }}\n")
		}},
		{"action not pinned to SHA", "R-CI-001", func(t *testing.T, r string) {
			write(t, r, ".github/workflows/ci.yml", "jobs:\n  test:\n    steps:\n      - uses: actions/checkout@v4\n")
		}},
		{"workflow does not parse", "R-CI-001", func(t *testing.T, r string) { write(t, r, ".github/workflows/ci.yml", "jobs: [") }},
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
		case "R-STRUCT-001", "R-REG-001", "R-CAT-001", "R-DOC-001": // DOC-001: AGENTS.md links to the missing ARCHITECTURE.md
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

// Without a scope decision the built-in defaults apply: services go, apps typescript, no exceptions.
func TestLanguageBuiltinDefaults(t *testing.T) {
	r := validRepo(t)
	replace(t, r, registryPath, "scope: services/*", "scope: none")
	replace(t, r, registryPath, "scope: apps/*", "scope: none")
	_, vs := Run(r)
	var got []string
	for _, v := range vs {
		if v.Rule == "R-CAT-002" {
			got = append(got, v.Path)
			if !strings.Contains(v.Problem, "built-in default") {
				t.Errorf("problem %q does not name the built-in default", v.Problem)
			}
		}
	}
	// services/core (go) and apps/web (typescript) pass; services/auth (typescript) loses its exception.
	if len(got) != 1 || got[0] != "services/auth" {
		t.Fatalf("R-CAT-002 paths = %v, want [services/auth]", got)
	}
}
