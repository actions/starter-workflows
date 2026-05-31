# GitHub Enterprise Server (GHES) Sync Policy

**Date:** 2026-05-30  
**Current Status:** Agentic workflows are excluded from GHES sync  
**Policy Owner:** GitHub Actions team

---

## Executive Summary

Agentic workflows are **not synced to GitHub Enterprise Server (GHES)** as of May 2026. This document explains why, the technical constraints, and the timeline for GHES support.

---

## Current Sync Configuration

### ✅ Workflows Synced to GHES

The `script/sync-ghes/settings.json` file defines which workflows are synchronized:

```json
{
  "folders": [
    "../../ci",           // ✅ 53 CI workflows
    "../../automation",   // ✅ 5 automation workflows
    "../../code-scanning" // ✅ 77 code-scanning workflows
    "../../pages"         // ✅ 9 pages workflows (read-only)
  ]
}
```

**Total synced:** 144 workflows (76% of 188)

### ❌ Workflows NOT Synced to GHES

```json
{
  "excluded_folders": [
    "../../agentic",     // ❌ 11 agentic workflows (NEW — not synced)
    "../../deployments"  // ❌ 29 deployment workflows (excluded separately)
  ]
}
```

**Total excluded:** 40 workflows (21% of 188)

---

## Why Agentic Workflows Are Excluded

### Technical Reasons

#### 1. **Agentic Workflows Use Features Not Available in GHES**
- **Managed Agent Execution:** Agentic workflows require GitHub's hosted Claude AI agents
- **GHES Limitation:** GitHub Enterprise Server runs on-premises without access to managed AI services
- **Timeline for Fix:** Requires GitHub to build and distribute Claude agent support for GHES
- **Status:** No public roadmap announced as of May 2026

#### 2. **Unsupported Actions in GHES**
Agentic workflows may use GitHub-hosted actions not yet available in GHES:
- `github/managed-agents/*` (hypothetical, for AI execution)
- Custom Claude-specific actions
- Cloud integration actions (AWS, Azure, GCP)

**Current GHES Enabled Actions (31 total):**
```
actions/cache
actions/checkout
actions/configure-pages
actions/create-release
actions/delete-package-versions
actions/deploy-pages
actions/download-artifact
actions/jekyll-build-pages
actions/setup-dotnet
actions/setup-go
actions/setup-java
actions/setup-node
actions/setup-python
actions/stale
actions/starter-workflows
actions/upload-artifact
actions/upload-pages-artifact
actions/upload-release-asset
github/codeql-action
```

#### 3. **Format Incompatibility (Secondary)**
- Agentic workflows use `.md` format instead of `.yml`
- GHES sync script only processes `.yml` files (line 50 in `sync-ghes/index.ts`)
- Format processing would need to be updated to support markdown workflows

#### 4. **Deployment Workflows Are Also Excluded**
- 29 deployment workflows also excluded from GHES sync
- Reason: Use cloud-specific actions (AWS, Azure, GCP, etc.) not available in GHES
- Separate from agentic workflow exclusion, but similar pattern

---

## Policy: How Workflows Are Classified

The sync script (`sync-ghes/index.ts`) applies these filters:

```typescript
// Line 62-65: A workflow is enabled for GHES if:
const enabled =
  !isPartnerWorkflow &&  // Not created by cloud partners (AWS, Azure, GCP, etc.)
  (workflowProperties.enterprise === true || 
   basename(folder) !== 'code-scanning') &&
  (await checkWorkflow(workflowFilePath, enabledActions));
    ↑
   Uses only actions in the GHES allowlist
```

### Exclusion Rules

| Rule | Folder | Count | Rationale |
|------|--------|-------|-----------|
| **Partner Workflows** | code-scanning, deployments | ~20 | Partner-specific integrations (AWS, Azure, GCP) |
| **Unsupported Actions** | deployments, agentic | ~20 | Requires cloud/AI capabilities not in GHES |
| **Format Incompatibility** | agentic | 11 | Uses `.md` instead of `.yml` |
| **Enterprise Only** | code-scanning | ~10 | Requires enterprise features |

---

## Timeline for GHES Support

### Current Phase: Preview (May 2026 - ?)
- ✅ Agentic workflows available on GitHub.com
- ✅ Documentation published (AGENTIC_WORKFLOWS.md)
- ❌ Not available on GHES
- 📊 Collecting usage metrics and feedback

### Phase 2: GHES Support (Timeline TBD)
**Prerequisite:** GitHub Ships managed agent support for GHES
- [ ] Build Claude agent execution engine for GHES
- [ ] Distribute agent runtime to GHES releases
- [ ] Update sync script to include agentic folder
- [ ] Add agentic actions to GHES enabled actions list
- [ ] Test on GHES 23.4+ (estimated)

**Estimated Timeline:** Q4 2026 – Q2 2027 (speculative, no public commitment)

### Phase 3: General Availability (Timeline TBD)
- [ ] Agentic workflows promoted to "stable" category
- [ ] Full GHES feature parity with GitHub.com
- [ ] Comprehensive GHES documentation
- [ ] Enterprise support and SLAs in place

---

## How to Request GHES Support

If you need agentic workflows on your GitHub Enterprise Server instance:

