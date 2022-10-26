<!--
IMPORTANT:

This repository contains configuration for what users see when they click on the `Actions` tab and the setup page for Code Scanning.

It is not:
* A playground to try out scripts
* A place for you to create a workflow for your repository
-->

## Pre-requisites

- [ ] Prior to submitting a new workflow, please apply to join the GitHub Technology Partner Program: [partner.github.com/apply](https://partner.github.com/apply?partnershipType=Technology+Partner).

---

### **Please note that at this time we are only accepting new starter workflows for Code Scanning. Updates to existing starter workflows are fine.**

---

## Tasks

**For _all_ workflows, the workflow:**

- [ ] Should be contained in a `.yml` file with the language or platform as its filename, in lower, [_kebab-cased_](https://en.wikipedia.org/wiki/Kebab_case) format (for example, [`docker-image.yml`](https://github.com/actions/starter-workflows/blob/main/ci/docker-image.yml)).  Special characters should be removed or replaced with words as appropriate (for example, "dotnet" instead of ".NET").
- [ ] Should use sentence case for the names of workflows and steps (for example, "Run tests").
- [ ] Should be named _only_ by the name of the language or platform (for example, "Go", not "Go CI" or "Go Build").
- [ ] Should include comments in the workflow for any parts that are not obvious or could use clarification.
- [ ] Should specify least priviledge [permissions](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#modifying-the-permissions-for-the-github_token) for `GITHUB_TOKEN` so that the workflow runs successfully.

**For _CI_ workflows, the workflow:**

- [ ] Should be preserved under [the `ci` directory](https://github.com/actions/starter-workflows/tree/main/ci).
- [ ] Should include a matching `ci/properties/*.properties.json` file (for example, [`ci/properties/docker-publish.properties.json`](https://github.com/actions/starter-workflows/blob/main/ci/properties/docker-publish.properties.json)).
- [ ] Should run on `push` to `branches: [ $default-branch ]` and `pull_request` to `branches: [ $default-branch ]`.
- [ ] Packaging workflows should run on `release` with `types: [ created ]`.
- [ ] Publishing workflows should have a filename that is the name of the language or platform, in lower case, followed by "-publish" (for example, [`docker-publish.yml`](https://github.com/actions/starter-workflows/blob/main/ci/docker-publish.yml)).

**For _Code Scanning_ workflows, the workflow:**

- [ ] Should be preserved under [the `code-scanning` directory](https://github.com/actions/starter-workflows/tree/main/code-scanning).
- [ ] Should include a matching `code-scanning/properties/*.properties.json` file (for example, [`code-scanning/properties/codeql.properties.json`](https://github.com/actions/starter-workflows/blob/main/code-scanning/properties/codeql.properties.json)), with properties set as follows:
  - [ ] `name`: Name of the Code Scanning integration.
  - [ ] `creator`: Name of the organization/user producing the Code Scanning integration.
  - [ ] `description`: Short description of the Code Scanning integration.
  - [ ] `categories`: Array of languages supported by the Code Scanning integration.
  - [ ] `iconName`: Name of the SVG logo representing the Code Scanning integration. This SVG logo must be present in [the `icons` directory](https://github.com/actions/starter-workflows/tree/main/icons).
- [ ] Should run on `push` to `branches: [ $default-branch, $protected-branches ]` and `pull_request` to `branches: [ $default-branch ]`. We also recommend a `schedule` trigger of `cron: $cron-weekly` (for example, [`codeql.yml`](https://github.com/actions/starter-workflows/blob/c59b62dee0eae1f9f368b7011cf05c2fc42cf084/code-scanning/codeql.yml#L14-L21)).

**Some general notes:**

- [ ] This workflow must _only_ use actions that are produced by GitHub, [in the `actions` organization](https://github.com/actions), **or**
- [ ] This workflow must _only_ use actions that are produced by the language or ecosystem that the workflow supports.  These actions must be [published to the GitHub Marketplace](https://github.com/marketplace?type=actions).  We require that these actions be referenced using the full 40 character hash of the action's commit instead of a tag.  Additionally, workflows must include the following comment at the top of the workflow file:
    ```
    # This workflow uses actions that are not certified by GitHub.
    # They are provided by a third-party and are governed by
    # separate terms of service, privacy policy, and support
    # documentation.
    ```
- [ ] Automation and CI workflows should not send data to any 3rd party service except for the purposes of installing dependencies.
- [ ] Automation and CI workflows cannot be dependent on a paid service or product.
