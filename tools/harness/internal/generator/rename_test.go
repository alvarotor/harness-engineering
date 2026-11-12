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

// repoCopy copies the repository without .git into a temp folder, keeping file modes.
func repoCopy(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "repo")
	err := filepath.WalkDir(repoRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		rel, _ := filepath.Rel(repoRoot, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p) //nolint:gosec // copying repository files this test owns
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm()) //nolint:gosec // target is inside t.TempDir()
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

func runRename(t *testing.T, repo string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command("/bin/sh", append([]string{filepath.Join(repo, "scripts", "rename.sh")}, args...)...) //nolint:gosec // fixed script path inside the test fixture
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

func read(t *testing.T, p string) string {
	t.Helper()
	data, err := os.ReadFile(p) //nolint:gosec // file inside t.TempDir()
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// currentSlug reads the slug the way rename.sh does, so the test never hard-codes the template slug.
func currentSlug(t *testing.T, repo string) string {
	t.Helper()
	for _, line := range strings.Split(read(t, filepath.Join(repo, "infra", "compose", "compose.yaml")), "\n") {
		if strings.HasPrefix(line, "name:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
	}
	t.Fatal("no name: in compose.yaml")
	return ""
}

func TestRename_ReplacesEveryFormAndKeepsModes(t *testing.T) {
	repo := repoCopy(t)
	old := currentSlug(t, repo)
	forms := []string{old, strings.ReplaceAll(old, "-", "_"), strings.ToUpper(strings.ReplaceAll(old, "-", "_"))}
	script := filepath.Join(repo, "scripts", "bootstrap.sh")
	before, _ := os.Stat(script)

	out, code := runRename(t, repo, "acme-shop")
	if code != 0 {
		t.Fatalf("exit = %d, output:\n%s", code, out)
	}
	err := filepath.WalkDir(repo, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(p, ".rename-tmp") {
			t.Errorf("temporary file left: %s", p)
		}
		data := read(t, p)
		for _, f := range forms {
			if strings.Contains(data, f) {
				t.Errorf("%s still contains %q", p, f)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := currentSlug(t, repo); got != "acme-shop" {
		t.Errorf("compose name = %q", got)
	}
	if s := read(t, filepath.Join(repo, "infra", "env", "local", "compose.env")); !strings.Contains(s, "ACME_SHOP_SECRETS_DIR=") {
		t.Errorf("compose.env not renamed:\n%s", s)
	}
	if s := read(t, filepath.Join(repo, "contracts", "AGENTS.md")); !strings.Contains(s, "acme_shop.<domain>") {
		t.Errorf("proto package not renamed to snake case:\n%s", s)
	}
	after, _ := os.Stat(script)
	if before.Mode() != after.Mode() {
		t.Errorf("bootstrap.sh mode %v -> %v", before.Mode(), after.Mode())
	}

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not installed")
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = filepath.Join(repo, "tools", "harness")
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("renamed harness does not build: %v\n%s", err, out)
	}
}

func TestRename_RefusesWithoutChanges(t *testing.T) {
	repo := repoCopy(t)
	old := currentSlug(t, repo)
	composePath := filepath.Join(repo, "infra", "compose", "compose.yaml")
	want := read(t, composePath)
	for _, tc := range []struct{ slug, msg string }{
		{"Acme", "kebab-case"}, {"1shop", "kebab-case"}, {"acme_shop", "kebab-case"}, {"acme-", "kebab-case"},
		{"a--b", "kebab-case"}, {"a/b", "kebab-case"}, {old, "already"},
	} {
		t.Run(tc.slug, func(t *testing.T) {
			out, code := runRename(t, repo, tc.slug)
			if code != 1 || !strings.Contains(out, tc.msg) {
				t.Errorf("exit = %d, output %q; want exit 1 with %q", code, out, tc.msg)
			}
			if read(t, composePath) != want {
				t.Error("refused rename changed compose.yaml")
			}
		})
	}
}

func TestRename_UsageError(t *testing.T) {
	repo := repoCopy(t)
	if out, code := runRename(t, repo); code != 2 || !strings.Contains(out, "usage:") {
		t.Errorf("exit = %d, output %q; want exit 2 with usage", code, out)
	}
}
