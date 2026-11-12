package check

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Compose, container and CI rules (docs/standards/validation.md). Files are parsed loosely: unknown keys never fail.

// dockerSockets are the host paths of the container runtime socket, including the /run alias.
var dockerSockets = []string{"/var/run/docker.sock", "/run/docker.sock"}

const infraPrefix = "infra/"

var (
	composeGlobs  = []string{"infra/compose/*.yaml", "services/*/compose.yaml", "apps/*/compose.yaml", "templates/*/compose.yaml"}
	workflowGlobs = []string{".github/workflows/*.yml", ".github/workflows/*.yaml"}
	fullSHARe     = regexp.MustCompile(`@[0-9a-fA-F]{40}$`)
)

// composeFile is one parsed Compose file. Services are kept as loose maps.
type composeFile struct {
	path     string
	fragment bool // deployable fragment (R-COMPOSE-005); infra/ services are hardened case by case
	services map[string]map[string]any
	err      error
}

func loadCompose(r *Repo) []composeFile {
	var out []composeFile
	for _, pattern := range composeGlobs {
		matches, _ := fs.Glob(os.DirFS(r.Root), pattern)
		for _, rel := range matches {
			cf := parseCompose(r, rel)
			cf.fragment = !strings.HasPrefix(rel, infraPrefix)
			out = append(out, cf)
		}
	}
	return out
}

func parseCompose(r *Repo, rel string) composeFile {
	cf := composeFile{path: rel}
	data, err := r.read(rel)
	if err != nil {
		cf.err = err
		return cf
	}
	var doc struct {
		Services map[string]map[string]any `yaml:"services"`
	}
	cf.err = yaml.Unmarshal(data, &doc)
	cf.services = doc.Services
	return cf
}

// checkComposePorts also reports Compose files that do not parse: every Compose rule needs the parse.
func checkComposePorts(r *Repo) []Violation {
	var out []Violation
	for _, cf := range loadCompose(r) {
		if cf.err != nil {
			out = append(out, Violation{"R-COMPOSE-001", cf.path, "cannot parse Compose file: " + cf.err.Error(), "fix the YAML; `make compose-config` shows the error"})
			continue
		}
		for _, name := range slices.Sorted(maps.Keys(cf.services)) {
			if name != "traefik" && publishes(cf.services[name]) {
				out = append(out, Violation{"R-COMPOSE-001", cf.path, fmt.Sprintf("service %q publishes host ports; only traefik may", name), "remove ports and route through Traefik"})
			}
		}
	}
	return out
}

func publishes(svc map[string]any) bool {
	switch p := svc["ports"].(type) {
	case nil:
		return false
	case []any:
		return len(p) > 0
	default:
		return true // unknown shape: treat as published
	}
}

func checkDockerSocket(r *Repo) []Violation {
	var out []Violation
	for _, cf := range loadCompose(r) {
		if cf.err != nil {
			continue
		}
		for _, name := range slices.Sorted(maps.Keys(cf.services)) {
			ro, mounted := socketMount(cf.services[name])
			if !mounted {
				continue
			}
			switch {
			case name != "socket-proxy":
				out = append(out, Violation{"R-COMPOSE-002", cf.path, fmt.Sprintf("service %q mounts the container runtime socket; only socket-proxy may", name), "route the request through socket-proxy instead of mounting the socket"})
			case !ro:
				out = append(out, Violation{"R-COMPOSE-002", cf.path, "socket-proxy mounts the container runtime socket read-write", "mount it read-only (:ro)"})
			}
		}
	}
	return out
}

// socketMount reports whether a service mounts a runtime socket and whether every such mount is read-only.
func socketMount(svc map[string]any) (ro, mounted bool) {
	vols, _ := svc["volumes"].([]any)
	for _, v := range vols {
		var source string
		var readOnly bool
		switch m := v.(type) {
		case string:
			parts := strings.Split(m, ":")
			source = parts[0]
			readOnly = len(parts) > 2 && slices.Contains(strings.Split(parts[2], ","), "ro")
		case map[string]any:
			source, _ = m["source"].(string)
			readOnly, _ = m["read_only"].(bool)
		}
		if !slices.Contains(dockerSockets, source) {
			continue
		}
		if mounted {
			ro = ro && readOnly
		} else {
			ro = readOnly
		}
		mounted = true
	}
	return ro, mounted
}

