## MODIFIED Requirements

### Requirement: Orchestrated order of work
The orchestrator SHALL classify every request into lane S (no observable behavior change), M (feature inside an existing service, no new decision) or L (new service, decision, infrastructure or CI, contract break, destructive migration, or security trigger), choosing the higher lane when in doubt. The orchestrator MUST NOT edit files, and MUST reject a handoff that lacks evidence of the checks it claims.

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

### Requirement: Human approval gates
Agents SHALL stop and request owner approval before: starting implementation of a lane M or L change, accepting or superseding a decision, adding a dependency on a new infrastructure component or external service, changing security-sensitive behavior, deploying to production, running destructive data migrations, or lowering a quality gate. Archiving a reviewed change SHALL NOT require owner approval.

#### Scenario: Implementer needs a new infrastructure component
- **WHEN** an implementer finds a task requires an infrastructure component not covered by an accepted decision
- **THEN** it stops, hands off through the orchestrator to the architect, and makes no change that introduces the component

#### Scenario: Reviewed change is archived
- **WHEN** a lane M or L change has passed review and the validation command
- **THEN** the planner archives it without waiting for owner approval
