# Reviewer Profile

## Agent
- `reviewer_simple`

## Expected Coverage
- correctness
- security
- performance
- test quality
- coding rules
- UI impact checks when relevant

## Review Priorities
- 1. specification alignment with user intent, issue requirements, and repository contracts
- 2. regression risk across existing flows, state transitions, auth, permissions, empty states, and error paths
- 3. design integrity across FSD boundaries, dependency direction, duplication, and unnecessary abstraction
- 4. validation adequacy for the touched area
- 5. readability and maintainability

## Blocking Rule
- Block when at least one `P1` or higher finding has confidence `>= 0.65`.
- If `reviewer_simple` explicitly states that code or test changes are required before merge, treat that as blocking even when no severity label is emitted.

## Required Response Format
- The reviewer must return findings only.
- If findings exist, each finding must include:
  - severity
  - confidence
  - file path
  - rationale
- If no findings exist, the reviewer must state exactly:
  - `No findings. Gate clean.`
- If the reviewer output does not satisfy this format, treat it as:
  - `Reviewer output invalid for gate classification.`

## Exit Condition
- The loop ends only when no blocking finding remains or an explicit external blocker is documented.
