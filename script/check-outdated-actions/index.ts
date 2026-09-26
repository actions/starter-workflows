#!/usr/bin/env npx ts-node
import { promises as fs } from "fs";
import path from "path";

// Major and minor tags move. v5 stays current for every v5.x.y release.
// v5.2 stays current for every v5.2.y release. A patch tag does not move.
// A commit SHA with no version comment does not move. It is current only when
// that commit is the highest stable release. The README cites the GitHub
// release documentation this follows.

const REPO_ROOT = path.resolve(__dirname, "../..");
const USES_LINE =
  /uses:\s*['"]?actions\/([A-Za-z0-9_.-]+)((?:\/[A-Za-z0-9_.-]+)*)@([^'"\s#]+)['"]?(?:\s+#\s*(.*))?/;
const VERSION_IN_TEXT = /v(\d+(?:\.\d+){0,2})\b/;
const SHA = /^[0-9a-f]{40}$/i;

type Specificity = "major" | "minor" | "patch";

interface Version {
  major: number;
  minor: number;
  patch: number;
  specificity: Specificity;
  tag: string;
}

interface Reference {
  file: string;
  line: number;
  repo: string;
  uses: string;
  ref: string;
  comment: string;
}

interface Finding {
  file: string;
  line: number;
  message: string;
}

interface ReleaseCatalog {
  newest: Version | null;
  newestCommit: string | null;
  byCommit: Map<string, Version>;
}

const SPECIFICITY_RANK: Record<Specificity, number> = {
  major: 1,
  minor: 2,
  patch: 3,
};

export function parseVersion(tag: string): Version | null {
  const match = /^v(\d+)(?:\.(\d+)(?:\.(\d+))?)?$/i.exec(tag);
  if (!match) {
    return null;
  }
  const specificity: Specificity =
    match[3] !== undefined ? "patch" : match[2] !== undefined ? "minor" : "major";
  return {
    major: Number(match[1]),
    minor: match[2] !== undefined ? Number(match[2]) : 0,
    patch: match[3] !== undefined ? Number(match[3]) : 0,
    specificity,
    tag: `v${match[1]}${match[2] !== undefined ? `.${match[2]}` : ""}${match[3] !== undefined ? `.${match[3]}` : ""}`,
  };
}

export function compareVersions(left: Version, right: Version): number {
  if (left.major !== right.major) {
    return left.major - right.major;
  }
  if (left.minor !== right.minor) {
    return left.minor - right.minor;
  }
  return left.patch - right.patch;
}

export function preferVersion(candidate: Version, current: Version): Version {
  const compared = compareVersions(candidate, current);
  if (compared > 0) {
    return candidate;
  }
  if (compared < 0) {
    return current;
  }
  return SPECIFICITY_RANK[candidate.specificity] > SPECIFICITY_RANK[current.specificity]
    ? candidate
    : current;
}

// A reference is behind when a stable release is newer at the specificity
// the workflow wrote down. Comparing against the single highest release is
// enough: any newer major, minor, or patch is represented by that release.
export function isOutdated(pin: Version, newest: Version): boolean {
  if (pin.specificity === "major") {
    return newest.major > pin.major;
  }
  if (pin.specificity === "minor") {
    return newest.major > pin.major || (newest.major === pin.major && newest.minor > pin.minor);
  }
  return compareVersions(newest, pin) > 0;
}

function versionFromComment(comment: string): Version | null {
  const match = VERSION_IN_TEXT.exec(comment.trim());
  if (!match) {
    return null;
  }
  return parseVersion(match[0]);
}

function headers(): Record<string, string> {
  const token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN || "";
  const result: Record<string, string> = {
    Accept: "application/vnd.github+json",
    "User-Agent": "starter-workflows-check-outdated-actions",
    "X-GitHub-Api-Version": "2022-11-28",
  };
  if (token) {
    result.Authorization = `Bearer ${token}`;
  }
  return result;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function githubGet(apiPath: string): Promise<unknown> {
  const url = `https://api.github.com${apiPath}`;
  for (let attempt = 0; attempt < 4; attempt++) {
    const response = await fetch(url, { headers: headers() });
    if (response.status === 403 || response.status === 429) {
      const remaining = response.headers.get("x-ratelimit-remaining");
      if (remaining === "0") {
        throw new Error(
          "GitHub API rate limit exceeded. Set GITHUB_TOKEN or GH_TOKEN to a token that can read public repositories. Unauthenticated requests allow 60 per hour."
        );
      }
      const retryAfter = Number(response.headers.get("retry-after") || "2");
      await sleep((retryAfter + attempt) * 1000);
      continue;
    }
    if (!response.ok) {
      const body = await response.text();
      throw new Error(`GitHub API ${response.status} for ${apiPath}: ${body.slice(0, 300)}`);
    }
    return response.json();
  }
  throw new Error(`GitHub API retries exhausted for ${apiPath}`);
}

async function listStableReleaseTags(repo: string): Promise<string[]> {
  const tags: string[] = [];
  for (let page = 1; page < 20; page++) {
    const releases = (await githubGet(
      `/repos/actions/${repo}/releases?per_page=100&page=${page}`
    )) as Array<{ tag_name?: string; draft?: boolean; prerelease?: boolean }>;
    for (const release of releases) {
      if (release.draft || release.prerelease || !release.tag_name) {
        continue;
      }
      tags.push(release.tag_name);
    }
    if (releases.length < 100) {
      break;
    }
  }
  return tags;
}

interface GitRef {
  ref: string;
  object: { type: string; sha: string };
}

async function listTagRefs(repo: string): Promise<GitRef[]> {
  const refs: GitRef[] = [];
  const seen = new Set<string>();
  for (let page = 1; page < 20; page++) {
    const pageRefs = (await githubGet(
      `/repos/actions/${repo}/git/matching-refs/tags?per_page=100&page=${page}`
    )) as GitRef[];
    let added = 0;
    for (const ref of pageRefs) {
      if (seen.has(ref.ref)) {
        continue;
      }
      seen.add(ref.ref);
      refs.push(ref);
      added++;
    }
    if (pageRefs.length < 100 || added === 0) {
      break;
    }
  }
  return refs;
}

async function commitForAnnotatedTag(repo: string, tagObjectSha: string): Promise<string> {
  const tag = (await githubGet(`/repos/actions/${repo}/git/tags/${tagObjectSha}`)) as {
    object?: { sha?: string };
  };
  if (!tag.object?.sha) {
    throw new Error(`Annotated tag ${tagObjectSha} on actions/${repo} has no commit`);
  }
  return tag.object.sha.toLowerCase();
}

async function catalogFor(repo: string, needCommits: boolean): Promise<ReleaseCatalog> {
  const releaseTags = await listStableReleaseTags(repo);
  let newest: Version | null = null;
  const parsedByTag = new Map<string, Version>();
  for (const tag of releaseTags) {
    const version = parseVersion(tag);
    if (!version) {
      continue;
    }
    parsedByTag.set(tag, version);
    newest = newest ? preferVersion(version, newest) : version;
  }

  const byCommit = new Map<string, Version>();
  let newestCommit: string | null = null;
  if (needCommits && parsedByTag.size > 0) {
    const refs = await listTagRefs(repo);
    for (const ref of refs) {
      const tag = ref.ref.replace(/^refs\/tags\//, "");
      const version = parsedByTag.get(tag);
      if (!version) {
        continue;
      }
      const commit =
        ref.object.type === "tag"
          ? await commitForAnnotatedTag(repo, ref.object.sha)
          : ref.object.sha.toLowerCase();
      const existing = byCommit.get(commit);
      byCommit.set(commit, existing ? preferVersion(version, existing) : version);
      if (newest && version.tag === newest.tag) {
        newestCommit = commit;
      }
    }
  }

  if (needCommits && newest && !newestCommit) {
    throw new Error(
      `Could not resolve the commit for actions/${repo} release ${newest.tag}.`
    );
  }

  return { newest, newestCommit, byCommit };
}

async function workflowFiles(dir: string): Promise<string[]> {
  const found: string[] = [];
  const entries = await fs.readdir(dir, { withFileTypes: true });
  for (const entry of entries) {
    if (entry.name === ".git" || entry.name === "node_modules") {
      continue;
    }
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      found.push(...(await workflowFiles(full)));
    } else if (entry.isFile() && (entry.name.endsWith(".yml") || entry.name.endsWith(".yaml"))) {
      found.push(full);
    }
  }
  return found;
}

async function findReferences(): Promise<Reference[]> {
  const references: Reference[] = [];
  for (const file of await workflowFiles(REPO_ROOT)) {
    const text = await fs.readFile(file, "utf8");
    const relative = path.relative(REPO_ROOT, file);
    const lines = text.split(/\r?\n/);
    for (let index = 0; index < lines.length; index++) {
      const line = lines[index];
      if (line.trim().startsWith("#")) {
        continue;
      }
      const match = USES_LINE.exec(line);
      if (!match) {
        continue;
      }
      const repo = match[1];
      const subpath = match[2] || "";
      const ref = match[3];
      const comment = (match[4] || "").trim();
      references.push({
        file: relative,
        line: index + 1,
        repo,
        uses: `actions/${repo}${subpath}@${ref}`,
        ref,
        comment,
      });
    }
  }
  return references;
}

function finding(reference: Reference, message: string): Finding {
  return {
    file: reference.file,
    line: reference.line,
    message: `${reference.file}:${reference.line}: ${message}`,
  };
}

function behind(uses: string, newest: Version): string {
  return `${uses} is behind ${newest.tag}.`;
}

function judge(reference: Reference, catalog: ReleaseCatalog): Finding | null {
  if (!catalog.newest) {
    return finding(
      reference,
      `${reference.uses} has no stable release with a v<major>, v<major>.<minor>, or v<major>.<minor>.<patch> tag.`
    );
  }

  const newest = catalog.newest;

  if (SHA.test(reference.ref)) {
    const commented = reference.comment ? versionFromComment(reference.comment) : null;
    if (commented) {
      if (isOutdated(commented, newest)) {
        return finding(reference, behind(`${reference.uses} (${commented.tag})`, newest));
      }
      return null;
    }
    // Do not call isOutdated here. A release tag of v5 or v5.2 would apply
    // moving-tag rules and hide a newer release on that same line. The commit
    // itself is current only when it is the highest stable release.
    if (catalog.newestCommit && reference.ref.toLowerCase() === catalog.newestCommit) {
      return null;
    }
    const resolved = catalog.byCommit.get(reference.ref.toLowerCase());
    if (resolved) {
      return finding(reference, behind(`${reference.uses} (${resolved.tag})`, newest));
    }
    return finding(
      reference,
      `${reference.uses} does not match the highest stable release. ${newest.tag} is available.`
    );
  }

  const pinned = parseVersion(reference.ref);
  if (!pinned) {
    return finding(reference, `${reference.uses} is not a version tag. ${newest.tag} is available.`);
  }
  if (isOutdated(pinned, newest)) {
    return finding(reference, behind(reference.uses, newest));
  }
  return null;
}

async function main(): Promise<number> {
  if (typeof fetch !== "function") {
    throw new Error("This script requires Node.js 20 or newer, which provides global fetch. CI uses Node.js 20.");
  }

  const references = await findReferences();
  const repos = Array.from(new Set(references.map((reference) => reference.repo))).sort();
  const needsCommits = new Set(
    references
      .filter((reference) => SHA.test(reference.ref) && !versionFromComment(reference.comment))
      .map((reference) => reference.repo)
  );

  const catalogs = new Map<string, ReleaseCatalog>();
  for (const repo of repos) {
    catalogs.set(repo, await catalogFor(repo, needsCommits.has(repo)));
  }

  const findings: Finding[] = [];
  for (const reference of references) {
    const catalog = catalogs.get(reference.repo);
    if (!catalog) {
      continue;
    }
    const result = judge(reference, catalog);
    if (result) {
      findings.push(result);
    }
  }

  findings.sort((left, right) => left.file.localeCompare(right.file) || left.line - right.line);

  const summary = `Checked ${references.length} references to ${repos.length} actions in the actions organization. ${findings.length} are behind a published release.`;
  if (findings.length === 0) {
    console.log(summary);
    return 0;
  }

  for (const item of findings) {
    console.log(item.message);
    if (process.env.GITHUB_ACTIONS) {
      console.log(`::error file=${item.file},line=${item.line}::${item.message}`);
    }
  }
  console.log(summary);
  return 1;
}

if (require.main === module) {
  main()
    .then((code) => {
      process.exitCode = code;
    })
    .catch((error: unknown) => {
      console.error(error instanceof Error ? error.message : error);
      process.exitCode = 1;
    });
}
