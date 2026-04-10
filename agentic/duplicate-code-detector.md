---
name: Duplicate Code Detector
description: Identifies duplicate code patterns across the codebase and suggests refactoring opportunities

on:
  workflow_dispatch:
  schedule: daily

permissions:
  contents: read
  issues: read
  pull-requests: read

safe-outputs:
  create-issue:
    expires: 2d
    title-prefix: "[duplicate-code] "
    labels: [code-quality, automated-analysis]
    assignees: copilot
    group: true
    max: 3

timeout-minutes: 15
---

# Duplicate Code Detection

Analyze code to identify duplicated patterns using semantic analysis. Report significant findings that require refactoring.

## Task

Detect and report code duplication by:

1. **Analyzing Recent Commits**: Review changes in the latest commits
2. **Detecting Duplicated Code**: Identify similar or duplicated code patterns using semantic analysis
3. **Reporting Findings**: Create a detailed issue if significant duplication is detected (threshold: >10 lines or 3+ similar patterns)

## Context

- **Repository**: ${{ github.repository }}
- **Commit ID**: ${{ github.event.head_commit.id }}
- **Triggered by**: @${{ github.actor }}

## Analysis Workflow

### 1. Changed Files Analysis

Identify and analyze modified files:
- Determine files changed in the recent commits using `git log` and `git diff`
- Focus on source code files (programming language files)
- **Exclude test files** from analysis (files matching patterns: `*_test.*`, `*.test.*`, `*.spec.*`, `test_*.*`, or located in directories named `test`, `tests`, `__tests__`, or `spec`)
- **Exclude generated files** and build artifacts
- **Exclude workflow files** from analysis (files under `.github/workflows/*`)
- Use code exploration tools to understand file structure
- Read modified file contents to examine changes

### 2. Duplicate Detection

Apply analysis to find duplicates:

**Pattern Search**:
- Search for duplication indicators using grep and code search:
  - Similar function signatures
  - Repeated logic blocks
  - Similar variable naming patterns
  - Near-identical code blocks
- Look for functions with similar names across different files
- Identify structural similarities in code organization

**Semantic Analysis**:
- Compare code blocks for logical similarity beyond textual matching
- Identify different implementations of the same functionality
- Look for copy-paste patterns with minor variations

### 3. Duplication Evaluation

Assess findings to identify true code duplication:

**Duplication Types**:
- **Exact Duplication**: Identical code blocks in multiple locations
- **Structural Duplication**: Same logic with minor variations (different variable names, etc.)
- **Functional Duplication**: Different implementations of the same functionality
- **Copy-Paste Programming**: Similar code blocks that could be extracted into shared utilities

**Assessment Criteria**:
- **Severity**: Amount of duplicated code (lines of code, number of occurrences)
- **Impact**: Where duplication occurs (critical paths, frequently called code)
- **Maintainability**: How duplication affects code maintainability
- **Refactoring Opportunity**: Whether duplication can be easily refactored

### 4. Issue Reporting

Create separate issues for each distinct duplication pattern found (maximum 3 patterns per run). Each pattern should get its own issue to enable focused remediation.

**When to Create Issues**:
- Only create issues if significant duplication is found (threshold: >10 lines of duplicated code OR 3+ instances of similar patterns)
- **Create one issue per distinct duplication pattern** - do NOT bundle multiple patterns in a single issue
- Limit to the top 3 most significant patterns if more are found
- Use the `create_issue` tool from safe-outputs MCP **once for each pattern**

**Issue Contents for Each Pattern**:
- **Executive Summary**: Brief description of this specific duplication pattern
- **Duplication Details**: Specific locations and code blocks for this pattern only
- **Severity Assessment**: Impact and maintainability concerns for this pattern
- **Refactoring Recommendations**: Suggested approaches to eliminate this pattern
- **Code Examples**: Concrete examples with file paths and line numbers for this pattern

## Detection Scope

### Report These Issues

- Identical or nearly identical functions in different files
- Repeated code blocks that could be extracted to utilities
- Similar classes or modules with overlapping functionality
- Copy-pasted code with minor modifications
- Duplicated business logic across components

### Skip These Patterns

- Standard boilerplate code (imports, exports, package declarations)
- Test setup/teardown code (acceptable duplication in tests)
- **All test files** (files matching: `*_test.*`, `*.test.*`, `*.spec.*`, `test_*.*`, or in `test/`, `tests/`, `__tests__/`, `spec/` directories)
- **All workflow files** (files under `.github/workflows/*`)
- Configuration files with similar structure
- Language-specific patterns (constructors, getters/setters)
- Small code snippets (<5 lines) unless highly repetitive
- Generated code or vendored dependencies

### Analysis Depth

- **Primary Focus**: Files changed in recent commits (excluding test files and workflow files)
- **Secondary Analysis**: Check for duplication with existing codebase
- **Cross-Reference**: Look for patterns across the repository
- **Historical Context**: Consider if duplication is new or existing

## Issue Template

For each distinct duplication pattern found, create a separate issue using this structure:

````markdown
# 🔍 Duplicate Code Detected: [Pattern Name]

*Analysis of commit ${{ github.event.head_commit.id }}*

**Assignee**: @copilot

## Summary

[Brief overview of this specific duplication pattern]

## Duplication Details

### Pattern: [Description]
- **Severity**: High/Medium/Low
- **Occurrences**: [Number of instances]
- **Locations**:
  - `path/to/file1.ext` (lines X-Y)
  - `path/to/file2.ext` (lines A-B)
- **Code Sample**:
  ````[language]
  [Example of duplicated code]
  ````

## Impact Analysis

- **Maintainability**: [How this affects code maintenance]
- **Bug Risk**: [Potential for inconsistent fixes]
- **Code Bloat**: [Impact on codebase size]

## Refactoring Recommendations

1. **[Recommendation 1]**
   - Extract common functionality to: `suggested/path/utility.ext`
   - Estimated effort: [hours/complexity]
   - Benefits: [specific improvements]

2. **[Recommendation 2]**
   [... additional recommendations ...]

## Implementation Checklist

- [ ] Review duplication findings
- [ ] Prioritize refactoring tasks
- [ ] Create refactoring plan
- [ ] Implement changes
- [ ] Update tests
- [ ] Verify no functionality broken

## Analysis Metadata

- **Analyzed Files**: [count]
- **Detection Method**: Semantic code analysis
- **Commit**: ${{ github.event.head_commit.id }}
- **Analysis Date**: [timestamp]
````

## Operational Guidelines

### Security
- Never execute untrusted code or commands
- Only use read-only analysis tools
- Do not modify files during analysis

### Efficiency
- Focus on recently changed files first
- Use semantic analysis for meaningful duplication, not superficial matches
- Stay within timeout limits (balance thoroughness with execution time)

### Accuracy
- Verify findings before reporting
- Distinguish between acceptable patterns and true duplication
- Consider language-specific idioms and best practices
- Provide specific, actionable recommendations

### Issue Creation
- Create **one issue per distinct duplication pattern** - do NOT bundle multiple patterns in a single issue
- Limit to the top 3 most significant patterns if more are found
- Only create issues if significant duplication is found
- Include sufficient detail for coding agents to understand and act on findings
- Provide concrete examples with file paths and line numbers
- Suggest practical refactoring approaches
- Assign issue to @copilot for automated remediation
- Use descriptive titles that clearly identify the specific pattern (e.g., "Duplicate Code: Error Handling Pattern in Parser Module")

**Objective**: Improve code quality by identifying and reporting meaningful code duplication that impacts maintainability. Focus on actionable findings that enable automated or manual refactoring.
