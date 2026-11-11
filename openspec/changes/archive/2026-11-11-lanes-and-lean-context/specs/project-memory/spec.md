## ADDED Requirements

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
