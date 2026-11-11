## ADDED Requirements

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