1. **Contact GitHub Support:** Open a support ticket requesting agentic workflow support for GHES
   - Reference your GHES version
   - Describe your use case
   - Helps GitHub prioritize GHES feature requests

2. **Vote on GitHub Roadmap:** If GitHub opens voting for agentic GHES support
   - Signal demand to GitHub's product team
   - Links: https://github.com/orgs/github/discussions

3. **Workaround for GHES Users:**
   - Use traditional CI/code-scanning workflows (fully supported)
   - Manually create workflows inspired by agentic templates
   - Implement external agent execution (outside GitHub Actions)

---

## Deployment Workflows: Similar Exclusion Pattern

**Note:** Deployment workflows (29 total) are also excluded from GHES, for related reasons:

| Workflow | Status | Reason |
|----------|--------|--------|
| AWS workflows | ❌ Excluded | AWS action not in GHES enabled list |
| Azure workflows | ❌ Excluded | Azure action not in GHES enabled list |
| GCP workflows | ❌ Excluded | GCP action not in GHES enabled list |
| Generic Deploy (Kubernetes) | ✅ Synced | Uses `actions/checkout` only |

**Implication:** Enterprise customers relying on cloud deployments need to:
- Create custom workflows using only GHES-compatible actions
- Use external CI/CD systems alongside GitHub Actions
- Wait for cloud provider actions to be added to GHES enabled list

---

## Technical Details: The Sync Process

### Sync Script Location
- **Script:** `script/sync-ghes/index.ts`
- **Config:** `script/sync-ghes/settings.json`
- **Purpose:** Periodically syncs compatible workflows from `main` to `ghes` branch

### Sync Algorithm

```
1. Load settings.json (enabled actions, folder list)
2. For each folder in config:
   a. Find all .yml files
   b. Load corresponding .properties.json
   c. Check if workflow uses only enabled actions
   d. Classify as "compatible" or "incompatible"
3. Checkout ghes branch
4. Delete all workflows from old sync
5. Restore read-only folders (pages)
6. Copy compatible workflows from main
7. Update artifact action versions (v4 → v3 for GHES)
```

### Why Agentic Isn't Listed
- Never added to `settings.json` "folders" array
- Sync script doesn't process `.md` files (only `.yml`)
- Even if listed, would be filtered out due to unsupported actions

---

## For Repository Maintainers

### Adding Support for New GHES Actions

If GitHub enables new actions for GHES:

1. **Update `script/sync-ghes/settings.json`:**
   ```json
   "enabledActions": [
     "actions/new-action",
     // ... existing actions
   ]
   ```

2. **If new folder needs syncing:**
   ```json
   "folders": [
     "../../agentic",  // Add if/when agentic is ready
     // ... existing folders
   ]
   ```

3. **If format changes (e.g., support `.md`):**
   - Update line 50 in `sync-ghes/index.ts`
   - Change from `extname(e.name) === ".yml"` to also accept `.md`
   - Add `.md` → `.yml` conversion logic (if needed)

4. **Test the sync:**
   ```bash
   cd script/sync-ghes
   npx ts-node index.ts
   git status  # Verify correct workflows synced
   git diff ghes/main  # Review changes
   ```

---

## FAQ

### Q: Why isn't my agentic workflow available in our GHES instance?
**A:** Agentic workflows require Claude agent execution, which isn't available in GHES yet. This is a feature gap GitHub is working on. Use traditional CI/code-scanning workflows in the meantime.

### Q: When will agentic workflows be available on GHES?
**A:** No official timeline. Monitor the [GitHub roadmap](https://github.com/orgs/github/discussions) for announcements. Contact GitHub Support if this is critical for your organization.

### Q: Can we run agentic workflows on-premises?
**A:** Not yet. Agentic workflows require GitHub-hosted Claude agents. Self-hosted or on-premises execution is not currently supported.

### Q: How do we request this feature?
**A:** File a support ticket with GitHub. Reference your GHES version and use case. This helps GitHub prioritize feature development.

### Q: What about the deployment workflows that are excluded?
**A:** Same pattern — they require cloud actions not available in GHES. As cloud providers add official GHES support, those workflows can be synced.

---

## Recommendation

### Current Status: Informational
This is an **informational document** not a change request. The current exclusion policy is intentional and necessary.

### Future Actions (When GHES Support Is Announced)
1. **Announce in Docs:** Add a note to AGENTIC_WORKFLOWS.md about GHES timeline
2. **Update README:** Clarify GHES support status in main directory structure
3. **Update Sync Script:** Add agentic folder and actions when ready
4. **Test Thoroughly:** Ensure agentic workflows work identically on GHES

### Recommended Documentation Updates (Now)
- [ ] Add GHES support timeline to AGENTIC_WORKFLOWS.md
- [ ] Add note in README under agentic section
- [ ] Create internal tracking issue for GHES readiness

---

## Tracking & Monitoring

### Open Questions
- [ ] What is GitHub's official timeline for agentic GHES support?
- [ ] Will agentic agents run self-hosted or only GitHub-hosted?
- [ ] Will deployment workflows get cloud action support on GHES?
- [ ] Are there alternative approaches for GHES users?

### How to Track
- Monitor [GitHub roadmap](https://github.com/orgs/github/discussions)
- Watch for releases of `actions/managed-agents` or similar
- Subscribe to GitHub blog for feature announcements
