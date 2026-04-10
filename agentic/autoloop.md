---
description: |
  An iterative optimization loop inspired by Karpathy's Autoresearch and Claude Code's /loop.
  Runs on a configurable schedule to autonomously improve a target artifact toward a measurable goal.
  Each iteration: reads the program definition, proposes a change, evaluates against a metric,
  and accepts or rejects the change. Tracks all iterations in a rolling GitHub issue.
  - User defines the optimization goal and evaluation criteria in a program.md file
  - Accepts changes only when they improve the metric (ratchet pattern)
  - Persists state between runs via repo memory
  - Creates draft PRs for accepted improvements
  - Maintains a living experiment log as a GitHub issue

on:
  schedule: every 6h
  workflow_dispatch:
  slash_command:
    name: autoloop

permissions: read-all

timeout-minutes: 45

network:
  allowed:
  - defaults
  - node
  - python
  - rust
  - java
  - dotnet

safe-outputs:
  add-comment:
    max: 5
    target: "*"
    hide-older-comments: false
  create-pull-request:
    draft: true
    title-prefix: "[Autoloop] "
    labels: [automation, autoloop]
    protected-files: fallback-to-issue
    max: 2
  push-to-pull-request-branch:
    target: "*"
    title-prefix: "[Autoloop] "
    max: 2
  create-issue:
    title-prefix: "[Autoloop] "
    labels: [automation, autoloop]
    max: 2
  update-issue:
    target: "*"
    title-prefix: "[Autoloop] "
    max: 1

tools:
  web-fetch:
  github:
    toolsets: [all]
  bash: true
  repo-memory: true

imports:
  - shared/reporting.md

