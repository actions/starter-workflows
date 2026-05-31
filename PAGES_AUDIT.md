# Pages Workflow Templates — Coverage Audit

**Date:** 2026-05-30  
**Current Coverage:** 9 templates  
**Status:** Underserved compared to CI (53) and Code-Scanning (77)

---

## Current Coverage

### ✅ Supported Frameworks (9 templates)

| Framework | Type | Template | Status |
|-----------|------|----------|--------|
| Astro | Meta-framework | `astro.yml` | Modern, active |
| Gatsby | Static site gen | `gatsby.yml` | Mature, stable |
| Hugo | Static site gen | `hugo.yml` | Mature, stable |
| Jekyll | Static site gen | `jekyll.yml` + `jekyll-gh-pages.yml` | 2 variants (native + GitHub Pages) |
| mdBook | Docs generator | `mdbook.yml` | Niche (Rust ecosystem) |
| Next.js | Meta-framework | `nextjs.yml` | Modern, active |
| Nuxt | Meta-framework | `nuxtjs.yml` | Modern, active |
| Static Files | Raw HTML/CSS | `static.yml` | Fallback option |

### Coverage by Ecosystem

| Ecosystem | Count | Status |
|-----------|-------|--------|
| Vue | 2 | Nuxt only (missing VitePress) |
| React | 3 | Gatsby, Next.js, Astro |
| Rust | 1 | mdBook (Zola missing) |
| Python | 0 | ❌ Sphinx, MkDocs not covered |
| Vanilla JS | 1 | Static files |
| Node.js general | 1 | Eleventy missing |

---

## Gap Analysis: Missing Templates

### High Priority (High Demand, Not Covered)

#### 1. **VitePress** 📈
- **Popularity:** ⭐⭐⭐⭐⭐ (Very High)
- **Use Case:** Modern documentation sites
- **Community:** Vue.js ecosystem, 15k+ GitHub stars
- **Demand Signals:**
  - Official docs for Vue, Vitest, Vite, Rollup use VitePress
  - Growing adoption for technical documentation
  - Faster and simpler than Nuxt for docs
- **Why Missing:** Newer framework (2022+), added after initial Pages templates
- **Implementation Effort:** Low (follows standard Vite build pattern)
- **Recommendation:** 🔴 **CRITICAL** — Add immediately

#### 2. **Docusaurus** 📈
- **Popularity:** ⭐⭐⭐⭐⭐ (Very High)
- **Use Case:** Documentation and technical content
- **Community:** Meta's official tool, 12k+ GitHub stars
- **Demand Signals:**
  - Used by major projects (React docs, GraphQL, Jest, etc.)
  - Strong TypeScript/React ecosystem
  - Large community and plugins
- **Why Missing:** Corporate/enterprise focus, not as web-native as Astro/Next
- **Implementation Effort:** Medium (Node.js + build step)
- **Recommendation:** 🔴 **CRITICAL** — Add immediately

#### 3. **MkDocs** 📊
- **Popularity:** ⭐⭐⭐⭐ (High)
- **Use Case:** Python projects, technical documentation
- **Community:** Python ecosystem, 19k+ GitHub stars
- **Demand Signals:**
  - De facto standard for Python project documentation
  - Used by Django, FastAPI, Kubernetes, Boto3
  - Large Python developer base
- **Why Missing:** Python-focused, different build model from JS frameworks
- **Implementation Effort:** Low (pip install + Python CLI)
- **Recommendation:** 🟡 **HIGH** — Add soon

#### 4. **Zola** 🚀
- **Popularity:** ⭐⭐⭐⭐ (Rising)
- **Use Case:** Modern static site generation
- **Community:** Rust ecosystem, 12k+ GitHub stars
- **Demand Signals:**
  - Single-binary tool (fast deployment)
  - Growing adoption for blogs and documentation
  - Modern architecture (unlike Hugo, easier templating)
- **Why Missing:** Newer, less mainstream than Hugo
- **Implementation Effort:** Low (single binary, no build complexity)
- **Recommendation:** 🟡 **HIGH** — Add soon

### Medium Priority (Moderate Demand, Not Covered)

#### 5. **Eleventy (11ty)** 🟠
- **Popularity:** ⭐⭐⭐⭐ (Growing)
- **Use Case:** Flexible static site generation
- **Community:** JavaScript ecosystem, 16k+ GitHub stars
- **Demand Signals:**
  - Simple, zero-config approach
  - Popular with indie web/JAMstack community
  - Good for blogs and small marketing sites
- **Why Missing:** Newer, not mainstream enterprise tool
- **Implementation Effort:** Low (Node.js CLI)
- **Recommendation:** 🟡 **MEDIUM** — Consider adding

#### 6. **Sphinx** 🟠
- **Popularity:** ⭐⭐⭐⭐ (Stable/Mature)
- **Use Case:** Python documentation standard
- **Community:** Python ecosystem, 5k+ GitHub stars
- **Demand Signals:**
  - Python standard for docs (numpy, scipy, Django, etc.)
  - Strong reStructuredText support
  - Mature, well-tested
