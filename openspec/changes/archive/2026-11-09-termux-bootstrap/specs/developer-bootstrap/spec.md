## ADDED Requirements

### Requirement: Native Make setup on supported hosts
The bootstrap process MUST use the host's package or developer-tools mechanism to make GNU Make
available. It MUST NOT install, recognize, or invoke a private executable named `make-proj`.

#### Scenario: Make is already installed
- **WHEN** bootstrap runs on macOS, Linux, or Termux and GNU Make is available as `make`
- **THEN** bootstrap uses `make` for the repository's Make targets

#### Scenario: Make is missing on macOS or Linux
- **WHEN** bootstrap runs on macOS or Linux without `make`
- **THEN** it uses the supported native installation path or reports the exact host installation
  command, without creating a private Make executable

### Requirement: Termux host detection and Make setup
The bootstrap process MUST identify Termux before applying generic Linux setup and MUST use
Termux's package manager to install GNU Make when `make` is missing.

#### Scenario: Make is missing in Termux
- **WHEN** bootstrap detects Termux and `make` is unavailable
- **THEN** it runs `pkg install make` when `pkg` is available, or reports that exact command for
  the owner to run

#### Scenario: Termux is detected on an Android Linux kernel
- **WHEN** the host reports Linux and the Termux environment is present
- **THEN** bootstrap follows the Termux path and does not use the generic Linux package-manager
  path

### Requirement: Termux mise setup guidance
When Termux is detected and mise is unavailable, the bootstrap process MUST NOT download or run a
generic Linux mise binary. It MUST report `pkg install mise` for the owner to run and stop before
attempting pinned tool installation.

#### Scenario: mise is missing in Termux
- **WHEN** bootstrap detects Termux and cannot find mise
- **THEN** it reports `pkg install mise`, performs no generic Linux mise download, and stops before
  `mise install`

#### Scenario: Android toolchain prerequisites are unavailable
- **WHEN** repository tooling is checked in Termux and pinned tools or Docker Compose are missing
- **THEN** `make doctor` reports the missing prerequisites, and documentation makes clear that
  Termux Make/bootstrap support does not guarantee the full repository check or Docker workflow

### Requirement: Use the OS Make executable
Repository instructions and bootstrap verification MUST use `make` directly and MUST NOT require
`make-proj` as an alternative command.

#### Scenario: Agent checks repository tooling
- **WHEN** an agent follows the repository's tool-check instructions
- **THEN** it runs `make doctor` and follows the existing approval gate before running bootstrap
