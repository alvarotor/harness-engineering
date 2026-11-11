# Service catalog

[`services.yaml`](services.yaml) is the single source for what deployables exist, who owns them, which
data they hold, and what they depend on. `tools/harness` validates it (R-CAT-*).

- Add the entry **in the OpenSpec change that approves the service**, as `status: planned`.
- Switch it to `active` in the change that lands the code.
- `owner` is `unassigned` until teams exist; set it when ownership is defined (it will feed `CODEOWNERS`).
- `personal_data` must list every personal-data field stored. It is the GDPR inventory (D-0012).
- `depends_on` lists runtime dependencies on other deployables (sync calls, JWKS, events consumed).

## Dependency overview (current)

No deployables yet (the product is not chosen, Q6). Add a Mermaid `flowchart` here with the first entries.
