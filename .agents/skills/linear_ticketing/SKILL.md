---
name: linear-ticketing
description: "Create and structure Linear work for this repository by first discussing detailed functional and non-functional requirements, then choosing the right Linear artifact shape: a single issue, a parent issue with sub-issues, or a project with milestones and issues. Use when users want to plan work in Linear, batch-create a project and its child issues after requirements alignment, or turn product requests into actionable tickets without deciding concrete implementation methods."
---

# linear-ticketing

## Goal
- Turn a product request into the right Linear structure after thorough requirement discovery.
- Decide whether the work should be tracked as a single issue, a parent issue with sub-issues, or a project with milestones and issues.
- Capture functional requirements, non-functional requirements, business rules, risks, constraints, and explicit non-goals.
- Do not implement code in this skill.
- Do not decide concrete implementation methods in this skill.

## Required Input
- Feature, problem, or initiative summary.
- Expected user or business outcome.

## Optional Input
- Team, project, label, or priority policies in Linear.
- Delivery expectations such as urgency, target date, or stakeholder visibility needs.
- Existing related issues, projects, or dependencies.

## Core Operating Rules
- Start with requirement discovery, not ticket creation.
- Ask focused follow-up questions until the material scope is stable enough to ticket.
- Challenge vague, contradictory, or underspecified requests instead of silently filling gaps.
- Discuss detailed functionality and non-functional requirements thoroughly with the user.
- Treat implementation approach as out of scope unless the user explicitly says a technical constraint is already fixed and relevant to scope.
- If important uncertainty remains, either record it as an open question or create discovery issues first.

## What to Discuss
- Primary user or actor and the target user journey.
- In-scope user-visible behavior and excluded behavior.
- Business rules, permissions, and policy decisions.
- Edge cases, failure cases, and abuse or moderation concerns.
- Non-functional requirements such as performance, reliability, observability, security, privacy, accessibility, and analytics when relevant.
- Dependencies on other teams, approvals, external services, or prior decisions.
- Success conditions and concrete acceptance criteria at the behavior level.

## What Not to Discuss Here
- API shape unless externally fixed and necessary to define scope.
- Database schema, internal architecture, or file structure.
- Concrete library, framework, or service choices unless already mandated.
- Refactor plans, class design, or exact implementation sequencing inside an issue.

## Linear Artifact Decision Model
- Create a single issue when the work is small, bounded, and independently deliverable.
- Create a parent issue with sub-issues when the work is larger than one issue but does not need project-level progress tracking.
- Create a project with milestones and issues when the work has a clear outcome, spans multiple phases or ownership slices, or needs stakeholder-facing status tracking.
- Do not create a project by default.
- If the user explicitly wants a project, support it, but say when a lighter structure would also be sufficient.
- Prefer decomposition by user outcome, workflow slice, policy boundary, or risk boundary.
- Do not default to backend/frontend/infra decomposition unless the user explicitly wants that framing.
- Keep each child small enough for one implementation delivery whenever possible.

## Creation Gate
Do not create tickets until all of the following are true.

- The target outcome is clear.
- The chosen Linear artifact shape is justified.
- The main functional scope is settled.
- Important non-functional requirements or constraints are either settled or explicitly listed as open questions.
- The split between child issues is understandable without implementation details.

## Output Contract
Create only the artifact shape that matches the decision model.

### Single Issue Required Sections
1. Context
2. Target outcome
3. In scope
4. Out of scope
5. Acceptance Criteria
6. Non-functional requirements or constraints
7. Dependencies and open questions

### Parent Issue Required Sections
1. Context
2. Target outcome
3. In scope
4. Out of scope
5. Functional requirements
6. Non-functional requirements
7. Risks, abuse, and policy considerations
8. Open questions
9. Child issue map with order and dependencies

### Project Required Sections
1. Summary
2. Target outcome
3. Why this needs a project
4. In scope
5. Out of scope
6. Success signals
7. Constraints and risks
8. Open questions
9. Milestone plan
10. Linked issue plan with order and dependencies

### Child Issue Required Sections
- Parent or project reference
- Outcome
- In scope
- Out of scope
- Acceptance Criteria
- Non-functional requirements or constraints
- Dependencies
- Open questions

Template text is available at `assets/templates.md`.

## Linear Write Path
Pick one method.

1. Linear MCP preferred for local Codex execution.
- Use existing Linear MCP configuration when available.
- Create the top-level artifact first, then attach milestones if needed, then create child issues, then wire dependencies and references.

2. If MCP is unavailable.
- Output markdown drafts for the chosen artifact shape so the user can copy them manually.

## Final Check Before Writing
- The ticket structure matches the size and uncertainty of the work.
- The created artifacts reflect requirements, not implementation guesses.
- Scope boundaries are clear.
- Dependencies are explicit.
- Non-functional requirements are included only when they matter.
- Open questions remain visible instead of being buried or guessed away.
