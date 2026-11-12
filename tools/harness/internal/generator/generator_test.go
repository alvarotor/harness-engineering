// Package generator tests scripts/new-service.sh and scripts/rename.sh against a copy of the repository.
package generator

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is the repository, relative to this package (tools/harness/internal/generator).
var repoRoot, _ = filepath.Abs("../../../..")

// fixture copies the script and the template into a temp repository, plus a coverage artifact a developer may
// have left in the template. Tests never write into the real templates folder.
func fixture(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "repo")
	copyTree(t, filepath.Join(repoRoot, "templates", "go-service"), filepath.Join(repo, "templates", "go-service"))
	script := filepath.Join(repoRoot, "scripts", "new-service.sh")
	copyTree(t, script, filepath.Join(repo, "scripts", "new-service.sh"))
	for _, p := range []string{"coverage.out", "bin/service"} {
		write(t, filepath.Join(repo, "templates", "go-service", p), "artifact")
	}
	return repo
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		data, err := os.ReadFile(p) //nolint:gosec // copying repository files this test owns
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o700) //nolint:gosec // executable copies of the script and templates
	})
	if err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil { //nolint:gosec // test fixture inside t.TempDir()
		t.Fatal(err)
	}
}

// runScript runs the copied script and returns its combined output and exit code.
func runScript(t *testing.T, repo string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command("/bin/sh", append([]string{filepath.Join(repo, "scripts", "new-service.sh")}, args...)...) //nolint:gosec // fixed script path inside the test fixture
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(out), exitErr.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(out), 0
}

func TestNewService_GeneratesServiceWithoutPlaceholders(t *testing.T) {
	repo := fixture(t)
	root := t.TempDir()
	out, code := runScript(t, repo, "order-api", root)
	if code != 0 {
		t.Fatalf("exit = %d, output:\n%s", code, out)
	}
	dest := filepath.Join(root, "services", "order-api")

	for _, f := range []string{"go.mod", "Dockerfile", "Makefile", "compose.yaml", "README.md", "AGENTS.md", ".dockerignore", "cmd/service/main.go", "internal/platform/config.go"} {
		if _, err := os.Stat(filepath.Join(dest, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}
	for _, f := range []string{"coverage.out", "bin"} {
		if _, err := os.Stat(filepath.Join(dest, f)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s was copied; coverage and build artifacts must be skipped", f)
		}
	}
	err := filepath.WalkDir(dest, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(p) //nolint:gosec // generated file inside t.TempDir()
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "__NAME__") || strings.Contains(string(data), "__ENVNAME__") {
			t.Errorf("placeholder left in %s", p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	gomod, _ := os.ReadFile(filepath.Join(dest, "go.mod")) //nolint:gosec // generated file inside t.TempDir()
	if !strings.Contains(string(gomod), "module harness-engineering/services/order-api") {
		t.Errorf("go.mod = %q", gomod)
	}
	compose, _ := os.ReadFile(filepath.Join(dest, "compose.yaml")) //nolint:gosec // generated file inside t.TempDir()
	if !strings.Contains(string(compose), "${IMAGE_ORDER_API:-harness-engineering-order-api:local}") {
		t.Errorf("compose.yaml does not use the UPPER_SNAKE image variable:\n%s", compose)
	}
	if !strings.Contains(out, "make -C "+dest+" test") {
		t.Errorf("next steps missing from output:\n%s", out)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "services"))
	if len(entries) != 1 || entries[0].Name() != "order-api" {
		t.Errorf("services folder = %v, want only order-api (no temp leftovers)", entries)
	}
}

func TestNewService_GeneratedServiceTestsPass(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not installed")
	}
	repo := fixture(t)
	root := t.TempDir()
	if out, code := runScript(t, repo, "demo", root); code != 0 {
		t.Fatalf("exit = %d, output:\n%s", code, out)
	}
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = filepath.Join(root, "services", "demo")
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated service tests failed: %v\n%s", err, out)
	}
}

func TestNewService_RefusesInvalidNameWithoutChanges(t *testing.T) {
	repo := fixture(t)
	for _, name := range []string{"", "Orders", "1api", "order_api", "order/api", "../escape", "-api", "api-", "api--x", "api-x-", "api\nx"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			out, code := runScript(t, repo, name, root)
			if code != 1 || !strings.Contains(out, "kebab-case") {
				t.Errorf("exit = %d, output %q; want exit 1 naming the rule", code, out)
			}
			entries, _ := os.ReadDir(root)
			if len(entries) != 0 {
				t.Errorf("refused name still changed %s: %v", root, entries)
			}
		})
	}
}

func TestNewService_RefusesExistingFolderWithoutChanges(t *testing.T) {
	repo := fixture(t)
	root := t.TempDir()
	dest := filepath.Join(root, "services", "demo")
	write(t, filepath.Join(dest, "keep.txt"), "mine")

	out, code := runScript(t, repo, "demo", root)
	if code != 1 || !strings.Contains(out, "already exists") {
		t.Fatalf("exit = %d, output %q; want exit 1 naming the existing folder", code, out)
	}
	entries, _ := os.ReadDir(dest)
	if len(entries) != 1 || entries[0].Name() != "keep.txt" {
		t.Errorf("existing folder changed: %v", entries)
	}
	data, _ := os.ReadFile(filepath.Join(dest, "keep.txt")) //nolint:gosec // test fixture inside t.TempDir()
	if string(data) != "mine" {
		t.Errorf("keep.txt = %q, want unchanged", data)
	}
}

func TestNewService_UsageError(t *testing.T) {
	repo := fixture(t)
	if out, code := runScript(t, repo); code != 2 || !strings.Contains(out, "usage:") {
		t.Errorf("exit = %d, output %q; want exit 2 with usage", code, out)
	}
}
