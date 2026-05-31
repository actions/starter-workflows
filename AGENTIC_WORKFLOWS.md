# Agentic Workflows Guide

Agentic workflows are intelligent, AI-powered GitHub Actions workflows that use Claude AI to automate complex development tasks. Unlike traditional workflows which execute fixed scripts, agentic workflows use natural language prompts to enable Claude to analyze your codebase, logs, and repository context, then take intelligent action.

---

## Quick Start

Agentic workflows are available in the **Agentic** category when you create a new GitHub Actions workflow. They follow the same setup process as other starter workflows.

---

## Format: Markdown Instead of YAML

Unlike traditional starter workflows (which use `.yml` format), agentic workflows use **Markdown (`.md`)** format. This allows both structured YAML configuration and natural language prompts in a single file.

### Structure

```markdown
---
# YAML Frontmatter: workflow configuration
name: My Agentic Workflow
description: What this workflow does
on: [trigger events]
permissions: [required permissions]

# Special agentic features
safe-outputs:
  action-name:
    property: value
tools:
  cache-memory: true
  web-fetch:

timeout-minutes: 10
---

# Markdown Body: Claude's Prompt

## Context
- Repository details
- Configuration variables

## Instructions
1. Task description
2. Analysis protocol
3. Action guidelines
4. Safety constraints
```

### Key Components

| Component | Purpose | Example |
|-----------|---------|---------|
| **Trigger** | When the workflow runs | `workflow_run`, `schedule`, `pull_request` |
| **Permissions** | GitHub API access | `contents: read`, `actions: read`, `issues: write` |
| **safe-outputs** | Actions Claude can safely execute | Create issues, add comments, create PRs |
| **tools** | External capabilities | `cache-memory`, `web-fetch` |
| **Markdown body** | Claude's instructions/prompt | Analysis protocol, decision rules, output format |

---

## Available Workflows

### 1. CI Doctor
- **Purpose:** Investigates failed CI workflows to identify root causes
- **Trigger:** When a monitored workflow fails
- **Actions:** Creates diagnostic issues with analysis
- **Use case:** Reduce debugging time by automatically analyzing CI failures

### 2. Code Simplifier
- **Purpose:** Analyzes code and suggests simplifications
- **Trigger:** Manual dispatch or pull request review
- **Actions:** Creates PR comments with suggestions
- **Use case:** Automated code quality feedback during review

### 3. Daily Doc Updater
- **Purpose:** Updates documentation based on code changes
- **Trigger:** Daily schedule
- **Actions:** Updates markdown files, creates commits
- **Use case:** Keep docs in sync with implementation

### 4. Daily Test Improver
- **Purpose:** Analyzes test coverage and suggests improvements
- **Trigger:** Daily schedule
- **Actions:** Creates issues with improvement suggestions
- **Use case:** Incremental test coverage expansion

### 5. Daily Repo Status
- **Purpose:** Generates daily repository health report
- **Trigger:** Daily schedule (e.g., 6am)
- **Actions:** Creates discussion post or issue summary
- **Use case:** Daily digest of repo metrics

### 6. Daily Team Status
- **Purpose:** Summarizes daily activity for team sync
- **Trigger:** Daily schedule before standup
- **Actions:** Creates summary comments or posts
- **Use case:** Automated status update generation

### 7. Duplicate Code Detector
- **Purpose:** Finds and reports duplicate code patterns
- **Trigger:** Pull request
- **Actions:** Creates PR comments highlighting duplicates
- **Use case:** Identify refactoring opportunities

### 8. Issue Triage
- **Purpose:** Automatically categorizes and prioritizes issues
- **Trigger:** New issue created
- **Actions:** Adds labels, assigns priority, creates summary
- **Use case:** Reduce manual issue triage work

### 9. PR Fix
- **Purpose:** Suggests and implements fixes for PR feedback
- **Trigger:** Manual dispatch with context
- **Actions:** Creates commits with suggested fixes
- **Use case:** Accelerate PR resolution cycles

### 10. Repo Assist
- **Purpose:** General-purpose repository assistant
- **Trigger:** Manual dispatch with custom task
- **Actions:** Analyzes code, suggests improvements
- **Use case:** Ad-hoc automation for dev tasks

### 11. Repository Quality Improver
- **Purpose:** Comprehensive repository quality analysis and improvement
- **Trigger:** Schedule or manual dispatch
- **Actions:** Creates issues for improvements, refactoring suggestions
- **Use case:** Systematic code quality improvements

---

## Safety & Constraints

