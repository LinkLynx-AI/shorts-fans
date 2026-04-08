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
