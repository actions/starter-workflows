# Check for outdated actions

This script fails when a workflow in this repository depends on an older release of an action published in the [`actions` organization](https://github.com/actions). That organization is where GitHub publishes actions such as `actions/checkout` and `actions/setup-node`. This repository's [contributing guide](../../CONTRIBUTING.md) treats actions outside that organization as a separate case, pinned by commit SHA, so this check leaves them alone.

The script walks every `.yml` and `.yaml` file under the repository root, including `.github/workflows` and the starter-workflow directories. It skips `node_modules` and `.git`. A workflow added in a new directory is included without a folder list to update.

## How a version is judged

GitHub's release documentation for actions says a major tag such as `v1` moves to the latest related minor or patch, and a minor tag such as `v1.1` moves to the latest related patch. A release tag such as `v1.0.0` is the unchanging release. See [Using immutable releases and tags to manage your action's releases](https://docs.github.com/en/actions/how-tos/create-and-publish-actions/using-immutable-releases-and-tags-to-manage-your-actions-releases).

The check follows that scheme. It compares the version the workflow wrote down with the highest stable release, and it keeps the specificity of what was written:

| Reference | Current while | Error when |
| --- | --- | --- |
| `v5` | the highest stable release is still a `v5` release | a `v6` or later release exists |
| `v5.2` | the highest stable release is still `v5.2.y` | `v5.3` or a later major exists |
| `v5.2.1` | that release is the highest stable release | any higher stable release exists |

A stable release is a GitHub Release that is not a draft and not a prerelease, whose tag is `v<major>`, `v<major>.<minor>`, or `v<major>.<minor>.<patch>` with nothing after it. `v3.1.0-node20` is ignored. The highest release is the highest of those numbers.

The Releases API returns the release that was published most recently, which is not always the highest number. `actions/cache` published `v5.1.0` after `v6.1.0`. This check treats `v6.1.0` as the newer release. A workflow that says `actions/cache@v5` is an error because `v6.1.0` exists. A workflow that says `actions/cache@v6` is current.

## Commit SHAs

The script reads `uses:` from the workflow file text. A version comment beside a commit SHA is not part of the YAML value, and a YAML parser would drop it. The comment is the version:

```yaml
uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
uses: actions/checkout@692973e3d937129bcbf40652eb9f2f61becf3332 # actions/checkout@v4
```

`# v4.2.2` is a patch pin. `# v4` and `# actions/checkout@v4` are major pins. The same table above applies.

A commit SHA with no version comment is current when that commit is the highest stable release. Any other commit is an error, including a commit that matches an older release and a commit that matches no stable release.

## Run it

Use Node.js 20 or newer. The script calls the global `fetch` added in Node.js 18, and CI runs Node.js 20. From this directory:

```bash
npm ci
npx ts-node ./index.ts
```

The script reads `GITHUB_TOKEN`, then `GH_TOKEN`. In GitHub Actions the workflow sets `GITHUB_TOKEN` to `github.token` on the checker step only. `npm ci` does not receive it, because a pull request can change the lifecycle scripts in `package.json` and npm runs those scripts during install. Locally, `gh auth token` prints a token that can read public repositories. Unauthenticated requests are limited to 60 per hour. One run asks GitHub once per action repository (about 20), plus another request for each repository that is pinned by a commit SHA with no version comment, plus one request per annotated tag that has to be peeled to a commit. A second unauthenticated run in the same hour can be rejected.

Each outdated reference is one line:

```text
ci/go.yml:22: actions/checkout@v4 is behind v7.0.1.
```

The process exits 0 when every reference is current. It exits 1 when any reference is behind, or when the GitHub API request fails. An API failure is an error rather than a pass, because a skipped repository would hide a stale pin.

When `GITHUB_ACTIONS` is set, the same line is also emitted as a workflow command:

```text
::error file=ci/go.yml,line=22::ci/go.yml:22: actions/checkout@v4 is behind v7.0.1.
```

## Continuous integration

[`.github/workflows/check-outdated-actions.yaml`](../../.github/workflows/check-outdated-actions.yaml) runs `npm ci` and `npx ts-node ./index.ts` in this directory on every push and pull request. The job permission is `contents: read`. The token is used only to read public release data for `actions/*` repositories.
