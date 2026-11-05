# Design

ADR-0019 supersedes ADR-0018 only for the agent's authority to execute bootstrap after a failed doctor check. The script remains the setup and repair mechanism and keeps its existing pinned-tool and Docker-check behavior.

The session-start instructions will describe the proposed action and pause for explicit approval. This proposal does not add a preview mode or change the bootstrap script. During apply, audit every path in D-0018's `enforced_by` list and every textual reference to the old automatic agent flow; preserve files whose behavior is purely technical and record that audit in the handoff.

## Owner gate

The owner approved this OpenSpec change. ADR-0018 is superseded, ADR-0019 is accepted, and the enforced startup flow has been updated. Archiving remains a separate owner-approved step.