steps:
  - name: Check which programs are due
    run: |
      python3 - << 'PYEOF'
      import os, json, re, glob, sys
      from datetime import datetime, timezone, timedelta

      programs_dir = ".autoloop/programs"
      state_file = ".autoloop/state.json"
      template_file = os.path.join(programs_dir, "example.md")

      # Bootstrap: create programs directory and template if missing
      if not os.path.isdir(programs_dir):
          os.makedirs(programs_dir, exist_ok=True)
          bt = chr(96)  # backtick — avoid literal backticks that break gh-aw compiler
          template = "\n".join([
              "<!-- AUTOLOOP:UNCONFIGURED -->",
              "<!-- Remove the line above once you have filled in your program. -->",
              "<!-- Autoloop will NOT run until you do. -->",
              "",
              "# Autoloop Program",
              "",
              "<!-- Rename this file to something meaningful (e.g. training.md, coverage.md).",
              "     The filename (minus .md) becomes the program name used in issues, PRs,",
              "     and slash commands. Want multiple loops? Add more .md files here. -->",
              "",
              "## Goal",
              "",
              "<!-- Describe what you want to optimize. Be specific about what 'better' means. -->",
              "",
              "REPLACE THIS with your optimization goal.",
              "",
              "## Target",
              "",
              "<!-- List files Autoloop may modify. Everything else is off-limits. -->",
              "",
              "Only modify these files:",
              f"- {bt}REPLACE_WITH_FILE{bt} -- (describe what this file does)",
              "",
              "Do NOT modify:",
              "- (list files that must not be touched)",
              "",
              "## Evaluation",
              "",
              "<!-- Provide a command and the metric to extract. -->",
              "",
              f"{bt}{bt}{bt}bash",
              "REPLACE_WITH_YOUR_EVALUATION_COMMAND",
              f"{bt}{bt}{bt}",
              "",
              f"The metric is {bt}REPLACE_WITH_METRIC_NAME{bt}. **Lower/Higher is better.** (pick one)",
              "",
          ])
          with open(template_file, "w") as f:
              f.write(template)
          # Leave the template unstaged — the agent will create a draft PR with it
          print(f"BOOTSTRAPPED: created {template_file} locally (agent will create a draft PR)")

      # Find all program files
      program_files = sorted(glob.glob(os.path.join(programs_dir, "*.md")))
      if not program_files:
          # Fallback to single-file locations
          for path in [".autoloop/program.md", "program.md"]:
              if os.path.isfile(path):
                  program_files = [path]
                  break

      if not program_files:
          print("NO_PROGRAMS_FOUND")
          os.makedirs("/tmp/gh-aw", exist_ok=True)
          with open("/tmp/gh-aw/autoloop.json", "w") as f:
              json.dump({"due": [], "skipped": [], "unconfigured": [], "no_programs": True}, f)
          sys.exit(0)

      os.makedirs("/tmp/gh-aw", exist_ok=True)
      now = datetime.now(timezone.utc)
      due = []
      skipped = []
      unconfigured = []

      # Schedule string to timedelta
      def parse_schedule(s):
          s = s.strip().lower()
          m = re.match(r"every\s+(\d+)\s*h", s)
          if m:
              return timedelta(hours=int(m.group(1)))
          m = re.match(r"every\s+(\d+)\s*m", s)
          if m:
              return timedelta(minutes=int(m.group(1)))
          if s == "daily":
              return timedelta(hours=24)
          if s == "weekly":
              return timedelta(days=7)
          return None  # No per-program schedule — always due

      for pf in program_files:
          name = os.path.splitext(os.path.basename(pf))[0]
          with open(pf) as f:
              content = f.read()

          # Check sentinel
          if "<!-- AUTOLOOP:UNCONFIGURED -->" in content:
              unconfigured.append(name)
              continue

          # Check for TODO/REPLACE placeholders
          if re.search(r'\bTODO\b|\bREPLACE', content):
              unconfigured.append(name)
              continue

          # Parse optional YAML frontmatter for schedule
          schedule_delta = None
          fm_match = re.match(r"^---\s*\n(.*?)\n---\s*\n", content, re.DOTALL)
          if fm_match:
              for line in fm_match.group(1).split("\n"):
                  if line.strip().startswith("schedule:"):
                      schedule_str = line.split(":", 1)[1].strip()
                      schedule_delta = parse_schedule(schedule_str)

          # Read lightweight state file (committed to repo, not repo-memory)
          # state.json tracks: last_run timestamps, pause flags, recent statuses
          state = {}
          if os.path.isfile(state_file):
              try:
                  with open(state_file) as f:
                      all_state = json.load(f)
                  state = all_state.get(name, {})
              except (json.JSONDecodeError, ValueError):
                  pass

          last_run = None
          lr = state.get("last_run")
          if lr:
              try:
                  last_run = datetime.fromisoformat(lr.replace("Z", "+00:00"))
              except ValueError:
                  pass

          # Check if paused (e.g., plateau or recurring errors)
          if state.get("paused"):
              skipped.append({"name": name, "reason": f"paused: {state.get('pause_reason', 'unknown')}"})
              continue

          # Auto-pause on plateau: 5+ consecutive rejections
          recent = state.get("recent_statuses", [])[-5:]
          if len(recent) >= 5 and all(s == "rejected" for s in recent):
              skipped.append({"name": name, "reason": "plateau: 5 consecutive rejections"})
              continue

          # Check if due based on per-program schedule
          if schedule_delta and last_run:
              if now - last_run < schedule_delta:
                  skipped.append({"name": name, "reason": "not due yet",
                                  "next_due": (last_run + schedule_delta).isoformat()})
                  continue

          due.append(name)

      result = {"due": due, "skipped": skipped, "unconfigured": unconfigured, "no_programs": False}

      os.makedirs("/tmp/gh-aw", exist_ok=True)
      with open("/tmp/gh-aw/autoloop.json", "w") as f:
          json.dump(result, f, indent=2)

      print("=== Autoloop Program Check ===")
      print(f"Programs due:          {due or '(none)'}")
      print(f"Programs skipped:      {[s['name'] for s in skipped] or '(none)'}")
      print(f"Programs unconfigured: {unconfigured or '(none)'}")

      if not due and not unconfigured:
          print("\nNo programs due this run. Exiting early.")
          sys.exit(1)  # Non-zero exit skips the agent step
      PYEOF

---

# Autoloop

An iterative optimization agent that proposes changes, evaluates them against a metric, and keeps only improvements — running autonomously on a schedule.

## Command Mode

Take heed of **instructions**: "${{ steps.sanitized.outputs.text }}"

