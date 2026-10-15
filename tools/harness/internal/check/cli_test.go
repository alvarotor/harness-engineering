package check

import (
	"bytes"
	"strings"
	"testing"
)

func TestMain_Usage(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Main(nil, &out, &errOut); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if code := Main([]string{"check", "-bogus"}, &out, &errOut); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
}

func TestMain_Clean(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Main([]string{"check", "-root", validRepo(t)}, &out, &errOut); code != 0 {
		t.Fatalf("exit = %d, output:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "0 violations") {
		t.Errorf("output = %q", out.String())
	}
}

func TestMain_ViolationsExplainFix(t *testing.T) {
	root := validRepo(t)
	remove(t, root, "CONTRIBUTING.md")
	var out, errOut bytes.Buffer
	if code := Main([]string{"check", "-root", root}, &out, &errOut); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	for _, want := range []string{"R-STRUCT-001  CONTRIBUTING.md:", "fix: ", "1 violations", "docs/standards/validation.md"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q:\n%s", want, out.String())
		}
	}
}