func checkComposeRoutes(r *Repo) []Violation {
	var out []Violation
	for _, cf := range loadCompose(r) {
		if cf.err != nil {
			continue
		}
		for _, name := range slices.Sorted(maps.Keys(cf.services)) {
			svc := cf.services[name]
			if _, ok := svc["container_name"]; ok {
				out = append(out, Violation{"R-COMPOSE-003", cf.path, fmt.Sprintf("service %q sets container_name", name), "remove it; Compose names containers, so stacks and replicas do not collide"})
			}
			if _, ok := svc["links"]; ok {
				out = append(out, Violation{"R-COMPOSE-003", cf.path, fmt.Sprintf("service %q uses links", name), "remove it; services reach each other through networks"})
			}
			labels := labelsOf(svc)
			for _, key := range slices.Sorted(maps.Keys(labels)) {
				if strings.Contains(key, "loadbalancer.server.port") && labels[key] == "8081" {
					out = append(out, Violation{"R-COMPOSE-003", cf.path, fmt.Sprintf("service %q routes the ops port 8081 via %s", name, key), "remove the label; port 8081 serves probes only and is never routed"})
				}
			}
		}
	}
	return out
}

// labelsOf accepts both label forms of the Compose spec: a map and a list of "key=value" strings.
func labelsOf(svc map[string]any) map[string]string {
	out := map[string]string{}
	switch l := svc["labels"].(type) {
	case map[string]any:
		for k, v := range l {
			out[k] = fmt.Sprint(v)
		}
	case []any:
		for _, item := range l {
			k, v, _ := strings.Cut(fmt.Sprint(item), "=")
			out[k] = v
		}
	}
	return out
}

func checkComposeImages(r *Repo) []Violation {
	var out []Violation
	for _, cf := range loadCompose(r) {
		if cf.err != nil {
			continue
		}
		for _, name := range slices.Sorted(maps.Keys(cf.services)) {
			img, ok := cf.services[name]["image"].(string)
			if ok && !strings.Contains(img, "${") && !strings.Contains(img, "@sha256:") {
				out = append(out, Violation{"R-COMPOSE-004", cf.path, fmt.Sprintf("service %q image %q is not pinned by digest", name, img), "pin it: image@sha256:<digest> (docker buildx imagetools inspect <image>)"})
			}
		}
	}
	return out
}

// checkHardening: deployable containers drop every capability, forbid privilege gain and keep the root filesystem read-only.
func checkHardening(r *Repo) []Violation {
	var out []Violation
	for _, cf := range loadCompose(r) {
		if cf.err != nil || !cf.fragment {
			continue
		}
		for _, name := range slices.Sorted(maps.Keys(cf.services)) {
			svc := cf.services[name]
			if !hasItem(svc["cap_drop"], "ALL") {
				out = append(out, Violation{"R-COMPOSE-005", cf.path, fmt.Sprintf("service %q does not drop all capabilities", name), "add cap_drop: [ALL]"})
			}
			if !hasItem(svc["security_opt"], "no-new-privileges:true") {
				out = append(out, Violation{"R-COMPOSE-005", cf.path, fmt.Sprintf("service %q does not set no-new-privileges", name), "add security_opt: [\"no-new-privileges:true\"]"})
			}
			if ro, _ := svc["read_only"].(bool); !ro {
				out = append(out, Violation{"R-COMPOSE-005", cf.path, fmt.Sprintf("service %q has a writable root filesystem", name), "add read_only: true and write to a tmpfs or a named volume"})
			}
		}
	}
	return out
}

// hasItem reports whether a YAML list contains the exact string want.
func hasItem(v any, want string) bool {
	l, _ := v.([]any)
	return slices.ContainsFunc(l, func(x any) bool { return x == want })
}

// checkDockerfiles covers active deployables and the service templates, which every new deployable copies.
func checkDockerfiles(r *Repo) []Violation {
	var paths []string
	if r.Catalog != nil {
		for _, d := range r.Catalog.Deployables {
			if d.Status == "active" {
				paths = append(paths, d.Path+"/Dockerfile")
			}
		}
	}
	tpl, _ := fs.Glob(os.DirFS(r.Root), "templates/*/Dockerfile")
	paths = append(paths, tpl...)

	var out []Violation
	for _, p := range paths {
		data, err := r.read(p)
		if err != nil {
			continue // R-CAT-004 reports a missing Dockerfile of an active deployable
		}
		out = append(out, dockerfileProblems(p, data)...)
	}
	return out
}

