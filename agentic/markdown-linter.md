---
description: Runs Markdown quality checks using Super Linter and creates issues for violations
on:
  workflow_dispatch:
  schedule:
    - cron: "0 14 * * 1-5" # 2 PM UTC, weekdays only

permissions:
  contents: read
  actions: read
  issues: read
  pull-requests: read

safe-outputs:
  create-issue:
    expires: 2d
    title-prefix: "[linter] "
    labels: [automation, code-quality]
  noop:

name: Markdown Linter
timeout-minutes: 15

imports:
  - shared/reporting.md

jobs:
  super_linter:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: read
      statuses: write
    steps:
      - name: Checkout repository
        uses: actions/checkout@v6.0.2
        with:
          fetch-depth: 0
          persist-credentials: false

      - name: Super-linter
        uses: super-linter/super-linter@v8.5.0
        id: super-linter
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          CREATE_LOG_FILE: "true"
          LOG_FILE: super-linter.log
          DEFAULT_BRANCH: main
          ENABLE_GITHUB_ACTIONS_STEP_SUMMARY: "true"
          VALIDATE_MARKDOWN: "true"
          VALIDATE_ALL_CODEBASE: "false"

      - name: Check for linting issues
        id: check-results
        run: |
          if [ -f "super-linter.log" ] && [ -s "super-linter.log" ]; then
            if grep -qE "ERROR|WARN|FAIL" super-linter.log; then
              echo "needs-linting=true" >> "$GITHUB_OUTPUT"
            else
              echo "needs-linting=false" >> "$GITHUB_OUTPUT"
            fi
          else
            echo "needs-linting=false" >> "$GITHUB_OUTPUT"
          fi

      - name: Upload super-linter log
        if: always()
        uses: actions/upload-artifact@v7
        with:
          name: super-linter-log
          path: super-linter.log
          retention-days: 7
steps:
  - name: Download super-linter log
    uses: actions/download-artifact@v8
    with:
      name: super-linter-log
      path: /tmp/gh-aw/
tools:
  cache-memory: true
  edit:
  bash:
    - "*"
---

# Markdown Quality Report

You are an expert documentation quality analyst. Your task is to analyze the Super Linter Markdown output and create a comprehensive issue report for the repository maintainers.

## Context

- **Repository**: ${{ github.repository }}
- **Triggered by**: @${{ github.actor }}
- **Run ID**: ${{ github.run_id }}

## Your Task

1. **Read the linter output** from `/tmp/gh-aw/super-linter.log` using the bash tool
2. **Analyze the findings**:
   - Categorize errors by severity (critical, high, medium, low)
   - Identify patterns in the errors
   - Determine which errors are most important to fix first
   - Note: This workflow only validates Markdown files
3. **Create a detailed issue** with the following structure:

### Issue Title
Use format: "Markdown Quality Report - [Date] - [X] issues found"

### Issue Body Structure

```markdown
## 🔍 Markdown Linter Summary

**Date**: [Current date]
**Total Issues Found**: [Number]
**Run ID**: ${{ github.run_id }}

## 📊 Breakdown by Severity

- **Critical**: [Count and brief description]
- **High**: [Count and brief description]
- **Medium**: [Count and brief description]
- **Low**: [Count and brief description]

## 📁 Issues by Category

### [Category/Rule Name]
- **File**: `path/to/file`
  - Line [X]: [Error description]
  - Suggested fix: [How to resolve]

[Repeat for other categories]

## 🎯 Priority Recommendations

1. [Most critical issue to address first]
2. [Second priority]
3. [Third priority]

## 📋 Full Linter Output

<details>
<summary>Click to expand complete linter log</summary>

```
[Include the full linter output here]
```

</details>

## 🔗 References

- [Link to workflow run](${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }})
- [Super Linter Documentation](https://github.com/super-linter/super-linter)
```

## Important Guidelines

- **Be concise but thorough**: Focus on actionable insights
- **Prioritize issues**: Not all linting errors are equal
- **Provide context**: Explain why each type of error matters for documentation quality
- **Suggest fixes**: Give practical recommendations
- **Use proper formatting**: Make the issue easy to read and navigate
- **If no errors found**: Call `noop` celebrating clean markdown

**Important**: Always call exactly one safe-output tool before finishing (`create_issue` or `noop`).

```json
{"noop": {"message": "No action needed: [brief explanation of what was analyzed and why]"}}
```
