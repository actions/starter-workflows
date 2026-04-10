---
description: |
  This workflow performs strategic project planning by maintaining and updating the project roadmap.
  Analyzes repository state including open issues, PRs, and completed work to formulate
  a comprehensive project plan. Creates or updates a planning discussion with prioritized
  tasks, dependencies, and suggested new issues (via gh commands but doesn't create them).
  Incorporates maintainer feedback from comments on the plan.

on:
  schedule: daily
  workflow_dispatch:

permissions: read-all

network: defaults

safe-outputs:
  mentions: false
  allowed-github-references: []
  create-discussion: # needed to create the project plan discussion
    title-prefix: "${{ github.workflow }}"
    category: "announcements"
    close-older-discussions: true

tools:
  github:
    toolsets: [all]
    # If in a public repo, setting `lockdown: false` allows
    # reading issues, pull requests and comments from 3rd-parties
    # If in a private repo this has no particular effect.
    lockdown: false
    min-integrity: none # This workflow is allowed to examine and comment on any issues
  web-fetch:

timeout-minutes: 15
---

# Agentic Planner

## Job Description

Your job is to act as a planner for the GitHub repository ${{ github.repository }}.

1. First study the state of the repository including, open issues, pull requests, completed issues.

   1a. As part of this, look for the open discussion with title starting with "${{ github.workflow }}", which is the existing project plan. Read the plan, and any comments on the plan. If no such discussion exists, ignore this step.

   1b. You can read code, search the web and use other tools to help you understand the project and its requirements.

2. Formulate a plan for the remaining work to achieve the objectives of the project.

   2a. The project plan should be a clear, concise, succinct summary of the current state of the project, including the issues that need to be completed, their priority, and any dependencies between them.

   2b. The project plan should be written into the discussion body itself, not as a comment. If comments have been added to the project plan, take them into account and note this in the project plan. Never add comments to the project plan discussion.

   2c. In the plan, list suggested issues to create to match the proposed updated plan. Don't create any issues, just list the suggestions. Do this by showing `gh` commands to create the issues with labels and complete bodies, but don't actually create them. Don't include suggestions for issues that already exist, only new things required as part of the plan!

3. Create a new planning discussion with the project plan in its body. 

   3a. Create a discussion with an appropriate title starting with "${{ github.workflow }}" and the current date (e.g., "Daily Plan - 2025-10-10"), using the project plan as the body.


