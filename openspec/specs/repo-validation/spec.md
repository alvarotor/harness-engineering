# repo-validation Specification

## Purpose
Defines the single repository validation command, the rules it checks, and the testing standard for the validation tooling itself.

## Requirements

### Requirement: Single validation entry point
The repository SHALL provide one command that runs all available repository validations and exits non-zero if any fail.

#### Scenario: Clean repository
- **WHEN** the validation command runs on a compliant repository
- **THEN** it exits with status 0 and reports the number of rules checked

#### Scenario: Violations found
- **WHEN** one or more rules are violated
- **THEN** each violation is printed with the rule ID, the offending path, what is wrong, and how to fix it
- **AND** the command exits non-zero

### Requirement: Validated rule set
Validation SHALL check at least: required repository files; registry schema, uniqueness and ADR consistency (existence, status, index); that `enforced_by` paths exist; supersession links; service catalog schema; catalog/filesystem agreement for deployables; required files of active deployables; language rules; role definitions and adapters; scoped agent instructions; that relative documentation links resolve; and that every rule is documented.

#### Scenario: Undocumented deployable
- **WHEN** a directory exists under `apps/` or `services/` without a catalog entry
- **THEN** validation fails and asks for a catalog entry or removal of the directory

#### Scenario: Broken documentation link
- **WHEN** a Markdown file links to a relative path that does not exist
- **THEN** validation fails and names the file and the link

#### Scenario: Rule not documented
- **WHEN** a validation rule exists without an entry in the validation standard
- **THEN** validation fails

### Requirement: Validation tooling meets the testing standard
The validation tooling SHALL itself have unit tests for every rule and line coverage of at least 80 %.

#### Scenario: Coverage below threshold
- **WHEN** the validation tool's test coverage falls below 80 %
- **THEN** the coverage check fails and reports the measured value

### Requirement: One-command machine bootstrap
The repository SHALL provide one executable script that prepares a supported machine (macOS or Linux) from only git and curl: it ensures a make program, installs the pinned tool versions without administrator rights, checks the container runtime, and runs the tool check. It MUST verify the integrity of any binary it downloads, MUST NOT install anything needing administrator rights without an interactive user, and MUST be safe to re-run.

#### Scenario: Fresh machine
- **WHEN** the script runs on a supported machine with only git and curl
- **THEN** it installs the pinned tools and reports ready, or lists every remaining problem with its fix and exits non-zero

#### Scenario: Tool check fails later
- **WHEN** the tool check reports a mismatch
- **THEN** it tells the user to run the bootstrap script

#### Scenario: Tampered download
- **WHEN** a downloaded binary does not match its pinned checksum
- **THEN** it is not installed and the problem is reported

#### Scenario: Unsupported OS
- **WHEN** the script runs on an OS other than macOS or Linux
- **THEN** it exits non-zero and states the supported OSes

#### Scenario: Not executable
- **WHEN** the bootstrap script loses its executable permission
- **THEN** repository validation fails with the command to fix it

### Requirement: Compose publishing and socket rules are validated
Repository validation SHALL check the Compose files of the infrastructure, of every active deployable and of the template against the publishing, socket, naming and image rules. Each violation MUST name the service and the file.

#### Scenario: Non-edge service publishes a port
- **WHEN** a Compose service other than the edge proxy declares a published host port
- **THEN** validation fails, names the service and the Compose file, and says that only the edge proxy may publish ports

#### Scenario: Socket mounted outside the socket proxy
- **WHEN** a Compose service other than the socket proxy mounts the container runtime socket, or mounts it read-write
- **THEN** validation fails and names the service

#### Scenario: Third-party image without digest
- **WHEN** a Compose service uses a literal image reference without a digest
- **THEN** validation fails and names the service and the image reference

### Requirement: Deployable Compose services drop capabilities
Repository validation SHALL fail when a Compose service in a deployable fragment does not drop all Linux capabilities, does not set `no-new-privileges`, or does not run with a read-only root filesystem. Services in the infrastructure Compose files are exempt and are reviewed case by case.

#### Scenario: Deployable service without capability drop
- **WHEN** a deployable's Compose service does not drop all capabilities
- **THEN** validation fails and names the service and the missing setting

### Requirement: Dockerfile user and health check are in the shipped stage
Repository validation SHALL check that every base image in a Dockerfile of an active deployable or of the template is pinned by digest, and that the final stage sets a non-root `USER` and an active `HEALTHCHECK`. A value set only in a build stage MUST NOT satisfy the rule.

#### Scenario: Dockerfile runs as root
- **WHEN** an active deployable's final stage sets no `USER`, or sets `USER root`
- **THEN** validation fails and names the Dockerfile

#### Scenario: Non-root user only in the build stage
- **WHEN** a Dockerfile sets `USER` in a build stage and the final stage has no `USER`
- **THEN** validation fails and names the Dockerfile and the final stage

#### Scenario: Final stage pinned and healthy
- **WHEN** the final stage sets a non-root `USER`, a `HEALTHCHECK`, and every base image is digest-pinned
- **THEN** the Dockerfile passes this rule
