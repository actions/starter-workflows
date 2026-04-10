---
description: Daily automated link checker that finds and fixes broken links in documentation files
on:
  schedule: daily on weekdays
permissions: read-all
timeout-minutes: 60
network:
  allowed:
    - node
    - python
    - github
steps:
  - name: Checkout repository
    uses: actions/checkout@v4
    with:
      fetch-depth: 0
      persist-credentials: false

  - name: Check and test all documentation links
    id: link-check
    run: |
      echo "# Link Check Results" > /tmp/link-check-results.md
      echo "" >> /tmp/link-check-results.md
      
      # Find all markdown files in docs directory and README
      echo "Finding all markdown files..."
      MARKDOWN_FILES=$(find docs README.md -type f -name "*.md" 2>/dev/null || echo "")
      
      if [ -z "$MARKDOWN_FILES" ]; then
        echo "No markdown files found"
        echo "no_files=true" >> $GITHUB_OUTPUT
        exit 0
      fi
      
      # Extract all links from markdown files
      echo "## Links Found" >> /tmp/link-check-results.md
      echo "" >> /tmp/link-check-results.md
      
      # Use grep to find markdown links and HTTP(S) URLs
      for file in $MARKDOWN_FILES; do
        echo "Checking $file..."
        # Extract markdown links [text](url)
        grep -oP '\[([^\]]+)\]\(([^\)]+)\)' "$file" | grep -oP '\(([^\)]+)\)' | tr -d '()' >> /tmp/all-links.txt 2>/dev/null || true
        # Extract plain HTTP(S) URLs
        grep -oP 'https?://[^\s<>"]+' "$file" >> /tmp/all-links.txt 2>/dev/null || true
      done
      
      # Remove duplicates and sort
      if [ -f /tmp/all-links.txt ]; then
        sort -u /tmp/all-links.txt > /tmp/unique-links.txt
        LINK_COUNT=$(wc -l < /tmp/unique-links.txt)
        echo "Found $LINK_COUNT unique links" >> /tmp/link-check-results.md
        echo "" >> /tmp/link-check-results.md
      else
        echo "No links found" >> /tmp/link-check-results.md
        echo "no_links=true" >> $GITHUB_OUTPUT
        exit 0
      fi
      
      # Test each link
      echo "## Link Test Results" >> /tmp/link-check-results.md
      echo "" >> /tmp/link-check-results.md
      echo "Testing links..." >> /tmp/link-check-results.md
      
      BROKEN_COUNT=0
      WORKING_COUNT=0
      
      while IFS= read -r url; do
        # Skip relative links and anchors
        if [[ "$url" == "#"* ]] || [[ "$url" != "http"* ]]; then
          continue
        fi
        
        # Test the link with curl
        HTTP_CODE=$(curl -L -s -o /dev/null -w "%{http_code}" --max-time 10 "$url" 2>/dev/null || echo "000")
        
        if [[ "$HTTP_CODE" =~ ^2 ]] || [[ "$HTTP_CODE" =~ ^3 ]]; then
          WORKING_COUNT=$((WORKING_COUNT + 1))
          echo "✅ $url (HTTP $HTTP_CODE)" >> /tmp/link-check-results.md
        else
          BROKEN_COUNT=$((BROKEN_COUNT + 1))
          echo "❌ $url (HTTP $HTTP_CODE)" >> /tmp/link-check-results.md
        fi
      done < /tmp/unique-links.txt
      
      echo "" >> /tmp/link-check-results.md
      echo "**Summary:** $WORKING_COUNT working, $BROKEN_COUNT broken" >> /tmp/link-check-results.md
      
      # Output results
      echo "broken_count=$BROKEN_COUNT" >> $GITHUB_OUTPUT
      echo "working_count=$WORKING_COUNT" >> $GITHUB_OUTPUT
      
      cat /tmp/link-check-results.md
    shell: bash

tools:
  github:
    toolsets: [default]
  cache-memory: true
  web-fetch:

safe-outputs:
  create-pull-request:
    title-prefix: "[link-checker] "
    labels: [documentation, automated]
    draft: false
    protected-files: fallback-to-issue
    if-no-changes: "warn"
  noop:
---

# Daily Link Checker & Fixer

You are an automated link checker and fixer agent. Your job is to find and fix broken links in the documentation files of this repository.

## Your Mission

Your workflow has already collected and tested all links in the previous step. Use the test results to identify broken links and fix them where possible.

## Step 1: Review Link Check Results

The link check step has already run and created a report at `/tmp/link-check-results.md`. Read this file to see:
- All links found in the documentation
- Which links are working (✅) and which are broken (❌)
- HTTP status codes for each link

Use bash to read the file:
```bash
cat /tmp/link-check-results.md
```

## Step 2: Load Cache Memory

Check cache memory for previously identified unfixable broken links:
- Load the cache memory to see if there are any broken links we've tried to fix before but couldn't
- These are links that are permanently broken or removed from the internet
- Skip these links to avoid repeated attempts

The cache memory should store a JSON object with this structure:
```json
{
  "unfixable_links": [
    {
      "url": "https://example.com/removed-page",
      "reason": "404 Not Found - content removed",
      "first_seen": "2026-02-17"
    }
  ],
  "last_run": "2026-02-17"
}
```

## Step 3: Research and Fix Broken Links

For each broken link found in the test results (but NOT in the unfixable list):

1. **Investigate the link:**
   - Determine what the link was supposed to point to based on:
     - The link text in the markdown
     - The context around the link
     - The surrounding documentation
   
2. **Search for alternatives:**
   - Use web-fetch to search for if the content has moved to a new URL
   - Try common alternatives (www vs non-www, http vs https, with/without trailing slash)
   - Look for redirects or updated documentation
   - Check if there's an official replacement
   
3. **Fix the link:**
   - If you find a working replacement URL, use the `edit` tool to update the markdown file
   - Replace the broken URL with the working one
   - Make sure to preserve the link text and formatting
   
4. **Document unfixable links:**
   - If a link truly cannot be fixed (content permanently removed, no alternatives found):
     - Add it to the unfixable_links list in cache memory
     - Include the URL, reason, and date
     - This prevents future runs from wasting time on the same broken link

## Step 4: Update Cache Memory

After processing all broken links:
- Update the cache memory with any new unfixable links
- Update the "last_run" timestamp
- Save the updated cache memory

## Step 5: Create Pull Request or Noop

Based on your work:

**If you fixed any links:**
- Use the `create-pull-request` safe output to create a PR with your fixes
- In the PR body, include:
  - A summary of how many links were fixed
  - A list of the broken links and their replacements
  - Any links that were added to the unfixable list
- Title format: "Fix broken documentation links"

**If no links needed fixing:**
- Use the `noop` safe output with a clear message like:
  - "All documentation links are working correctly" (if no broken links found)
  - "All broken links are in the unfixable list, no new fixes available" (if broken links exist but can't be fixed)

## Important Guidelines

- **Be thorough:** Check every broken link carefully
- **Preserve context:** When replacing links, make sure the new URL points to equivalent or better content
- **Document everything:** Keep the cache memory up to date with unfixable links
- **Be selective:** Only add links to the unfixable list if you've genuinely tried to find alternatives
- **Use web-fetch wisely:** Try to fetch the broken URL and check for redirects or alternatives
- **Relative links:** Focus only on HTTP(S) links. Skip relative links and anchors (they're tested differently)

## Example Cache Memory Update

```json
{
  "unfixable_links": [
    {
      "url": "https://old-docs.example.com/api/v1",
      "reason": "Documentation site shut down, no replacement found despite searching",
      "first_seen": "2026-02-17"
    }
  ],
  "last_run": "2026-02-17"
}
```

## Context

- Repository: `${{ github.repository }}`
- Run daily on weekdays to catch broken links early
- Link test results are available at `/tmp/link-check-results.md`