If these are non-empty (not ""), then you have been triggered via `/autoloop <instructions>`. The instructions may be:
- **A one-off directive targeting a specific program**: e.g., `/autoloop training: try a different approach to the loss function`. The text before the colon is the program name (matching a file in `.autoloop/programs/`). Execute it as a single iteration for that program, then report results.
- **A general directive**: e.g., `/autoloop try cosine annealing`. If no program name prefix is given and only one program exists, use that one. If multiple exist, ask which program to target.
- **A configuration change**: e.g., `/autoloop training: set metric to accuracy instead of loss`. Update the relevant program file and confirm.

Then exit — do not run the normal loop after completing the instructions.

## Multiple Programs

Autoloop supports **multiple independent optimization loops** in the same repository. Each loop is defined by a separate markdown file in `.autoloop/programs/`. For example:

```
.autoloop/programs/
├── training.md      ← optimize model training
├── coverage.md      ← maximize test coverage
└── build-perf.md    ← minimize build time
```

Each program runs independently with its own:
- Goal, target files, and evaluation command
- Metric tracking and best-metric history
- Experiment log issue: `[Autoloop: {program-name}] Experiment Log {YYYY-MM}`
- Branch namespace: `autoloop/{program-name}/iteration-<N>-<desc>`
- PR title prefix: `[Autoloop: {program-name}]`
- Repo memory namespace: keyed by program name

On each scheduled run, a lightweight pre-step checks which programs are due (based on per-program schedules and `last_run` timestamps). **If no programs are due, the workflow exits before the agent starts — zero agent cost.** Only due programs get iterated.

### Per-Program Schedule and Timeout

Programs can optionally specify their own schedule and timeout in a YAML frontmatter block at the top of the file (after the sentinel, if present):

```markdown
---
schedule: every 1h
timeout-minutes: 30
---

# Autoloop Program
...
```

- **`schedule`**: Controls how often this program runs. On each workflow trigger, check if the program is due based on its schedule and the `last_run` timestamp in memory. If the program's schedule hasn't elapsed since its last run, skip it. If omitted, the program runs on every workflow trigger.
- **`timeout-minutes`**: Maximum time for this program's iteration. If omitted, the program shares the workflow's overall timeout.

This lets you run a fast coverage check every hour while running a slow training loop once a day — all from the same workflow.

## Program Definition

Each program file in `.autoloop/programs/` defines three things:

1. **Goal**: What the agent is trying to optimize (natural language description)
2. **Target**: Which files the agent is allowed to modify
3. **Evaluation**: How to measure whether a change is an improvement

The **program name** is the filename without the `.md` extension (e.g., `training.md` → program name is `training`).

### Setup Guard

A template program file is installed at `.autoloop/programs/example.md`. **Programs will not run until the user has edited them.** Each template contains a sentinel line:

```
<!-- AUTOLOOP:UNCONFIGURED -->
```

At the start of every run, check each program file for this sentinel. For any program where it is present:

