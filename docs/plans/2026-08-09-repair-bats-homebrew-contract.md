---
steward_contract: "3"
id: "repair-bats-homebrew-build"
title: "Repair BATS container build"
revision: 1
state: approved
created_at: "2026-08-09T18:03:19.944Z"
approved_at: "2026-08-09T18:16:00.962Z"
approved_by: "rdh"
frozen_body_sha256: "9d9440219a7e208a48969a603fbe403976938028c40ad218a862ddf71663d1de"
supersedes: null
---
# Repair BATS container build

## Outcome

The project test container builds successfully and can run its existing BATS integration suite again.

## Scope

### Preserve

- The existing BATS test entry point and support-library expectations remain compatible.

### Not in scope

- Unrelated dependency upgrades or CI workflow changes.

## Acceptance

- AC1: Building the project test container completes without the Homebrew untrusted-tap failure.
- AC2: The built container provides BATS, bats-support, and bats-assert in the locations expected by the existing integration tests.
- AC3: The container's existing integration-test entry point can execute the BATS suite.

## Constraints

- The repair must not globally disable or bypass package-source trust protections.
