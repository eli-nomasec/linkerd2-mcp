# Project Guidelines

## General guidelines

-   ALWAYS Read the relevant markdown files mentioned in these rules before taking action
-   When the user asks "update memory bank" - review all markdown files mentioned here

## Documentation Requirements

-   Update relevant documentation in /docs when modifying features
-   Keep README.md in sync with new capabilities
-   Maintain changelog entries in CHANGELOG.md

## Project goals and overview

Create and maintain high-level project goals and overview in /docs/project.md

## Architecture Decision Records

Create ADRs and follow them in /docs/architecture.md for:

-   Major dependency changes
-   Architectural pattern changes
-   New integration patterns
-   Database schema changes

## Tech stack and choices

Create and maintain tech stack decisions in /docs/techStack.md

## Tooling (building and running things)

Create and maintain knowledge required for running or building things in docs/tooling.md

## Progress

- Follow and maintain progress in /docs/progress.md

## Code Style & Patterns

-   Keep protobuf definitions in /proto/
-   Use buf for protobuf generation
-   Place generated code in /gen/
-   Use repository pattern for data access
-   Only panic (if needed) in main.go files

## Testing Standards

-   Unit tests required for business logic
-   Integration tests for API endpoints
-   E2E tests for critical user flows - use kubernetes for that
-   Never run anything on a k8s context other than colima-arm
-   All tests should be go tests (even e2e tests)