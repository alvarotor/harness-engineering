# project-memory Specification

## Purpose
Defines validation rules for durable decisions, preferences, and their supersession metadata in the project registry.

## Requirements

### Requirement: Preference supersession metadata
The decision registry validator MUST accept optional lifecycle `status` and `superseded_by` fields
on preferences. A superseded preference MUST identify an existing preference as its successor.

#### Scenario: A preference is superseded
- **WHEN** a preference has `status: superseded` and `superseded_by` names an existing preference
- **THEN** registry validation accepts the preference lifecycle metadata

#### Scenario: A superseded preference has no valid successor
- **WHEN** a preference is marked superseded but its `superseded_by` is missing or does not name an
  existing preference
- **THEN** registry validation reports a violation

#### Scenario: Current preference omits lifecycle metadata
- **WHEN** a preference has no `status` or `superseded_by` fields
- **THEN** registry validation treats it as current and accepts the entry

### Requirement: Preference enforcement metadata is validated
The decision registry validator MUST accept optional `rationale`, `enforced_by`, and
`planned_enforced_by` fields on preferences. Every path listed in a preference's `enforced_by` MUST
exist. Paths in `planned_enforced_by` are not required to exist yet.

#### Scenario: A preference records current and planned enforcement
- **WHEN** a preference includes a rationale, existing `enforced_by` paths, and planned paths
- **THEN** registry validation accepts the metadata

#### Scenario: A preference names a missing current enforcement path
- **WHEN** a preference's `enforced_by` list contains a path that does not exist
- **THEN** registry validation reports an R-REG-004 violation for that path

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
Instructions SHALL be ranked in the precedence order defined in `AGENTS.md` §2, with the explicit current owner request highest.

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

### Requirement: Current and historical memory are separate
The decision registry SHALL hold only current entries. Superseded or deprecated decisions and preferences, resolved questions and the corrections log SHALL live in a separate history file that agents read only when changing a decision. Validation MUST check both files together and MUST reject current entries in the history file.

#### Scenario: Superseded entry in history
- **WHEN** a superseded preference lives in the history file and its replacement lives in the registry
- **THEN** registry validation accepts both

#### Scenario: Current entry in history
- **WHEN** the history file contains an accepted decision, a current preference, an assumption or an open question
- **THEN** validation fails and asks to move it back to the registry

#### Scenario: History file does not parse
- **WHEN** the history file is not valid
- **THEN** registry validation reports a schema violation

### Requirement: Deferred choices are asked at their trigger
Product choices not yet made SHALL be recorded as open questions with a trigger (`needed_by`). Before work that reaches a trigger, the orchestrator MUST stop and ask the owner. The owner's answer MUST be recorded as a decision before the work continues, and the question MUST move to the history file.

#### Scenario: First product feature requested
- **WHEN** the owner requests the first product feature and the product question (Q1) is still open
- **THEN** the orchestrator runs the new-product checklist with the owner before any planning starts

#### Scenario: First event between services with messaging open
- **WHEN** a change would publish the first event from one service to another and the messaging question (Q4) is still open
- **THEN** work stops until the owner chooses the messaging option

#### Scenario: Owner answers a deferred question
- **WHEN** the owner answers a deferred question
- **THEN** the architect records a decision in the registry, and the question becomes resolved and moves to the history file
