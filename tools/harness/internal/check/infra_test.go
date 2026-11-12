package check

import (
	"strings"
	"testing"
)

// The infra rules must accept the files the templates and infra/ are meant to contain.
func TestInfraRulesAcceptValidFiles(t *testing.T) {
	root := validRepo(t)
	sha := strings.Repeat("a", 40)
	write(t, root, "infra/compose/base.yaml", `services:
  traefik:
    image: traefik:v3.1@sha256:abc
    ports: ["80:80"]
  socket-proxy:
    image: tecnativa/docker-socket-proxy@sha256:abc
    volumes:
      - type: bind
        source: /var/run/docker.sock
        target: /var/run/docker.sock
        read_only: true
  db:
    image: ${IMAGE_DB:-postgres:17}
    labels: ["traefik.enable=false", "x.loadbalancer.server.port=8080"]
    unknown_future_key: ok
`)
	write(t, root, "templates/go-service/Dockerfile", validDockerfile)
	write(t, root, ".github/workflows/ci.yml", "permissions:\n  contents: read\njobs:\n  test:\n    uses: ./.github/workflows/local.yml\n    steps:\n      - uses: actions/checkout@"+sha+"\n      - uses: docker://alpine:3.20\n        env:\n          V: ${{ github.sha }}\n        run: echo \"$V\"\n")
	_, vs := Run(root)
	for _, v := range vs {
		t.Errorf("unexpected violation: %s", v)
	}
}

// Only the final stage ships, so USER and HEALTHCHECK count only there.
func TestDockerfileProblems_FinalStageOnly(t *testing.T) {
	const pinned = "golang:1.26-alpine@sha256:abc"
	cases := []struct {
		name string
		data string
		want []string // expected problem substrings, one per violation
	}{
		{"final stage complete", "FROM " + pinned + " AS build\nRUN go build\nFROM scratch\nUSER 65532:65532\nHEALTHCHECK CMD [\"/s\"]\n", nil},
		{"USER only in builder", "FROM " + pinned + " AS build\nUSER 65532\nHEALTHCHECK CMD [\"/s\"]\nFROM scratch\nCOPY --from=build /s /s\n", []string{"no non-root USER", "no HEALTHCHECK"}},
		{"HEALTHCHECK only in builder", "FROM " + pinned + " AS build\nHEALTHCHECK CMD [\"/s\"]\nFROM scratch\nUSER 65532\n", []string{"no HEALTHCHECK"}},
		{"HEALTHCHECK NONE in final stage", "FROM scratch\nUSER 65532\nHEALTHCHECK CMD [\"/s\"]\nHEALTHCHECK none\n", []string{"no HEALTHCHECK"}},
		{"HEALTHCHECK NONE overrides a check in the same stage", "FROM scratch\nUSER 65532\nHEALTHCHECK NONE\n", []string{"no HEALTHCHECK"}},
		{"HEALTHCHECK re-enabled after NONE", "FROM scratch\nUSER 65532\nHEALTHCHECK NONE\nHEALTHCHECK CMD [\"/s\"]\n", nil},
		{"final stage stays root", "FROM scratch\nUSER root\nHEALTHCHECK CMD [\"/s\"]\n", []string{"no non-root USER"}},
		{"numeric root with group", "FROM scratch\nUSER 0:0\nHEALTHCHECK CMD [\"/s\"]\n", []string{"no non-root USER"}},
		{"empty USER", "FROM scratch\nUSER \nHEALTHCHECK CMD [\"/s\"]\n", []string{"no non-root USER"}},
		{"unpinned base and alias reference", "FROM alpine AS base\nFROM base\nUSER 1\nHEALTHCHECK CMD [\"/s\"]\n", []string{"base image \"alpine\""}},
		{"platform flag and continuation", "FROM --platform=$BUILDPLATFORM \\\n  " + pinned + " AS Build\nFROM build\nUSER 1\nHEALTHCHECK CMD [\"/s\"]\n", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dockerfileProblems("x/Dockerfile", []byte(tc.data))
			if len(got) != len(tc.want) {
				t.Fatalf("got %d violations, want %d: %v", len(got), len(tc.want), got)
			}
			for i, v := range got {
				if !strings.Contains(v.Problem, tc.want[i]) {
					t.Errorf("violation %d = %q, want it to contain %q", i, v.Problem, tc.want[i])
				}
			}
		})
	}
}
