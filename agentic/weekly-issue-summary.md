---
description: Creates weekly summary of issue activity including trends, charts, and insights every Monday

timeout-minutes: 20

on:
  schedule: weekly on monday
  workflow_dispatch:

permissions:
  issues: read

network:
  allowed:
    - defaults
    - python

tools:
  edit:
  bash:
    - "*"
  github:
    lockdown: true
    toolsets:
      - issues
    min-integrity: none # This workflow is allowed to examine and comment on any issues or PRs

safe-outputs:
  upload-asset:
  create-discussion:
    title-prefix: "[Weekly Summary] "
    category: "audits"
    close-older-discussions: true

steps:
  - name: Setup Python environment
    run: |
      mkdir -p /tmp/charts /tmp/data
      pip install --user --quiet numpy pandas matplotlib seaborn scipy
      python3 -c "import pandas, matplotlib, seaborn; print('Python environment ready')"
---

# Weekly Issue Summary

Create a comprehensive weekly summary of issue activity for repository ${{ github.repository }}.

## Step 1: Collect Issue Data

Use GitHub API tools to gather data for the past 30 days:

1. **Issue Activity Data**  -  Count of issues opened per day, closed per day, and running open count
2. **Issue Resolution Data**  -  Average time to close issues, distribution of issue lifespans, breakdown by label

Fetch enough issues to compute weekly and daily trends over the past 30 days. Use the GitHub toolset to query issues filtered by `created` and `closed` dates.

## Step 2: Generate Trend Charts

Write Python scripts to create exactly 2 high-quality trend charts and execute them via bash.

### Chart 1: Issue Activity Trends

Save data to `/tmp/data/issue_activity.csv` with columns: `date,opened,closed,open_total`

Generate a multi-line chart:

- Issues opened per week (bar or line)
- Issues closed per week (bar or line)
- Running total of open issues (secondary line)
- X-axis: last 12 weeks, Y-axis: count
- Save as `/tmp/charts/issue_activity_trends.png` at 300 DPI, 12×7 inches
- Use seaborn whitegrid style with a professional color palette

### Chart 2: Issue Resolution Time Trends

Save data to `/tmp/data/issue_resolution.csv` with columns: `date,avg_days,median_days`

Generate a line chart with moving average overlay:

- Average time to close (7-day moving average line)
- Median time to close
- Shaded variance band
- X-axis: last 30 days, Y-axis: days to resolution
- Save as `/tmp/charts/issue_resolution_trends.png` at 300 DPI, 12×7 inches

Run your Python scripts via bash and verify the charts exist before proceeding.

### Python Notes

- Use pandas for data manipulation and datetime handling
- Use `matplotlib.pyplot` and `seaborn` for visualization
- Apply `plt.tight_layout()` before saving
- Handle sparse data gracefully (use bar charts if fewer than 7 data points)
- Set `matplotlib.use('Agg')` to avoid display errors in headless environments

## Step 3: Upload Charts

Upload both chart images using the `upload-asset` safe output tool. Collect the returned URLs to embed in the discussion.

## Step 4: Create Weekly Discussion

Create a discussion with the title format: `Weekly Summary - [YYYY-MM-DD]`

### Formatting Guidelines

- Use `###` for main sections, `####` for subsections (discussion title is the h1)
- Wrap long lists in `<details><summary>` collapsible sections
- Keep critical information (overview, trends, statistics, recommendations) always visible
- Keep optional detail (full issue lists, verbose breakdowns) in collapsible sections

### Discussion Structure

```markdown
### 📊 Weekly Overview

[1–2 paragraphs: total issues opened and closed this week, how that compares to the previous week, key theme or pattern in the issues]

### 📈 Issue Activity Trends

#### Weekly Activity Patterns
![Issue Activity Trends]({chart_1_url})

[2–3 sentences: describe the trend  -  are issues accumulating, being resolved quickly, or holding steady?]

#### Resolution Time Analysis
![Issue Resolution Trends]({chart_2_url})

[2–3 sentences: how quickly are issues being resolved? improving or slowing down?]

### 🔑 Key Trends

[Bullet list of 3–5 notable patterns: common issue types, label distribution, new contributors filing issues, recurring topics, etc.]

### 📋 Summary Statistics

| Metric | This Week | Last Week | Trend |
|--------|-----------|-----------|-------|
| Issues Opened | X | X | ↑/↓/→ |
| Issues Closed | X | X | ↑/↓/→ |
| Currently Open | X | X | ↑/↓/→ |
| Avg Close Time | X days | X days | ↑/↓/→ |

<details>
<summary><b>Full Issue List (This Week)</b></summary>

[Numbered list of all issues opened this week with title, number, author, labels]

</details>

### 💡 Recommendations for Upcoming Week

[3–5 actionable suggestions: which issues to prioritize, patterns that suggest backlog growth, labels that need attention, etc.]
```

## Step 5: Notes

- If fewer than 7 days of data are available, generate charts with available data and note the limited range
- If no issues exist this week, still create a discussion noting the quiet week
- Always create the discussion even if charts fail to generate (omit chart sections and explain)
