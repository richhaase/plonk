---
steward_assessment: "3"
contract_path: "2026-08-09-repair-bats-homebrew-contract.md"
contract_id: "repair-bats-homebrew-build"
contract_revision: 1
contract_body_sha256: "9d9440219a7e208a48969a603fbe403976938028c40ad218a862ddf71663d1de"
change_identity: "sha256:4af136459259f47939d466a32d7ea60e8697b9ac647a4f1c9911f665eeeabfb3"
environment: "macOS 26.5.2 host; Docker client 29.7.2; Docker Engine 29.4.0 linux/arm64; Homebrew 6.0.15 in image"
assessor: "Codex"
created_at: "2026-08-09T18:56:54.608Z"
completed_at: "2026-08-09T18:57:22.852Z"
state: completed
assessment_body_sha256: "3f60f00ca98c7c6d8975377932c0e4a72a5ca331beebd1dd3ad1fed621201d13"
---
# Assessment: Repair BATS container build

## Provenance

- Immutable change: sha256:4af136459259f47939d466a32d7ea60e8697b9ac647a4f1c9911f665eeeabfb3
- Environment/context: macOS 26.5.2 host; Docker client 29.7.2; Docker Engine 29.4.0 linux/arm64; Homebrew 6.0.15 in image
- Assessor: Codex

## Overall

Outcome: pass

## Claim outcomes

### AC1: Building the project test container completes without the Homebrew untrusted-tap failure.
- Outcome: pass
- Evidence: E1
- Residual uncertainty: The build was exercised on Linux arm64; GitHub's hosted integration runner is Linux amd64.

### AC2: The built container provides BATS, bats-support, and bats-assert in the locations expected by the existing integration tests.
- Outcome: pass
- Evidence: E2
- Residual uncertainty: None observed.

### AC3: The container's existing integration-test entry point can execute the BATS suite.
- Outcome: pass
- Evidence: E3
- Residual uncertainty: None observed.

## Evidence log

### E1
- Command or artifact: `docker build --progress=plain -t plonk-test:latest .` against change `sha256:4af136459259f47939d466a32d7ea60e8697b9ac647a4f1c9911f665eeeabfb3`.
- Observation: The command exited 0. Homebrew trusted only the two named formulae, installed bats-support 0.3.0 and bats-assert 2.2.4, completed all 21 image stages, and tagged `plonk-test:latest`.

### E2
- Command or artifact: `docker run --rm --entrypoint bash plonk-test:latest -lc 'set -e; command -v bats; readlink -f /usr/local/lib/bats-support; readlink -f /usr/local/lib/bats-assert; test -f /usr/local/lib/bats-support/load.bash; test -f /usr/local/lib/bats-assert/load.bash; echo BATS_LIBRARIES_OK'`.
- Observation: The command exited 0, found BATS at `/home/linuxbrew/.linuxbrew/bin/bats`, resolved both `/usr/local/lib` links to installed Cellar directories, found both load files, and printed `BATS_LIBRARIES_OK`.

### E3
- Command or artifact: `docker run --rm plonk-test:latest all`.
- Observation: The command exited 0, verified every configured package manager, and completed the existing BATS suite with all 116 tests passing, including the suite's intentional skips.

## Contract observations

- The formula-scoped repair preserves Homebrew's trust boundary; no whole-tap trust or global trust bypass is present.

## Residual risks

- GitHub Actions will rebuild on Linux amd64 rather than the locally verified Linux arm64 platform. The repaired commands and formulae are platform-independent within Homebrew's Linux support, but that runner was not directly sampled here.

## Remediation

Classification: none
Next action: None.
