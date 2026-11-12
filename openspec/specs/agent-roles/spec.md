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
  for explicit owner approval without running an installer

#### Scenario: The owner has not approved an installation
- **WHEN** the owner asks to run checks or diagnostic bootstrap but has not approved installing a
  missing tool
- **THEN** the agent runs no installer or package manager

#### Scenario: The owner explicitly approves installation
- **WHEN** the owner approves installing a named missing tool
- **THEN** the agent may run only the approved installation action and then report the result; running diagnostic
  bootstrap remains non-installing

#### Scenario: An owner action is required
- **WHEN** the installation needs `sudo` or equivalent privilege, or bootstrap asks the owner to start Docker
- **THEN** the agent gives the owner the exact action or command and does not perform it

### Requirement: Developer-tool checks run only when requested
Agents MUST run `make doctor` only when the owner explicitly asks them to check developer-tool
availability. An unrelated request or a request to run another repository command MUST NOT trigger
an automatic `make doctor` check at session start.

#### Scenario: Owner asks for a developer-tool check
- **WHEN** the owner explicitly asks whether required developer tools are installed or asks the
  agent to run `make doctor`
- **THEN** the agent runs `make doctor` and reports its result

#### Scenario: A session starts without a tool-check request
- **WHEN** the owner starts a session with a request that does not ask to check developer-tool
  availability
- **THEN** the agent does not run `make doctor` automatically

#### Scenario: Owner asks for another repository command
- **WHEN** the owner asks the agent to run a repository command other than `make doctor`
- **THEN** the agent runs the requested command without first running `make doctor` automatically

### Requirement: Separated roles
The harness SHALL define exactly these roles: orchestrator, architect, planner, implementer, test-engineer, security-reviewer, reviewer, documentation and devops. Each role definition MUST state its purpose, responsibilities, allowed actions, required inputs, expected outputs, validation steps, handoff rules, stop conditions, and when human approval is required.

#### Scenario: Role definition missing a section
- **WHEN** a role definition lacks one of the required sections
- **THEN** repository validation fails and names the role and the missing section

#### Scenario: Unknown role added
- **WHEN** a role definition exists that is not in the defined role set
- **THEN** repository validation fails and asks for a decision change or a merge into an existing role

### Requirement: Orchestrated order of work
The orchestrator SHALL classify every request into one of the lanes defined in `AGENTS.md` §6 (S, M or L), choosing the higher lane when in doubt, and SHALL run that lane's steps in the order of the flow in `docs/agents/README.md`. The orchestrator MUST NOT edit files, and MUST reject a handoff that lacks evidence of the checks it claims.

#### Scenario: Handoff without evidence
- **WHEN** a role hands off claiming checks passed without including the commands and their output
- **THEN** the orchestrator returns the work to that role and does not advance the flow

#### Scenario: Repeated failure
- **WHEN** the same step fails twice with the same findings
- **THEN** the orchestrator escalates to the owner and stops looping

#### Scenario: Small fix
- **WHEN** the owner asks for a fix that changes no observable behavior and touches no lane-L path or topic
- **THEN** the orchestrator runs lane S: the implementer applies it and runs the validation command, with no OpenSpec change and no other roles

#### Scenario: Feature in an existing service
- **WHEN** the owner asks for a feature inside an existing service that needs no new decision
- **THEN** the orchestrator runs lane M: a light OpenSpec change, owner approval, implementation with tests and docs, review including documentation verification, archive

#### Scenario: Risky change
- **WHEN** a request matches a lane-L path or topic
- **THEN** the orchestrator runs planning, owner approval, implementation, independent testing, security review when triggered, review, documentation verification and owner approval of gated items before archive

#### Scenario: Lane too low
- **WHEN** a role finds that a lane S or M change touches a lane-L path or topic
- **THEN** the work returns to the orchestrator, which moves the change to the higher lane

### Requirement: Conditional security review
A change that matches a documented security review trigger SHALL be reviewed by the security-reviewer before the reviewer step, and SHALL require owner approval. Changes that match no trigger SHALL skip security review.

#### Scenario: Auth change
- **WHEN** a change modifies the identity service or token verification
- **THEN** the security-reviewer produces a report before the reviewer step

#### Scenario: Triggered change skipped security review
- **WHEN** the reviewer finds a triggered change without a security report
- **THEN** the review verdict is "changes requested"

### Requirement: Documentation ownership
Every role SHALL document its own work in the same step and list the documents it updated in its handoff. For each lane L change, the documentation role SHALL verify that every changed area has its required documentation, by comparing the change with the documents updated; for lane M the reviewer performs this verification. The documentation role SHALL fix descriptive gaps and report which role missed them. It MUST NOT create or change decisions, rules or agent instructions.

#### Scenario: Author documented the change
- **WHEN** a change's handoffs list every document the change requires
- **THEN** the documentation report marks each one "ok" and makes no edits

#### Scenario: Author missed a document
- **WHEN** a change adds an environment variable to a deployable and its README was not updated
- **THEN** the documentation role updates the README, and the report marks it "missing", naming the responsible role

#### Scenario: Gap needs a decision
- **WHEN** fixing a documentation gap would require a new decision or rule
- **THEN** the documentation role escalates to the architect through the orchestrator and makes no edit

#### Scenario: Lane M documentation check
- **WHEN** a lane M change misses a required document
- **THEN** the reviewer lists it as a finding and the verdict is "changes requested"

### Requirement: Tool adapters follow the canonical definition
Each role SHALL have a single canonical, tool-neutral definition. Every tool-specific adapter for a role MUST reference its canonical definition, and adapters MUST NOT redefine responsibilities.
Adapter permissions MUST enforce each role's write boundary for every tool that can write files, including the shell. No role MAY change agent adapters or agent tool configuration without explicit owner approval. The agent tool configuration MUST select the orchestrator as the default entry point.

#### Scenario: Adapter missing for a role
- **WHEN** a canonical role has no adapter for the configured agent tool
- **THEN** repository validation fails

#### Scenario: Read-only role runs a writing command
- **WHEN** a read-only role (orchestrator, reviewer, security-reviewer) runs a shell command outside its read-only allowlist
- **THEN** the tool asks the owner before running it

#### Scenario: Role edits its own permissions
- **WHEN** any role tries to edit an agent adapter or the agent tool configuration
- **THEN** the edit is denied, or for the architect it requires owner approval

#### Scenario: New session
- **WHEN** a session starts without an explicitly selected agent
- **THEN** the orchestrator is the active agent

### Requirement: Human approval gates
Agents SHALL stop and request owner approval before every gated item listed in `AGENTS.md` §7 (the single list of owner gates). Archiving a reviewed change SHALL NOT require owner approval.

#### Scenario: Implementer needs a new infrastructure component
- **WHEN** an implementer finds a task requires an infrastructure component not covered by an accepted decision
- **THEN** it stops, hands off through the orchestrator to the architect, and makes no change that introduces the component

#### Scenario: Reviewed change is archived
- **WHEN** a lane M or L change has passed review and the validation command
- **THEN** the planner archives it without waiting for owner approval
