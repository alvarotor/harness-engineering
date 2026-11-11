## ADDED Requirements

### Requirement: Deferred choices are asked at their trigger
Product choices not yet made SHALL be recorded as open questions with a trigger (`needed_by`). Before work that reaches a trigger, the orchestrator MUST stop and ask the owner. The owner's answer MUST be recorded as a decision before the work continues, and the question MUST move to the history file.

#### Scenario: First product feature requested
- **WHEN** the owner requests the first product feature and the product question (Q6) is still open
- **THEN** the orchestrator runs the new-product checklist with the owner before any planning starts

#### Scenario: First event between services with messaging open
- **WHEN** a change would publish the first event from one service to another and the messaging question (Q7) is still open
- **THEN** work stops until the owner chooses the messaging option

#### Scenario: Owner answers a deferred question
- **WHEN** the owner answers a deferred question
- **THEN** the architect records a decision in the registry, and the question becomes resolved and moves to the history file
