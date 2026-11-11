## MODIFIED Requirements

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
