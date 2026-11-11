# developer-bootstrap Specification

## Purpose
Defines how repository bootstrap and tooling checks acquire GNU Make and handle supported host prerequisites.

## Requirements

### Requirement: Native Make setup on supported hosts
Bootstrap MUST check for GNU Make and report the host's installation guidance when it is missing.
It MUST NOT install Make or invoke a package manager. It MUST NOT install, recognize, or invoke a
private executable named `make-proj`.

#### Scenario: Make is already installed
- **WHEN** bootstrap runs on macOS, Linux, or Termux and `make` is available
- **THEN** it reports Make as present and uses `make` for repository checks

#### Scenario: Make is missing on macOS or Linux
- **WHEN** bootstrap runs on macOS or Linux and `make` is unavailable
- **THEN** it reports the appropriate OS installation action and does not install or invoke a package manager

#### Scenario: Make is missing in Termux
- **WHEN** bootstrap runs in Termux and `make` is unavailable
- **THEN** it reports `pkg install make` for the owner without invoking `pkg`

### Requirement: Termux host detection and Make setup
Bootstrap MUST identify Termux before generic Linux and MUST report Termux package guidance when
Make is missing, without invoking `pkg`.

#### Scenario: Make is missing in Termux
- **WHEN** bootstrap detects Termux and `make` is unavailable
- **THEN** it reports `pkg install make` for the owner and makes no package-manager call

#### Scenario: Termux is detected on an Android Linux kernel
- **WHEN** the host reports Linux and Termux markers are present
- **THEN** bootstrap follows the Termux path and does not use generic Linux package-manager guidance

### Requirement: Use the OS Make executable
Repository instructions and bootstrap verification MUST use `make` directly and MUST NOT require
`mise` or `make-proj` as an alternative command.

#### Scenario: Agent checks repository tooling
- **WHEN** an agent follows repository tool-check instructions and `mise` is absent
- **THEN** it runs `make doctor`, which checks required commands directly without invoking mise

### Requirement: Presence-only developer tool checks
`make doctor` MUST check whether each documented developer-tool command exists on `PATH` and MUST
NOT compare its version to a repository pin.

#### Scenario: All required commands are present
- **WHEN** each required command is available on `PATH`
- **THEN** `make doctor` reports them present without checking their versions

#### Scenario: A required command is missing
- **WHEN** one or more required commands are unavailable on `PATH`
- **THEN** `make doctor` names each missing command and exits unsuccessfully

### Requirement: Diagnostic-only bootstrap
The bootstrap script MUST report missing developer commands and applicable OS installation guidance.
It MUST NOT invoke package managers, install tools, or download binaries.

#### Scenario: A tool is missing
- **WHEN** bootstrap detects that a required command is unavailable
- **THEN** it reports the command and installation guidance without attempting installation

#### Scenario: No tools are missing
- **WHEN** all required commands are available
- **THEN** bootstrap reports that the developer environment is ready without running installers

### Requirement: Prevent automatic Go toolchain downloads
Repository Make targets that invoke Go MUST set `GOTOOLCHAIN=local` and MUST NOT download or install
a Go toolchain. The Go compatibility floor declared by `go.mod` remains in force.

#### Scenario: The local Go toolchain is below the module floor
- **WHEN** a repository target invokes Go and the installed toolchain is below the `go.mod` floor
- **THEN** the command reports that the local toolchain is insufficient and does not download another toolchain

#### Scenario: Go is missing
- **WHEN** a repository target requires Go but no Go executable is installed
- **THEN** the command reports the missing prerequisite without attempting a download