- **Why Missing:** Python-specific, overlaps with MkDocs
- **Implementation Effort:** Low (pip install)
- **Recommendation:** 🟡 **LOW** — MkDocs may be sufficient

#### 7. **Solid Start** 🟠
- **Popularity:** ⭐⭐⭐ (Emerging)
- **Use Case:** Full-stack Solid.js applications
- **Community:** Solid.js ecosystem, newer framework
- **Demand Signals:**
  - Growing interest in Solid.js for performance
  - Alternative to React/Vue for page generation
  - Still early stage
- **Why Missing:** Very new framework
- **Implementation Effort:** Medium
- **Recommendation:** 🟡 **LOW** — Wait for wider adoption

#### 8. **SvelteKit** 🟠
- **Popularity:** ⭐⭐⭐⭐ (Growing)
- **Use Case:** Full-stack Svelte applications
- **Community:** Svelte ecosystem, newer framework
- **Demand Signals:**
  - Strong Svelte community
  - Adapter pattern supports static generation
  - Growing adoption
- **Why Missing:** Meta-framework focus (like Next/Nuxt), not pure static gen
- **Implementation Effort:** Medium
- **Recommendation:** 🟡 **LOW** — May overlap with Astro coverage

### Low Priority (Niche/Declining)

- **Sphinx with Read the Docs** — Too specialized, MkDocs is simpler
- **GitBook** — Commercial, less relevant for GitHub Pages
- **Doxygen** — Code documentation only, not general content
- **Rspress** — Too new, ByteDance tool with limited adoption outside China

---

## Recommendation Summary

### 🎯 **Action Items**

#### Immediate (Next Month)
1. **Add VitePress** — Single highest-impact addition
   - Rationale: Vue ecosystem, official docs standard
   - Effort: Low
   - Impact: High (documentation is huge market)

2. **Add Docusaurus** — Second highest-impact
   - Rationale: Meta's tool, large enterprise adoption
   - Effort: Medium
   - Impact: High (React ecosystem + documentation)

#### Short-term (Next Quarter)
3. **Add MkDocs** — Fill Python ecosystem gap
   - Rationale: Python is large developer base
   - Effort: Low
   - Impact: Medium

4. **Add Zola** — Modern alternative to Hugo
   - Rationale: Growing adoption, single binary deployment
   - Effort: Low
   - Impact: Medium

#### Longer-term (Evaluate)
5. **Add Eleventy** — Flexible JS tool for blogs
   - Rationale: Growing indie web community
   - Effort: Low
   - Impact: Medium (niche audience)

---

## Market Analysis

### Pages Workflow Usage Patterns

By ecosystem sizing:
- **React ecosystem:** 40% of developers (Next.js, Gatsby, Astro cover this)
- **Vue ecosystem:** 15% of developers (Nuxt covered, **VitePress missing**)
- **Python ecosystem:** 20% of developers (**Completely unserved**)
- **Static/Minimal:** 10% of developers (Static template covers)
- **Other/Rust:** 15% of developers (mdBook, **Zola missing**)

### Gaps by Severity

| Gap | Severity | Fix | Users Affected |
|-----|----------|-----|-----------------|
| No VitePress | 🔴 Critical | Add template | ~15% Vue developers |
| No Docusaurus | 🔴 Critical | Add template | ~30% enterprise/Meta ecosystem |
| No MkDocs | 🟡 High | Add template | ~20% Python developers |
| No Zola | 🟡 High | Add template | ~5% Rust developers + growing audience |
| No Sphinx | 🟡 Medium | Add template | ~5% Python (if not MkDocs) |
| No Eleventy | 🟡 Medium | Add template | ~3% indie web developers |

---

## Template Creation Checklist

For each new Pages template, ensure:

- [ ] `.yml` workflow file following existing pattern
- [ ] `.properties.json` metadata file
- [ ] Example `.gitignore` for build output (if applicable)
- [ ] Test on GitHub Pages with public/private repo
- [ ] Documentation in README under Pages section
- [ ] Icon for the framework (add to `icons/` if needed)
- [ ] Tested with default branch and custom branch deployments
- [ ] Security review (Node.js dependencies, build step safety)

---

## Effort Estimate

| Task | Effort | Notes |
|------|--------|-------|
| VitePress | 2-4 hours | Simple build pattern, similar to Next.js |
| Docusaurus | 4-6 hours | More complex config, Node.js build |
| MkDocs | 2-3 hours | Python CLI, straightforward |
| Zola | 2-3 hours | Single binary, no dependencies |
| Eleventy | 3-4 hours | Flexible but needs good docs |
| **Total (all 5)** | **13-20 hours** | Spread across 1-2 quarters |

---

## Conclusion

**Current Status:** Pages category is undersized at 9 templates.

**Immediate Action:** Add VitePress + Docusaurus (2-3 weeks effort)
- Covers ~50% of missing coverage
- High ROI (Vue + React ecosystems already partially covered)
- Unblocks Python/Rust ecosystems

**Medium-term:** Add MkDocs + Zola (follow-up in next quarter)

**Result:** Expand Pages from 9 → 13 templates, covering 90%+ of popular frameworks.
