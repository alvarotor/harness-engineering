## ADDED Requirements

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
