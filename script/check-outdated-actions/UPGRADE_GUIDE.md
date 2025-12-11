# Action upgrade guide

This guide documents best practices and learnings from upgrading actions in the starter-workflows repository.

## General upgrade process

1. **Research breaking changes first** - Always check release notes, changelog, and README for the action
2. **Analyze usage context** - Review how the action is used across all workflow files
3. **Test compatibility** - Ensure breaking changes don't affect the specific usage patterns
4. **Update systematically** - Use bulk operations when safe, manual review when necessary

## Version format patterns

Actions in this repository use several version formats. **Always preserve the specificity level** when upgrading:

| Format | Example | Specificity | Upgrade example |
|--------|---------|-------------|-----------------|
| Major only | `@v5` | Major | `@v6` |
| Major.minor | `@v5.2` | Minor | `@v6.0` |
| Full semver | `@v5.2.1` | Patch | `@v6.0.1` |
| SHA only | `@abc123...` | Pinned (assume patch) | `@newsha...` |
| SHA + major comment | `@abc123... # v4` | Major | `@newsha... # v6` |
| SHA + patch comment | `@abc123... # v4.2.2` | Patch | `@newsha... # v6.0.1` |

### SHA with version comment format

This is a common pattern for security-conscious workflows:

```yaml
uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
```

**The comment indicates the specificity level.** Here `# v4.2.2` means patch-level specificity.

When upgrading, update BOTH the SHA AND the comment, preserving specificity:

```yaml
uses: actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8 # v6.0.1
```

**NOT** `# v6` — that would lose specificity.

## Case study: actions/checkout v5 → v6

**Date:** December 2025  
**Files updated:** 161 files, 170 instances  
**Upgrade verdict:** ✅ Safe for all standard usage

### Breaking changes identified

1. **Credential storage location**
   - v5: Stored credentials directly in `.git/config`
   - v6: Stores credentials in separate file under `$RUNNER_TEMP`
   - **Impact:** Internal change, transparent to users
   - **Action required:** None for standard workflows

2. **Docker container actions requirement**
   - v6 requires Actions Runner v2.329.0+ for authenticated git commands from Docker container actions
   - **Impact:** Affects only workflows using Docker container actions with git operations
   - **Action required:** None for starter-workflows (no Docker container actions using git auth)

### Key findings

- **No workflow syntax changes required** - All existing checkout configurations work identically
- **Git commands continue to work** - `git fetch`, `git push`, etc. work automatically as before
- **Backward compatible** - All standard checkout scenarios maintained
- **Version specificity matters** - Match existing patterns (major-only versions stay major-only)

### Update approach used

**Important:** Always match the existing version specificity in each file:

- If using `@v5` → update to `@v6` (major version only)
- If using `@v5.2.0` → update to `@v6.0.1` (full semantic version)
- If using SHA with comment `# v4.2.2` → update SHA AND comment to `# v6.0.1` (preserve patch-level)
- If using SHA with comment `# v5` → update SHA AND comment to `# v6` (preserve major-only)

**⚠️ Common pitfall:** Using naive patterns like `s/@v5/@v6/g` will incorrectly match `@v5.2.0` and turn it into `@v6.2.0`!

**Correct approach - Check specificity first, then update accordingly:**

1. **Identify what specificity levels exist:**

   ```bash
   # See what version formats are actually used
   grep -r "actions/checkout@" --include="*.yml" | \
     sed 's/.*actions\/checkout@//' | \
     sed 's/\([0-9a-f]\{40\}\).*/SHA: \1/' | \
     sed 's/\(v[0-9.]*\).*/\1/' | \
     sort | uniq -c | sort -rn
   ```

   Example output:

   ```text
   170 v5
     1 SHA: 11bd71901bbe5b1630ceea73d27597364c9af683
   ```

   For SHA-pinned versions, also check the comment to determine specificity:

   ```bash
   grep -r "actions/checkout@[0-9a-f]\{40\}" --include="*.yml"
   ```

2. **Update each specificity level separately:**

   ```bash
   # For major-only versions (@v5 → @v6)
   # Use end-of-line OR followed by space/quote/etc (not a dot or digit)
   find . -name "*.yml" -type f -exec sed -i '' \
     's/\(actions\/checkout@\)v5\($\| \|"\|'\''`\)/\1v6\2/g' {} \;
   
   # For full semantic versions (@v5.x.y → @v6.0.1)
   find . -name "*.yml" -type f -exec sed -i '' \
     's/actions\/checkout@v5\.[0-9]\+\.[0-9]\+/actions\/checkout@v6.0.1/g' {} \;
   
   # For minor versions (@v5.x → @v6.0)
   find . -name "*.yml" -type f -exec sed -i '' \
     's/\(actions\/checkout@\)v5\.\([0-9]\+\)\($\| \|"\|'\''`\)/\1v6.0\3/g' {} \;
   ```