1. **Skip that program — do not run any iterations for it.**
2. If no setup issue exists for that program, create one titled `[Autoloop: {program-name}] Action required: configure your program` with:
   - A clear explanation that this program is installed but paused until configured.
   - A direct link to edit the file on GitHub (use the repository's default branch in the URL).
   - A brief guide: "Open the file, replace the placeholder sections with your project's goal, target files, and evaluation command, then remove the `<!-- AUTOLOOP:UNCONFIGURED -->` line."
   - Two or three example programs for inspiration (ML training, test coverage, build performance).

If **all** programs are unconfigured, exit after creating the setup issues. Otherwise, proceed with the configured programs.

**Important**: When creating or modifying template/program files during setup, always do so via a draft PR — never commit directly to the default branch. Only iteration state files (`state.json`) should be committed directly.

### Reading Programs

The pre-step has already determined which programs are due, unconfigured, or skipped. Read `/tmp/gh-aw/autoloop.json` at the start of your run to get:

- **`due`**: List of program names to run iterations for this run.
- **`unconfigured`**: Programs that still have the sentinel or placeholder content. For each unconfigured program:
  1. Check whether the program file actually exists on the default branch (use `git show HEAD:.autoloop/programs/{name}.md`). If it does NOT exist on the default branch, **you must create a draft PR** (branch: `autoloop/setup-template`) that adds the template file. The pre-step may have created the file locally in the working directory, so it will be available to commit — just create a branch, commit it, and open the PR.
  2. If no setup issue exists for this program, create one (see Setup Guard above).
  3. If the file already exists on the default branch and a setup issue already exists, then no action is needed for this program.
- **`skipped`**: Programs not due yet based on their per-program schedule — ignore these entirely.
- **`no_programs`**: If `true`, no program files exist at all. The pre-step should have bootstrapped a template locally. Follow the same steps as `unconfigured` above — create a draft PR with the template and a setup issue.

For each program in `due`:
1. Read the program file from `.autoloop/programs/{name}.md`.
2. Parse the three sections: Goal, Target, Evaluation.
3. Read the current state of all target files.
4. Read repo memory for that program's metric history (keyed by program name).

## Iteration Loop

Each run executes **one iteration per configured program**. For each program:

### Step 1: Read State

1. Read the program file to understand the goal, targets, and evaluation method.
2. Read `.autoloop/state.json` for this program's `best_metric` and `iteration_count`.
3. Read repo memory (keyed by program name) for detailed history:
   - `history`: Summary of recent iterations (last 20).
   - `rejected_approaches`: Approaches that were tried and failed (to avoid repeating).
   - `consecutive_errors`: Count of consecutive evaluation failures.

### Step 2: Analyze and Propose

1. Read the target files and understand the current state.
2. Review the history of previous iterations — what worked, what didn't.
3. **Think carefully** about what change is most likely to improve the metric. Consider:
   - What has been tried before and rejected (don't repeat failures).
   - What the evaluation criteria reward.
   - Small, targeted changes are more likely to succeed than large rewrites.
   - If many small optimizations have been exhausted, consider a larger architectural change.
4. Describe the proposed change in your reasoning before implementing it.

### Step 3: Implement

1. Create a fresh branch: `autoloop/{program-name}/iteration-<N>-<short-desc>` from the default branch.
2. Make the proposed changes to the target files only.
3. **Respect the program constraints**: do not modify files outside the target list.

### Step 4: Evaluate

1. Run the evaluation command specified in `program.md`.
2. Parse the metric from the output.
3. Compare against `best_metric` from memory.

### Step 5: Accept or Reject

**If the metric improved** (or this is the first run establishing a baseline):
1. Create a draft PR with:
   - Title: `[Autoloop: {program-name}] Iteration <N>: <short description of change>`
   - Body includes: what was changed, why, the old metric, the new metric, and the improvement delta.
   - AI disclosure: `🤖 *This change was proposed and validated by Autoloop.*`
2. Add an entry to the experiment log issue.
3. Update repo memory: add to `history`, reset `consecutive_errors` to 0.
4. Update `state.json`: set `best_metric`, increment `iteration_count`, set `last_run`, append `"accepted"` to `recent_statuses`. **Commit and push.**

**If the metric did not improve** (or evaluation failed):
1. Do NOT create a PR.
2. Update repo memory: add to `rejected_approaches` with what was tried, the resulting metric, and why it likely didn't work.
3. Add a "rejected" entry to the experiment log issue.
4. Update `state.json`: increment `iteration_count`, set `last_run`, append `"rejected"` to `recent_statuses`. **Commit and push.**

**If evaluation could not run** (build failure, missing dependencies, etc.):
1. Do NOT create a PR.
2. Update repo memory: increment `consecutive_errors`.
3. Add an "error" entry to the experiment log issue.
4. If `consecutive_errors` reaches 3+, set `paused: true` and `pause_reason` in `state.json`, and create an issue describing the problem.
5. Update `state.json`: increment `iteration_count`, set `last_run`, append `"error"` to `recent_statuses`. **Commit and push.**

## Experiment Log Issue

Maintain a single open issue **per program** titled `[Autoloop: {program-name}] Experiment Log {YYYY}-{MM}` as a rolling record of that program's iterations.

### Issue Body Format

```markdown
🤖 *Autoloop — an iterative optimization agent for this repository.*

## Program

**Goal**: {one-line summary from program.md}
**Target files**: {list of target files}
**Metric**: {metric name} ({higher/lower} is better)
**Current best**: {best_metric} (established in iteration {N})

## Iteration History

### Iteration {N} — {YYYY-MM-DD HH:MM UTC} — [Run]({run_url})
- **Status**: ✅ Accepted / ❌ Rejected / ⚠️ Error
- **Change**: {one-line description}
- **Metric**: {value} (previous best: {previous_best}, delta: {delta})
- **PR**: #{number} (if accepted)

### Iteration {N-1} — {YYYY-MM-DD HH:MM UTC} — [Run]({run_url})
- **Status**: ❌ Rejected
- **Change**: {one-line description}
- **Metric**: {value} (previous best: {previous_best}, delta: {delta})
- **Reason**: {why it was rejected}
```

### Format Rules

- Iterations in **reverse chronological order** (newest first).
- Each iteration heading links to its GitHub Actions run.
- Use `${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}` for the current run URL.
- Close the previous month's issue and create a new one at month boundaries.
- Maximum 50 iterations per issue; create a continuation issue if exceeded.

## State and Memory

Autoloop uses **two persistence layers**:

### 1. State file (`.autoloop/state.json`) — lightweight, committed to repo

This file is read by the **pre-step** (before the agent starts) to decide which programs are due. The agent **must update this file and commit it** at the end of every iteration. This is the only way the pre-step can check schedules, plateaus, and pause flags on future runs.

```json
{
  "training": {
    "last_run": "2025-01-15T12:00:00Z",
    "best_metric": 0.0234,
    "iteration_count": 17,
    "paused": false,
    "pause_reason": null,
    "recent_statuses": ["accepted", "rejected", "rejected", "accepted", "accepted"]
  },
  "coverage": {
    "last_run": "2025-01-15T06:00:00Z",
    "best_metric": 78.4,
    "iteration_count": 5,
    "paused": false,
    "pause_reason": null,
    "recent_statuses": ["accepted", "accepted", "rejected", "accepted", "accepted"]
  }
}
```

**After every iteration** (accepted, rejected, or error), update this program's entry in `state.json`:
- Set `last_run` to the current UTC timestamp.
- Update `best_metric` if the iteration was accepted.
- Increment `iteration_count`.
- Append the status (`"accepted"`, `"rejected"`, or `"error"`) to `recent_statuses` (keep last 10).
- Set `paused`/`pause_reason` if needed.
- **Commit and push** the updated `state.json` to the default branch.

### 2. Repo memory — full history for the agent

Use repo-memory (keyed by program name, e.g., `autoloop/training`) for detailed state the agent needs but the pre-step doesn't:

```json
{
  "program_name": "training",
  "history": [
    {
      "iteration": 17,
      "status": "accepted",
      "description": "Reduced learning rate warmup from 5 to 3 epochs",
      "metric": 0.0234,
      "previous_best": 0.0241,
      "pr": 42
    }
  ],
  "rejected_approaches": [
    {
      "iteration": 16,
      "description": "Switched from Adam to SGD with momentum",
      "metric": 0.0298,
      "reason": "SGD converges slower within the 5-minute budget"
    }
  ],
  "consecutive_errors": 0
}
```

## Guidelines

- **One change per iteration.** Keep changes small and targeted. A single hyperparameter tweak, a minor architectural modification, or a focused code optimization. This makes it clear what caused metric changes.
- **No breaking changes.** Target files must remain functional even if the iteration is rejected.
- **Respect the evaluation budget.** If the evaluation command has a time constraint (e.g., 5-minute training), respect it. Do not modify evaluation scripts or timeout settings.
- **Learn from history.** The rejected_approaches list exists to prevent repeating failures. Read it carefully before proposing changes.
- **Diminishing returns.** If the last 5 consecutive iterations were rejected, post a comment on the experiment log suggesting the user review the program definition — the optimization may have plateaued.
- **Transparency.** Every PR and comment must include AI disclosure with 🤖.
- **Safety.** Never modify files outside the target list. Never modify the evaluation script. Never modify program.md (except via `/autoloop` command mode).
- **Read AGENTS.md first**: before starting work, read the repository's `AGENTS.md` file (if present) to understand project-specific conventions.
- **Build and test**: run any build/test commands before creating PRs. If your changes break the build, reject the iteration.
