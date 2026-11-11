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
