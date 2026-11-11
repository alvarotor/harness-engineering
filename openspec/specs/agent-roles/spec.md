# agent-roles Specification

## Purpose
Defines agent responsibilities and owner-approval gates for repository setup and developer tool installation.

## Requirements

### Requirement: Owner approval before agent bootstrap
When repository checks report missing developer tools, an agent MUST list the missing commands and
relevant installation guidance, offer to install them, and wait for explicit owner approval before
running any installation command. Running `make doctor` or the diagnostic bootstrap MUST NOT install
tools. Approval to run checks is not approval to install. Commands requiring `sudo` remain the
owner's responsibility.

#### Scenario: A developer tool is missing
- **WHEN** `make doctor` or bootstrap reports a missing command
- **THEN** the agent lists the command, gives installation guidance, offers to install it, and waits
  for explicit owner approval

#### Scenario: Doctor reports missing tools
- **WHEN** the initial doctor check fails because a required command is missing
- **THEN** the agent summarizes the missing commands and installation guidance, offers to install
  them, and waits for explicit owner approval without running an installer

#### Scenario: The owner has not approved an installation
- **WHEN** the owner asks to run checks or diagnostic bootstrap but has not approved installing a
  missing tool
- **THEN** the agent runs no installer or package manager

#### Scenario: The owner explicitly approves installation
- **WHEN** the owner approves installing a named missing tool
- **THEN** the agent may run only the approved installation action and then report the result

#### Scenario: Owner approves bootstrap
- **WHEN** the owner explicitly asks to install a named missing tool
- **THEN** the agent may run only that installation action; running diagnostic bootstrap remains
  non-installing

#### Scenario: An installation requires elevated privileges
- **WHEN** the required installation command needs `sudo` or equivalent privilege
- **THEN** the agent gives the exact command to the owner and does not run it

#### Scenario: Bootstrap reports a separate owner action
- **WHEN** bootstrap reports an action requiring `sudo` or asks the owner to start Docker
- **THEN** the agent gives the owner the exact action or command and does not perform it