3. **SHA-pinned versions:** Update manually (can't be done with regex reliably)
   - Find SHA for desired version at <https://github.com/actions/checkout/releases>
   - v6.0.1 SHA: `8e8c483db84b4bee98b60c0593521ed34d9990e8`
   - **Preserve comment specificity:**
     - `@oldsha # v4.2.2` → `@newsha # v6.0.1` (patch-level stays patch-level)
     - `@oldsha # v5` → `@newsha # v6` (major stays major)
   - Example:

     ```yaml
     # Before
     uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
     # After
     uses: actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8 # v6.0.1
     ```

### Verification steps

```bash
# 1. Verify old references are completely gone
grep -r "checkout@v5" --include="*.yml"  # Should return nothing

# 2. Check new version distribution matches old distribution
grep -r "actions/checkout@" --include="*.yml" | \
  sed 's/.*actions\/checkout@//' | sed 's/[^0-9.v].*//' | \
  sort | uniq -c | sort -rn
# Compare counts: if you had 170 @v5, you should now have 170 @v6
# If you had 3 @v5.2.0, you should now have 3 @v6.0.1

# 3. Verify no invalid versions created (e.g., @v6.2.0 which doesn't exist)
grep -r "checkout@v6\.[1-9]" --include="*.yml"  # Should return nothing
grep -r "checkout@v6\.0\.[2-9]" --include="*.yml"  # Should return nothing

# 4. Check git diff for any unexpected changes
git diff --stat
git diff | grep "checkout@" | head -20  # Spot check actual changes
```

## Best practices for future upgrades

### ✅ Do

- **Read the changelog thoroughly** - Look for sections like "Breaking Changes", "Migration Guide", "What's New"
- **Check minimum requirements** - Runner versions, Node.js versions, OS compatibility
- **Look for credential/auth changes** - These often have security implications
- **Verify SHA matches version** - When using SHA-pinned versions, ensure SHA is correct for the version
- **Update comments alongside SHAs** - Keep version comments in sync with SHA references
- **Test in batches** - For large changes, consider updating a subset first
- **Document your findings** - Update this guide with new learnings

### ❌ Don't

- **Use partial string matches in sed** - `s/@v5/@v6/g` will incorrectly turn `@v5.2.0` into `@v6.2.0`!
- **Assume semantic versions are safe** - Even patch updates can have breaking changes
- **Ignore runner requirements** - Runner version requirements can break workflows on older runners
- **Skip reading release notes** - They contain critical migration information
- **Update blindly** - Each action may have unique breaking changes
- **Forget about SHA-pinned versions** - These need manual updates and comment changes

## Common breaking change patterns

### 1. Node.js runtime updates

- **Example:** v4 → v5 for many actions (Node 16 → Node 20)
- **Impact:** Requires minimum runner version
- **Check:** Look for "runtime" or "Node.js" in changelog

### 2. Input/output changes

- **Example:** Input renamed, new required input, output format changed
- **Impact:** Workflows using these inputs/outputs break
- **Check:** Review action.yml or documentation for input/output definitions

### 3. Authentication/credential changes

- **Example:** checkout v6 credential storage location
- **Impact:** May affect Docker containers, credential persistence
- **Check:** Look for "auth", "credentials", "token" changes

### 4. Behavior changes

- **Example:** Default values changed, new default behavior
- **Impact:** Silent behavior changes that may not cause failures
- **Check:** Look for "default" changes in changelog

### 5. Dependency requirements

- **Example:** Minimum Git version, Python version, etc.
- **Impact:** May not work on certain runners or containers
- **Check:** Look for "requirements", "minimum version" sections

## Resources

- [GitHub Actions changelog](https://github.blog/changelog/label/actions/)
- [actions/checkout releases](https://github.com/actions/checkout/releases)
- [Runner release notes](https://github.com/actions/runner/releases)

## Template for documenting upgrades

When upgrading a new action, add a case study in your commit or pull request description using this template:

```markdown
## Case study: {action-name} {old-version} → {new-version}

**Date:** {Month Year}
**Files updated:** {X} files, {Y} instances
**Upgrade verdict:** ✅ Safe / ⚠️ Review required / ❌ Breaking changes

### Breaking changes identified

1. **{Change title}**
   - Old behavior: {description}
   - New behavior: {description}
   - **Impact:** {who/what is affected}
   - **Action required:** {what needs to be done}

### Key findings

- {Finding 1}
- {Finding 2}

### Update approach used

{Commands or process used}

### Verification steps

{Commands used to verify}
```
