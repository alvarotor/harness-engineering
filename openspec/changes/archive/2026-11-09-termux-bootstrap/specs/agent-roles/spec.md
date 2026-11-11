## MODIFIED Requirements

### Requirement: Owner approval before agent bootstrap
When an agent's initial `make doctor` check fails, the agent MUST explain that it proposes to run
`./scripts/bootstrap.sh`, summarize the actions indicated by the check output and the script, and
stop for explicit owner approval before running it. A direct owner request to run the script counts
as approval for that run. Approval to run the script MUST NOT be treated as approval for separate
actions that require owner action, such as starting Docker or using `sudo`.

#### Scenario: Doctor reports missing tools
- **WHEN** the initial doctor check fails and the bootstrap script may install or configure tools
- **THEN** the agent summarizes those actions and waits for owner approval without running the
  script

#### Scenario: Owner approves bootstrap
- **WHEN** the owner explicitly approves the proposed bootstrap run or directly asks the agent to
  run the script
- **THEN** the agent may run `./scripts/bootstrap.sh`

#### Scenario: Bootstrap reports a separate owner action
- **WHEN** the script reports a step requiring `sudo` or asks the owner to start Docker
- **THEN** the agent stops and gives the owner the exact action or command to perform
