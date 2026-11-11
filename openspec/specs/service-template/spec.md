# service-template Specification

## Purpose
Defines the generated backend service skeleton: it meets the runtime contract out of the box, and generation is safe against bad names and existing folders.

## Requirements

### Requirement: Generated services meet the runtime contract
A service generated from the template SHALL build, pass its own tests with at least 80 % line coverage, answer liveness and readiness on the ops port, and drain on termination. On termination, readiness MUST fail before the process exits, and the process MUST exit within 20 seconds.

#### Scenario: Service is generated
- **WHEN** a service is generated from the template
- **THEN** its image builds, its tests pass with coverage of at least 80 %, and it answers liveness and readiness on the ops port

#### Scenario: Service receives termination
- **WHEN** the generated service receives the termination signal
- **THEN** its readiness check fails before the process exits, and the process exits within 20 seconds

### Requirement: Generated HTTP servers have timeouts
Each HTTP server in a generated service SHALL set explicit read-header, read, write and idle timeouts, so that a client cannot hold a connection open indefinitely. The generated tests MUST assert these values.

#### Scenario: Timeouts are configured
- **WHEN** a service is generated from the template
- **THEN** its tests fail if any server timeout is unset

### Requirement: Secrets are read from files
A generated service SHALL read secret values from files whose paths are given by environment variables. Compose services SHALL receive secret values as mounted files, and the deploy SHALL write each secret to a file on the host. Secret values MUST NOT be passed on the command line.

#### Scenario: Service needs a secret
- **WHEN** a generated service needs a secret
- **THEN** it reads the value from the file named by its `*_FILE` variable

#### Scenario: Local database password
- **WHEN** the local stack starts
- **THEN** the database reads its password from a mounted secret file

### Requirement: Generation is safe
The generator SHALL accept only kebab-case service names matching `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`. It MUST refuse a name that does not match, a name that resolves outside `services/` or `apps/`, or a target folder that already exists, and MUST change nothing when it refuses.

#### Scenario: Name is not kebab-case
- **WHEN** the owner generates a service with an uppercase name, an underscore, a leading digit or a trailing hyphen
- **THEN** the generator refuses, names the rule, and creates no files

#### Scenario: Name tries to escape the folder
- **WHEN** the owner generates a service with a name containing a path separator or `..`
- **THEN** the generator refuses and creates no files

#### Scenario: Folder already exists
- **WHEN** the owner generates a service into a folder that already exists
- **THEN** the generator refuses and leaves the existing folder unchanged
