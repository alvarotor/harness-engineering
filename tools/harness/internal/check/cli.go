package check

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

// Main runs the CLI and returns the process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "check" {
		_, _ = io.WriteString(stderr, "usage: harness check [-root DIR]\n")
		return 2
	}
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "repository root")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	n, violations := Run(*root)
	var b strings.Builder
	for _, v := range violations {
		fmt.Fprintf(&b, "%s\n\n", v)
	}
	fmt.Fprintf(&b, "harness: %d rules checked, %d violations\n", n, len(violations))
	if len(violations) > 0 {
		b.WriteString("rules and fixes: docs/standards/validation.md\n")
	}
	if _, err := io.WriteString(stdout, b.String()); err != nil {
		return 2
	}
	if len(violations) > 0 {
		return 1
	}
	return 0
}