### Safe Outputs
Agentic workflows use `safe-outputs` to restrict which GitHub API actions Claude can execute:

```yaml
safe-outputs:
  create-issue:
    title-prefix: "🤖 Automated:"
    labels: [automation]
  add-comment:
    # No restrictions — can comment anywhere
  create-pull-request:
    # Only when explicitly enabled
```

This ensures Claude cannot:
- Merge pull requests
- Delete branches or repositories
- Force-push to protected branches
- Approve reviews
- Other destructive actions

### Permissions
Each workflow declares the minimum GitHub API permissions needed:

```yaml
permissions:
  contents: read          # Read code
  issues: write           # Create/update issues
  pull-requests: read     # Read PR details
  actions: read           # Read workflow logs
```

Workflows only have access to declared permissions.

---

## Customization

### Configure Trigger Events
Edit the `on:` section in the frontmatter to match your workflow:

```yaml
on:
  schedule:
    - cron: '0 9 * * *'  # Daily at 9am UTC
  workflow_dispatch:     # Manual trigger button
  push:
    branches: [main]
```

### Adjust Timeout
Extend or reduce the workflow timeout based on task complexity:

```yaml
timeout-minutes: 30  # Default is usually 10-15
```

### Modify Constraints
Update `safe-outputs` and `tools` to match your needs:

```yaml
tools:
  cache-memory: true      # Enable Claude's memory cache
  web-fetch:             # Enable external web access
```

### Edit the Prompt
The markdown body is Claude's instruction set. Edit it to:
- Change analysis depth or focus
- Add organization-specific rules
- Adjust output format
- Add guardrails specific to your codebase

---

## Stability & Roadmap

### Current Status: Preview
Agentic workflows are a **new feature** (May 2026) and are in **preview/experimental status**:
- May not be available on GitHub Enterprise Server yet
- Format may change as the feature matures
- Feedback is welcome; please report issues
- Performance and behavior may vary

### Future Roadmap
- [ ] Transition from `.md` to standardized YAML format (if format stabilizes)
- [ ] Expand to GitHub Enterprise Server (`@latest-ghe`)
- [ ] Add more ready-made workflows based on customer feedback
- [ ] Introduce certified/supported tier for production use
- [ ] Enhanced debugging and logging capabilities

### When Format Changes
If agentic workflows transition to `.yml` format in the future:
1. **Migration path:** Existing `.md` files will continue to work
2. **Deprecation notice:** Will be announced 6+ months in advance
3. **Tooling:** Automated converter or migration guide will be provided
4. **Backward compatibility:** `.md` format will be supported during transition period

---

## Troubleshooting

### Workflow Not Triggering
- Check the `on:` trigger conditions match your repository state
- Verify the branch in `branches:` matches your default branch
- Ensure `workflow_dispatch` is enabled if manually triggering

### Claude Doesn't Take Expected Action
- Check `safe-outputs` includes the action you want
- Verify `permissions:` include required GitHub API scopes
- Review the markdown prompt — it controls Claude's behavior
- Check workflow logs for error details

### Workflow Timeout
- Increase `timeout-minutes` in the frontmatter
- Simplify the analysis scope in the markdown prompt
- Enable `cache-memory: true` to reuse prior analysis

### Unsupported on Your Server
- Check if your GitHub instance is version 3.10+ (for preview features)
- For GitHub Enterprise, agentic workflows may require a newer version
- File an issue if you need GHES support

---

## Examples & Recipes

### Daily Security Scan + Report
Combine multiple agentic workflows:
```yaml
- CI Doctor (when tests fail)
- Duplicate Code Detector (on every PR)
- Repository Quality Improver (nightly)
→ Result: Comprehensive daily security and quality digest
```

### PR Quality Gating
```yaml
- Code Simplifier (on PR)
- Duplicate Code Detector (on PR)
- PR Fix (if feedback given)
→ Result: Guided improvement loop for pull requests
```

### Documentation Auto-Sync
```yaml
- Daily Doc Updater (on schedule)
- Repo Assist (manual for edge cases)
→ Result: Docs always reflect current code
```

---

## Contributing & Feedback

Agentic workflows are new. We'd love your feedback:
- **Issues:** Report bugs or unexpected behavior
- **Discussions:** Suggest new workflows or use cases
- **Pull Requests:** Currently not accepting contributions, but this may change as the feature stabilizes

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## Learn More

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Actions Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Claude AI Documentation](https://claude.ai/docs)
- [Safe Outputs & Constraints](https://docs.github.com/en/actions/using-workflows/agentic-workflows)
