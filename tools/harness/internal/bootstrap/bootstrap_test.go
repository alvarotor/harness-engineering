// Package bootstrap tests scripts/bootstrap.sh with fake tools on an isolated PATH.
package bootstrap

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var script, _ = filepath.Abs("../../../../scripts/bootstrap.sh")

// Real utilities the script may use; everything else is faked or absent.
var realTools = []string{"dirname", "mkdir", "mktemp", "rm", "cut", "tar", "install", "sha256sum", "cat"}

type env struct {
	t     *testing.T
	bin   string
	home  string
	log   string
	extra []string
}

func newEnv(t *testing.T, osName string) *env {
	t.Helper()
	dir := t.TempDir()
	e := &env{t: t, bin: filepath.Join(dir, "bin"), home: filepath.Join(dir, "home"), log: filepath.Join(dir, "calls.log")}
	for _, d := range []string{e.bin, e.home} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	for _, tool := range realTools {
		p, err := exec.LookPath(tool)
		if err != nil {
			t.Skipf("%s not available: %v", tool, err)
		}
		if err := os.Symlink(p, filepath.Join(e.bin, tool)); err != nil {
			t.Fatal(err)
		}
	}
	e.fake("uname", `case "$1" in -s) echo `+osName+`;; -m) echo x86_64;; esac`)
	return e
}

// fake writes an executable that logs its name and args, then runs body.
func (e *env) fake(name, body string) {
	e.t.Helper()
	sh := "#!/bin/sh\necho \"" + name + " $*\" >> \"$FAKE_LOG\"\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(e.bin, name), []byte(sh), 0o700); err != nil { //nolint:gosec // test fixture inside t.TempDir()
		e.t.Fatal(err)
	}
}

// healthy installs fakes for a machine where everything is present and working.
func (e *env) healthy() {
	e.fake("make", "")
	e.fake("mise", "")
	e.fake("docker", "")
}

func (e *env) run() (string, int) {
	e.t.Helper()
	cmd := exec.Command("/bin/sh", script) //nolint:gosec // fixed script path under test
	cmd.Env = append([]string{"PATH=" + e.bin, "HOME=" + e.home, "FAKE_LOG=" + e.log}, e.extra...)
	out, _ := cmd.CombinedOutput()
	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	return string(out), code
}

func (e *env) calls() string {
	data, _ := os.ReadFile(e.log)
	return string(data)
}

func expect(t *testing.T, out string, code, wantCode int, wants ...string) {
	t.Helper()
	if code != wantCode {
		t.Errorf("exit = %d, want %d\n%s", code, wantCode, out)
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q:\n%s", w, out)
		}
	}
}

func TestHealthyMachine(t *testing.T) {
	e := newEnv(t, "Linux")
	e.healthy()
	e.extra = []string{"MISE_SHELL=bash"}
	out, code := e.run()
	expect(t, out, code, 0, "ready. Next: make check")
	for _, want := range []string{"mise install", "mise exec -- make doctor", "docker compose version"} {
		if !strings.Contains(e.calls(), want) {
			t.Errorf("expected call %q, got:\n%s", want, e.calls())
		}
	}
	if strings.Contains(out, "tip:") {
		t.Error("activation tip shown although mise is active")
	}
}

func TestUnsupportedOS(t *testing.T) {
	e := newEnv(t, "FreeBSD")
	out, code := e.run()
	expect(t, out, code, 1, "macOS and Linux only")
}

func TestMacWithoutMake(t *testing.T) {
	e := newEnv(t, "Darwin")
	e.fake("mise", "")
	e.fake("docker", "")
	out, code := e.run()
	expect(t, out, code, 1, "xcode-select --install", "not ready")
	if !strings.Contains(e.calls(), "mise install") {
		t.Error("should keep going and install tools even when make is missing")
	}
	if strings.Contains(e.calls(), "doctor") {
		t.Error("doctor must not run without make")
	}
}

func TestUsesMakeProjFallback(t *testing.T) {
	e := newEnv(t, "Linux")
	e.fake("make-proj", "")
	e.fake("mise", "")
	e.fake("docker", "")
	out, code := e.run()
	expect(t, out, code, 0, "Next: make-proj check")
	if !strings.Contains(e.calls(), "mise exec -- make-proj doctor") {
		t.Errorf("doctor should run through make-proj:\n%s", e.calls())
	}
}

func TestLinuxWithoutMakeOrSudoPrintsCommands(t *testing.T) {
	e := newEnv(t, "Linux")
	e.fake("mise", "")
	e.fake("docker", "")
	out, code := e.run()
	expect(t, out, code, 1, "sudo apt-get install -y make")
}

func TestDockerProblems(t *testing.T) {
	cases := map[string]struct{ docker, want string }{
		"missing":     {"", "Docker is missing"},
		"not running": {`[ "$1" = info ] && exit 1; exit 0`, "not running"},
		"no compose":  {`[ "$1" = compose ] && exit 1; exit 0`, "Compose v2 is missing"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newEnv(t, "Linux")
			e.fake("make", "")
			e.fake("mise", "")
			if tc.docker != "" {
				e.fake("docker", tc.docker)
			}
			out, code := e.run()
			expect(t, out, code, 1, tc.want)
		})
	}
}

func TestDoctorFailureIsReported(t *testing.T) {
	e := newEnv(t, "Linux")
	e.healthy()
	e.fake("mise", `case "$*" in *doctor*) exit 1;; esac`)
	out, code := e.run()
	expect(t, out, code, 1, "'make doctor' reported mismatches")
}

func TestMiseInstallFailureIsReported(t *testing.T) {
	e := newEnv(t, "Linux")
	e.healthy()
	e.fake("mise", `[ "$1" = install ] && exit 1; exit 0`)
	out, code := e.run()
	expect(t, out, code, 1, "'mise install' failed")
}

func TestMiseChecksumMismatchIsNotInstalled(t *testing.T) {
	e := newEnv(t, "Linux")
	e.fake("make", "")
	e.fake("docker", "")
	e.fake("curl", `while [ $# -gt 0 ]; do [ "$1" = -o ] && echo tampered > "$2"; shift; done`)
	out, code := e.run()
	expect(t, out, code, 1, "checksum mismatch")
	if _, err := os.Stat(filepath.Join(e.home, ".local/bin/mise")); err == nil {
		t.Error("mise was installed despite a checksum mismatch")
	}
}

func TestMiseDownloadFailureIsReported(t *testing.T) {
	e := newEnv(t, "Linux")
	e.fake("make", "")
	e.fake("docker", "")
	e.fake("curl", "exit 22")
	out, code := e.run()
	expect(t, out, code, 1, "could not download")
}