// dockerfileProblems checks the final stage only: earlier stages are discarded, so their USER and HEALTHCHECK do not count.
func dockerfileProblems(path string, data []byte) []Violation {
	var out []Violation
	stages := map[string]bool{}
	hasUser, hasHealth := false, false
	user := ""
	for _, line := range instructions(data) {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch strings.ToUpper(fields[0]) {
		case "FROM":
			hasUser, hasHealth, user = false, false, ""
			image, alias := fromArgs(fields[1:])
			if image != "" && image != "scratch" && !stages[strings.ToLower(image)] && !strings.Contains(image, "@sha256:") {
				out = append(out, Violation{"R-IMG-001", path, fmt.Sprintf("base image %q is not pinned by digest", image), "pin it: image@sha256:<digest> (docker buildx imagetools inspect <image>)"})
			}
			if alias != "" {
				stages[strings.ToLower(alias)] = true
			}
		case "USER":
			hasUser = true
			user = ""
			if len(fields) > 1 {
				user, _, _ = strings.Cut(fields[1], ":")
			}
		case "HEALTHCHECK":
			// HEALTHCHECK NONE disables a check inherited from a base image, so the last one decides.
			hasHealth = len(fields) < 2 || !strings.EqualFold(fields[1], "NONE")
		}
	}
	if !hasUser || user == "" || user == "root" || user == "0" {
		out = append(out, Violation{"R-IMG-001", path, "final stage has no non-root USER (missing, empty, root or 0)", "set a non-root user in the final stage, for example USER 65532:65532 (runtime contract)"})
	}
	if !hasHealth {
		out = append(out, Violation{"R-IMG-001", path, "final stage has no HEALTHCHECK instruction", "add HEALTHCHECK CMD [\"/service\",\"healthcheck\"] in the final stage (probes port 8081)"})
	}
	return out
}

// fromArgs returns the image and the stage alias of a FROM instruction, skipping flags such as --platform.
func fromArgs(args []string) (image, alias string) {
	for len(args) > 0 && strings.HasPrefix(args[0], "--") {
		args = args[1:]
	}
	if len(args) == 0 {
		return "", ""
	}
	if len(args) >= 3 && strings.EqualFold(args[1], "AS") {
		alias = args[2]
	}
	return args[0], alias
}

// instructions returns the logical Dockerfile lines: comments and blank lines dropped, continuations joined.
func instructions(data []byte) []string {
	var out []string
	cur := ""
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if cur == "" && (line == "" || strings.HasPrefix(line, "#")) {
			continue
		}
		cur += strings.TrimSuffix(line, "\\") + " "
		if !strings.HasSuffix(line, "\\") {
			out = append(out, strings.TrimSpace(cur))
			cur = ""
		}
	}
	if strings.TrimSpace(cur) != "" {
		out = append(out, strings.TrimSpace(cur))
	}
	return out
}

// workflow is one parsed GitHub Actions file.
type workflow struct {
	path string
	doc  any
	err  error
}

func loadWorkflows(r *Repo) []workflow {
	var out []workflow
	for _, pattern := range workflowGlobs {
		matches, _ := fs.Glob(os.DirFS(r.Root), pattern)
		for _, rel := range matches {
			w := workflow{path: rel}
			if data, err := r.read(rel); err != nil {
				w.err = err
			} else {
				w.err = yaml.Unmarshal(data, &w.doc)
			}
			out = append(out, w)
		}
	}
	return out
}

// checkActionPins requires every third-party GitHub Action to be pinned to a full commit SHA.
func checkActionPins(r *Repo) []Violation {
	var out []Violation
	for _, w := range loadWorkflows(r) {
		if w.err != nil {
			out = append(out, Violation{"R-CI-001", w.path, "cannot parse workflow: " + w.err.Error(), "fix the YAML"})
			continue
		}
		eachValue(w.doc, "uses", func(uses string) {
			if strings.HasPrefix(uses, "./") || strings.HasPrefix(uses, "docker://") || fullSHARe.MatchString(uses) {
				return
			}
			out = append(out, Violation{"R-CI-001", w.path, fmt.Sprintf("uses %q is not pinned to a full commit SHA", uses), "pin it to the 40-character commit SHA and keep the tag in a comment"})
		})
	}
	return out
}

// checkWorkflowHardening: least-privilege token permissions, and no expressions interpolated into shell scripts.
func checkWorkflowHardening(r *Repo) []Violation {
	var out []Violation
	for _, w := range loadWorkflows(r) {
		if w.err != nil {
			continue // R-CI-001 reports the parse error
		}
		top, _ := w.doc.(map[string]any)
		if _, ok := top["permissions"]; !ok {
			out = append(out, Violation{"R-CI-002", w.path, "workflow has no top-level permissions key", "add permissions: {contents: read} and widen per job only where needed"})
		}
		eachValue(w.doc, "run", func(script string) {
			if strings.Contains(script, "${{") {
				out = append(out, Violation{"R-CI-002", w.path, "run: script interpolates a ${{ }} expression (script injection)", "pass the value through env: and read the shell variable instead"})
			}
		})
	}
	return out
}

// eachValue calls fn for every string stored under key anywhere in the document, in a deterministic order.
func eachValue(node any, key string, fn func(string)) {
	switch n := node.(type) {
	case map[string]any:
		for _, k := range slices.Sorted(maps.Keys(n)) {
			if s, ok := n[k].(string); ok && k == key {
				fn(s)
			}
			eachValue(n[k], key, fn)
		}
	case []any:
		for _, v := range n {
			eachValue(v, key, fn)
		}
	}
}
