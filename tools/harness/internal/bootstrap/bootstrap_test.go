// Package bootstrap tests scripts/bootstrap.sh with fake tools on an isolated PATH.
package bootstrap

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	script, _   = filepath.Abs("../../../../scripts/bootstrap.sh")
	makefile, _ = filepath.Abs("../../../../Makefile")
)

// Real utilities needed by the shell script and Makefile recipes.
var realTools = []string{"bash"}

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
	e.fake("uname", `case "$1" in -s) echo `+osName+`;; esac`)
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

func (e *env) healthy(includeMake bool) {
	e.t.Helper()
	if includeMake {
		e.fake("make", "")
	}
	for _, tool := range []string{"go", "node", "buf", "golangci-lint", "openspec", "docker"} {
		e.fake(tool, "")
	}
}

func (e *env) fakeInstallers() {
	e.t.Helper()
	for _, tool := range []string{"pkg", "apt-get", "dnf", "sudo", "curl", "mise", "tar", "install"} {
		e.fake(tool, "")
	}
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

func (e *env) runDoctor() (string, int) {
	e.t.Helper()
	makePath, err := exec.LookPath("make")
	if err != nil {
		e.t.Skipf("make not available: %v", err)
	}
	cmd := exec.Command(makePath, "-f", makefile, "doctor") //nolint:gosec // fixed Makefile path
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
	e.healthy(true)
	out, code := e.run()
	expect(t, out, code, 0, "all required tools are present", "Next: make check")
	if strings.Contains(e.calls(), "mise") || strings.Contains(e.calls(), "install") || strings.Contains(e.calls(), "curl") {
		t.Errorf("diagnostics must not invoke installers or downloaders:\n%s", e.calls())
	}
}

func TestUnsupportedOS(t *testing.T) {
	e := newEnv(t, "FreeBSD")
	out, code := e.run()
	expect(t, out, code, 1, "supported hosts are macOS, Linux, and Termux")
}

func TestMissingMakePrintsHostGuidanceWithoutInstalling(t *testing.T) {
	cases := []struct {
		name  string
		os    string
		extra []string
		want  string
	}{
		{"macOS", "Darwin", nil, "xcode-select --install"},
		{"Linux", "Linux", nil, "sudo apt-get install -y make"},
		{"Termux by PREFIX", "Linux", []string{"PREFIX=/data/data/com.termux/files/usr"}, "pkg install make"},
		{"Termux by version", "Linux", []string{"TERMUX_VERSION=0.118.1"}, "pkg install make"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t, tc.os)
			e.healthy(false)
			e.fakeInstallers()
			e.extra = tc.extra
			out, code := e.run()
			expect(t, out, code, 1, "missing make", tc.want)
			if strings.Contains(e.calls(), "pkg ") || strings.Contains(e.calls(), "apt-get") || strings.Contains(e.calls(), "dnf") || strings.Contains(e.calls(), "sudo") || strings.Contains(e.calls(), "curl") {
				t.Errorf("bootstrap must only print guidance, got installer calls:\n%s", e.calls())
			}
		})
	}
}

func TestMissingToolsAreReported(t *testing.T) {
	e := newEnv(t, "Linux")
	e.fakeInstallers()
	out, code := e.run()
	expect(t, out, code, 1, "missing make", "missing go", "missing node", "missing buf", "missing golangci-lint", "missing openspec", "missing docker compose")
	if strings.Contains(e.calls(), "install") || strings.Contains(e.calls(), "curl") || strings.Contains(e.calls(), "mise") {
		t.Errorf("bootstrap must not install or download tools:\n%s", e.calls())
	}
}

func TestTermuxWithoutDockerComposeReportsPrerequisite(t *testing.T) {
	e := newEnv(t, "Linux")
	e.extra = append(e.extra, "TERMUX_VERSION=0.118.1")
	e.healthy(true)
	e.fake("docker", `[ "$1" = compose ] && exit 1; exit 0`)
	out, code := e.run()
	expect(t, out, code, 1, "missing docker compose", "Docker Compose availability is not guaranteed in Termux")
	if strings.Contains(out, "Docker Engine") || strings.Contains(out, "install Docker") {
		t.Errorf("Termux must not receive generic Linux Docker installation guidance:\n%s", out)
	}
}

func TestDoctorChecksPresenceWithoutRunningToolVersions(t *testing.T) {
	e := newEnv(t, "Linux")
	e.healthy(true)
	out, code := e.runDoctor()
	expect(t, out, code, 0, "ok go", "ok node", "ok buf", "ok golangci-lint", "ok openspec", "ok docker compose")
	if strings.Contains(e.calls(), "go ") || strings.Contains(e.calls(), "node ") || strings.Contains(e.calls(), "buf ") || strings.Contains(e.calls(), "golangci-lint ") || strings.Contains(e.calls(), "openspec ") {
		t.Errorf("doctor should check command presence without invoking version commands:\n%s", e.calls())
	}
}

// Doctor delegates to the bootstrap diagnostic, so the tool list lives in one place.
func TestDoctorReportsMissingTools(t *testing.T) {
	e := newEnv(t, "Linux")
	e.fake("make", "")
	e.fakeInstallers()
	out, code := e.runDoctor()
	expect(t, out, code, 2, "missing go", "missing node", "missing buf", "missing golangci-lint", "missing openspec", "missing docker compose", "install the reported tools yourself", "ask an AI agent to install them", "ask for your approval first")
	if strings.Contains(e.calls(), "apt-get") || strings.Contains(e.calls(), "curl") || strings.Contains(e.calls(), "pkg ") {
		t.Errorf("doctor must not install or download tools:\n%s", e.calls())
	}
}
