# Reviewer Profile

## Agent
- `reviewer`

## Expected Coverage
- security
- correctness
- performance
- test quality
- coding rules

## Blocking Rule
- Block when the reviewer stack returns at least one `P1` or higher finding with confidence `>= 0.65`.

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
