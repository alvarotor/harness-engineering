## ADDED Requirements

### Requirement: Separated roles
The harness SHALL define exactly these roles: orchestrator, architect, planner, implementer, test-engineer, security-reviewer, reviewer, documentation and devops. Each role definition MUST state its purpose, responsibilities, allowed actions, required inputs, expected outputs, validation steps, handoff rules, stop conditions, and when human approval is required.

#### Scenario: Role definition missing a section
- **WHEN** a role definition lacks one of the required sections
- **THEN** repository validation fails and names the role and the missing section

#### Scenario: Unknown role added
- **WHEN** a role definition exists that is not in the defined role set
- **THEN** repository validation fails and asks for a decision change or a merge into an existing role

### Requirement: Orchestrated order of work
Non-trivial work SHALL proceed in this order: planning, owner approval, implementation, independent testing, security review when triggered, review, documentation verification, owner approval of gated items, archive. The orchestrator MUST NOT edit files, and MUST reject a handoff that lacks evidence of the checks it claims.

#### Scenario: Handoff without evidence
- **WHEN** a role hands off claiming checks passed without including the commands and their output
- **THEN** the orchestrator returns the work to that role and does not advance the flow

#### Scenario: Repeated failure
- **WHEN** the same step fails twice with the same findings
- **THEN** the orchestrator escalates to the owner and stops looping

### Requirement: Conditional security review
A change that matches a documented security review trigger SHALL be reviewed by the security-reviewer before the reviewer step, and SHALL require owner approval. Changes that match no trigger SHALL skip security review.

#### Scenario: Auth change
- **WHEN** a change modifies the identity service or token verification
- **THEN** the security-reviewer produces a report before the reviewer step

#### Scenario: Triggered change skipped security review
- **WHEN** the reviewer finds a triggered change without a security report
- **THEN** the review verdict is "changes requested"

### Requirement: Documentation ownership
Every role SHALL document its own work in the same step and list the documents it updated in its handoff. The documentation role SHALL verify, for each change, that every changed area has its required documentation, by comparing the change with the documents updated. It SHALL fix descriptive gaps and report which role missed them. It MUST NOT create or change decisions, rules or agent instructions.

#### Scenario: Author documented the change
- **WHEN** a change's handoffs list every document the change requires
- **THEN** the documentation report marks each one "ok" and makes no edits

#### Scenario: Author missed a document
- **WHEN** a change adds an environment variable to a deployable and its README was not updated
- **THEN** the documentation role updates the README, and the report marks it "missing", naming the responsible role

#### Scenario: Gap needs a decision
- **WHEN** fixing a documentation gap would require a new decision or rule
- **THEN** the documentation role escalates to the architect through the orchestrator and makes no edit

### Requirement: Tool adapters follow the canonical definition
Each role SHALL have a single canonical, tool-neutral definition. Every tool-specific adapter for a role MUST reference its canonical definition, and adapters MUST NOT redefine responsibilities.

#### Scenario: Adapter missing for a role
- **WHEN** a canonical role has no adapter for the configured agent tool
- **THEN** repository validation fails

### Requirement: Human approval gates
Agents SHALL stop and request owner approval before: accepting or superseding a decision, adding a dependency on a new infrastructure component or external service, changing security-sensitive behavior, deploying to production, running destructive data migrations, lowering a quality gate, or archiving a change.

#### Scenario: Implementer needs a new infrastructure component
- **WHEN** an implementer finds a task requires an infrastructure component not covered by an accepted decision
- **THEN** it stops, hands off through the orchestrator to the architect, and makes no change that introduces the component
