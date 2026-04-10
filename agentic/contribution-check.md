---
name: "Contribution Check"
on:
  schedule: "every 4 hours"
  workflow_dispatch:

permissions:
  contents: read
  issues: read
  pull-requests: read

env:
  TARGET_REPOSITORY: ${{ vars.TARGET_REPOSITORY || github.repository }}

tools:
  github:
    toolsets: [default]
    lockdown: false
    min-integrity: none # This workflow is allowed to examine and comment on any issues

safe-outputs:
  create-issue:
    title-prefix: "[Contribution Check Report]"
    labels:
      - contribution-report
    close-older-issues: true
  add-labels:
    allowed: [spam, needs-work, outdated, lgtm]
    max: 4
    target: "*"
    target-repo: ${{ vars.TARGET_REPOSITORY }}
  add-comment:
    max: 10
    target: "*"
    target-repo: ${{ vars.TARGET_REPOSITORY }}
    hide-older-comments: true
---

## Target Repository

The target repository is `${{ env.TARGET_REPOSITORY }}`. All PR fetching and subagent dispatch use this value.

## Overview

You are an **orchestrator**. Your job is to dispatch PRs to the `contribution-checker` subagent for evaluation and compile the results into a single report issue in THIS repository (`${{ github.repository }}`).

You do NOT evaluate PRs yourself. You delegate each evaluation to `.github/agents/contribution-checker.agent.md`.

## Pre-filtered PR List

A `pre-agent` step has already queried and filtered PRs from `${{ env.TARGET_REPOSITORY }}`. The results are in `pr-filter-results.json` at the workspace root. Read this file first. It contains:

```json
{
  "pr_numbers": [18744, 18743, 18742],
  "skipped_count": 10,
  "evaluated_count": 3
}
```

If `pr_numbers` is empty, create a report stating no PRs matched the filters and skip dispatch.

## Step 1: Dispatch to Subagent

For each PR number in the comma-separated list, delegate evaluation to the **contribution-checker** subagent (`.github/agents/contribution-checker.agent.md`).

### How to dispatch

Call the contribution-checker subagent for each PR with this prompt:

```
Evaluate PR ${{ env.TARGET_REPOSITORY }}#<number> against the contribution guidelines.
```

The subagent accepts any `owner/repo#number` reference  -  the target repo is not hardcoded.

The subagent will return a single JSON object with the verdict and a comment for the contributor.

### Parallelism

- Dispatch **multiple PRs concurrently** when possible  -  the subagent evaluations are independent of each other.
- Each subagent call is stateless and self-contained. It fetches its own PR data.

### Collecting results

Gather all returned JSON objects. If a subagent call fails, record the PR with verdict `❓` and quality `triage:error` in the report.

### Posting comments

For each PR where the subagent returned a non-empty `comment` field and the quality is NOT `lgtm`, call the `add_comment` safe output tool to post the comment to the PR in the target repository. Pass the PR number and the comment body from the subagent result. The `add_comment` tool is pre-configured with `target-repo` pointing to the target repository  -  you do NOT need to specify the repo yourself.

Do NOT post comments to PRs with `lgtm` quality  -  those are ready for maintainer review and don't need additional feedback.

## Step 2: Compile Report

Create a single issue in THIS repository. Use the `skipped_count` from `pr-filter-results.json`. Build the report tables from the JSON objects returned by the subagent (use `number`, `title`, `author`, `lines`, and `quality` fields).

Follow the **report layout rules** below  -  they apply to every report this workflow produces.

### Report Layout Rules

Apply these principles to make the report scannable, warm, and actionable:

1. **Lead with the takeaway.** Open with a single-sentence human-readable summary that tells the maintainer what happened and what needs attention. No jargon, no counts-only headers. Example: *"We looked at 10 new PRs  -  6 look great, 3 need a closer look, and 1 doesn't fit the project guidelines."*

2. **Group by action, not by data.** Organize results into clear groups that answer "what should I do?" rather than listing raw rows. Use these groups (omit any group with zero items):
   - **Ready to review** 🟢  -  PRs that passed all checks
   - **Needs a closer look** 🟡⚠️  -  PRs that need discussion or focus work
   - **Off-guidelines** 🔴  -  PRs that don't align with CONTRIBUTING.md

3. **One table per group.** Keep tables short and focused. Columns:
   - PR (linked), Title (truncated to ~50 chars), Author, Lines changed, Quality signal
   - Do NOT include boolean checklist columns (on-topic, focused, deps, tests)  -  those are for the subagent, not the reader. The verdict emoji and quality signal are enough.

4. **Use whitespace generously.** Separate groups with blank lines and horizontal rules (`---`). Let each section breathe.

5. **End with context, not noise.** Close with a small stats line: `Evaluated: {n} · Skipped: {n} · Run: {run_link}`. Keep it quiet  -  one line, not a table.

6. **Tone: warm and constructive.** These reports help maintainers prioritize, not gatekeep. Use encouraging language for aligned PRs ("looking good", "ready for eyes"). Be matter-of-fact for off-guidelines PRs  -  no shaming.

### Example Report

```markdown
## Contribution Check  -  {date}

We looked at 4 new PRs  -  1 looks great, 2 need a closer look, and 1 doesn't fit the contribution guidelines.

---

### Ready to review 🟢

| PR | Title | Author | Lines | Quality |
|----|-------|--------|------:|---------|
| #4521 | Fix CLI flag parsing for unicode args | @alice | 125 | lgtm ✨ |

---

### Needs a closer look 🟡

| PR | Title | Author | Lines | Quality |
|----|-------|--------|------:|---------|
| #4515 | Refactor auth + add rate limiting | @bob | 310 | needs-work |
| #4510 | Add Redis caching layer | @carol | 88 | needs-work |

---

### Off-guidelines 🔴

| PR | Title | Author | Lines | Quality |
|----|-------|--------|------:|---------|
| #4519 | Add unrelated marketing page | @dave | 42 | spam |

---

Evaluated: 4 · Skipped: 10
```

## Step 3: Label the Report Issue

After creating the report issue, call the `add_labels` safe output tool to apply labels based on the quality signals reported by the subagent. Collect the distinct `quality` values from all returned rows and add each as a label. The `add_labels` tool is pre-configured with `target-repo` pointing to the target repository.

For example, if the batch contains rows with `lgtm`, `spam`, and `needs-work` quality values, apply all three labels: `lgtm`, `spam`, `needs-work`.

If any subagent call failed (❓), also apply `outdated`.

## Important

- **You are the orchestrator**  -  you dispatch and compile. You do NOT run the checklist yourself.
- **PR fetching and filtering is pre-computed**  -  a `pre-agent` step writes `pr-filter-results.json`. Read it at the start.
- **Subagent does the analysis**  -  `.github/agents/contribution-checker.agent.md` handles all per-PR evaluation logic.
- **Read from `${{ env.TARGET_REPOSITORY }}`**  -  read-only access via GitHub MCP tools.
- **Write to `${{ github.repository }}`**  -  reports go here as issues.
- **Use safe output tools for target repository interactions**  -  use `add-comment` and `add-labels` safe output tools to post comments and labels to PRs in the target repository `${{ env.TARGET_REPOSITORY }}`. Never use `gh` CLI or direct API calls for writes.
- Close the previous report issue when creating a new one (`close-older-issues: true`).
- Be constructive in assessments  -  these reports help maintainers prioritize, not gatekeep.