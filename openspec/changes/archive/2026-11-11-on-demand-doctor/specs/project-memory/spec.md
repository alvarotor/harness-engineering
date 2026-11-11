## ADDED Requirements

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
