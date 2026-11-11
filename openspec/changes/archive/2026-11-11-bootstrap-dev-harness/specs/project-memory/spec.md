## ADDED Requirements

### Requirement: Decisions are persisted in version control
Every accepted decision, preference, constraint, convention, assumption and open question SHALL be recorded in the decision registry. Every architectural decision SHALL also have an ADR. Conversation history MUST NOT be the only record of a decision.

#### Scenario: Owner states a new preference
- **WHEN** the owner states a preference, constraint, convention or correction
- **THEN** the registry gains or updates an entry with its source
- **AND** every file listed in the entry's `enforced_by` is updated in the same change

#### Scenario: Decision without ADR
- **WHEN** a registry decision references an ADR file that does not exist
- **THEN** repository validation fails and names the missing file

### Requirement: Instruction precedence
Instructions SHALL be ranked, highest first: the explicit current owner request, accepted OpenSpec specs, accepted ADRs and registry decisions, engineering standards, scoped `AGENTS.md`, then root `AGENTS.md` and repository defaults.

#### Scenario: Request conflicts with an accepted decision
- **WHEN** a request contradicts an accepted decision
- **THEN** the agent reports the conflict with the decision ID, the recorded text and a proposed resolution
- **AND** makes no change that depends on the conflict until the owner confirms

### Requirement: Controlled decision changes
An accepted decision SHALL change only through an OpenSpec change that adds a superseding ADR, marks the old ADR and registry entry `superseded` with a reference to the replacement, and applies the new decision to standards, templates, validation rules, agent instructions and affected services.

#### Scenario: Superseded decision without replacement
- **WHEN** a registry decision has status `superseded` but no existing `superseded_by` decision
- **THEN** repository validation fails

#### Scenario: Accepted ADR text edited in place
- **WHEN** a reviewer finds the decision text of an accepted ADR changed without a superseding ADR
- **THEN** the review is rejected

### Requirement: Language overrides are explicit
Backend services SHALL use the default backend language unless the registry lists the service path as an exception with a reason. Frontend apps SHALL use the frontend language.

#### Scenario: Undeclared non-default backend language
- **WHEN** the service catalog declares a backend service in a non-default language and the registry has no matching exception
- **THEN** repository validation fails and explains how to add a justified exception
