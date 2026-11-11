# delivery Specification

## Purpose
Defines how changes are checked, built once and promoted from dev to prod: pull request checks, merge-to-dev deploys, tag-based prod releases with owner approval, pinned host keys and pinned workflow dependencies.

## Requirements

### Requirement: Pull requests are checked and scanned
Every pull request SHALL run repository validation and the Compose merge check. For each changed deployable, the workflow MUST run that deployable's `lint`, `test` and `image` make targets and a scan of the built image. The pull request MUST also run a secret scan of the git history and a vulnerability check of the Go code. A failure of any blocking check MUST make the pull request check fail. Third-party image vulnerabilities MUST be reported but MUST NOT block.

#### Scenario: Changed deployable
- **WHEN** a pull request changes files under a deployable
- **THEN** the workflow runs that deployable's lint, test and image targets and scans its image

#### Scenario: Documentation-only change
- **WHEN** a pull request changes no deployable
- **THEN** the workflow runs repository validation and the scans, and no deployable target

#### Scenario: Secret committed
- **WHEN** a pull request adds a secret to the repository history
- **THEN** the secret scan fails the pull request check

#### Scenario: Vulnerable Go dependency
- **WHEN** a Go module reachable from the code has a known vulnerability
- **THEN** the vulnerability check fails the pull request check

#### Scenario: Third-party image with a known vulnerability
- **WHEN** a third-party image pinned in Compose has a known vulnerability
- **THEN** the report shows it and the pull request check is not blocked

### Requirement: Merges publish once and deploy to dev
A merge to main SHALL build every deployable image once, scan it, and publish it tagged with the git commit SHA. An image that fails the scan MUST NOT be published. Publishing MUST be the only job that holds package write permission, and it MUST run only from the main workflow. After publishing, the dev environment MUST be deployed with that commit SHA. While the dev environment has no deploy host configured, the deploy steps MUST be skipped and the workflow MUST succeed.

#### Scenario: Merge with dev host configured
- **WHEN** a commit is merged to main and the dev host is configured
- **THEN** each image is built once, scanned, published tagged with the commit SHA, and dev runs those images

#### Scenario: Image fails the scan
- **WHEN** a built image has a fixable high or critical vulnerability
- **THEN** the image is not published and dev is not deployed from that commit

#### Scenario: Dev host not configured
- **WHEN** a commit is merged to main and the dev environment has no deploy host configured
- **THEN** the images are still published, the deploy steps are skipped, and the workflow succeeds

### Requirement: Tags release to prod after approval
A tag `vX.Y.Z` on a main commit SHALL deploy to prod only after owner approval through the protected prod environment, and the release MUST deploy the image digests published for that commit without rebuilding them. Rollback MUST be done by re-running the release for an earlier tag.

#### Scenario: Tag on a main commit
- **WHEN** a `vX.Y.Z` tag points to a commit on main
- **THEN** the release waits for owner approval, then deploys the digests published for that commit

#### Scenario: Prod not configured
- **WHEN** a tag on main is released and the prod environment has no deploy host configured
- **THEN** the approval prompt is shown, and after approval the deploy steps are skipped and the workflow succeeds

#### Scenario: Rollback
- **WHEN** the owner needs to return prod to an earlier release
- **THEN** the owner re-runs the release for the earlier tag, which deploys that tag's commit digests after approval

### Requirement: Release checks run before approval and fail closed
A tag that is not on main SHALL fail before any approval is requested and deploy nothing. After approval, the digest lookup MUST fail closed when an image or its digest is missing, before any file is written or the host is contacted.

#### Scenario: Tag not on main
- **WHEN** a `vX.Y.Z` tag points to a commit that is not on main
- **THEN** the release fails before the approval prompt and deploys nothing

#### Scenario: Image missing after approval
- **WHEN** the approved release finds that an image for the commit is not published
- **THEN** the deploy fails before it writes the release env file or contacts the host

### Requirement: Deploy host keys are pinned
Once a deploy host is configured, the deploy SHALL verify the host's SSH key against the known-hosts entries configured in the environment (`DEPLOY_KNOWN_HOSTS`) with strict host key checking. It MUST NOT trust a key on first use. When the host key is not configured, the deploy MUST fail without sending any secret or file to the host.

#### Scenario: Host key is not configured
- **WHEN** a deploy runs with a deploy host but without `DEPLOY_KNOWN_HOSTS`
- **THEN** the deploy fails before any secret or file is sent, and names the missing variable

#### Scenario: Host key does not match
- **WHEN** the host presents a key that does not match the configured known-hosts entry
- **THEN** the deploy fails and sends nothing further to the host

### Requirement: Deploy secrets are files on the host
The deploy SHALL write each deploy secret to its own file with mode `0600` under the deploy user's home on the host, and the release env file SHALL hold only image digests and non-secret values. Secret values MUST NOT be passed on the command line or written into the release env file. The secret files and the runner copy MUST be removed after the deploy step.

#### Scenario: Secret handed to the deploy
- **WHEN** the deploy receives a secret entry
- **THEN** the value is written to its own `0600` file under the host secrets directory (`~/har/secrets`)

#### Scenario: Runner cleanup
- **WHEN** the deploy step ends, successfully or not
- **THEN** the runner's copies of the SSH key and the release env file are removed

### Requirement: Workflow dependencies are pinned and permissions are declared
Every action referenced by a workflow, except local reusable workflows in the same repository, SHALL be pinned to a full commit SHA. Every workflow SHALL declare `permissions:`. No `run:` script MUST interpolate a `${{ }}` expression; values reach scripts through `env:`. Repository validation MUST fail when an action is not pinned (R-CI-001) or when a workflow lacks `permissions:` or interpolates an expression in `run:` (R-CI-002).

#### Scenario: Action referenced by tag
- **WHEN** a workflow references an action by a version tag instead of a full commit SHA
- **THEN** repository validation fails and names the workflow and the action

#### Scenario: Workflow without permissions
- **WHEN** a workflow file does not declare `permissions:`
- **THEN** repository validation fails and names the workflow

#### Scenario: Expression in a run script
- **WHEN** a `run:` script contains `${{ }}`
- **THEN** repository validation fails and names the workflow and the step
