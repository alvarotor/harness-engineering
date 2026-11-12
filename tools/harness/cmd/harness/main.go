// Command harness validates the repository against docs/standards/validation.md.
package main

import (
	"os"

	"harness-engineering/tools/harness/internal/check"
)

func main() {
	os.Exit(check.Main(os.Args[1:], os.Stdout, os.Stderr))
}
