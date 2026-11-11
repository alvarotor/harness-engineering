# Design: preference metadata in registry validation

## Context

The registry memory protocol requires durable preferences to record rationale and their current or
planned enforcement paths. The registry validator currently accepts these fields only for decisions,
so strict YAML decoding rejects a conforming preference entry and R-REG-004 does not check preference
enforcement paths.

## Approach

- Extend the harness `Preference` representation to decode `rationale`, `enforced_by`, and
  `planned_enforced_by` as optional fields, preserving compatibility with existing entries.
- Keep the existing R-REG-004 meaning: each path listed in `enforced_by` must exist. Apply that
  check to both decisions and preferences; `planned_enforced_by` remains unchecked until its paths
  are implemented.
- Add unit coverage for preference metadata decoding and for valid and missing preference
  enforcement paths. Keep the validation rule documented in `docs/standards/validation.md` aligned
  with the behavior.

No accepted architectural decision changes, new dependencies, or service behavior are involved.
