# testtw

`testtw` is a minimal standalone test fixture project used for testing `krew` developer toolchain commands in an isolated environment.

## Purpose

- Tests project detection, shape detection, and config schema validation (`krewire.yaml`).
- Used by unit and integration tests (e.g. `internal/config/schema_test.go`) without affecting production workloads.
