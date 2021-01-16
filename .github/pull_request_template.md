This repository contains configuration for what users see when they click on the `Actions` tab and the setup page for Code Scanning.

It is not:
* A playground to try out scripts
* A place for you to create a workflow for your repository

---

**Please note that at this time we are only accepting new starter workflows for Code Scanning. Updates to existing starter workflows are fine.**

---

In the workflow and properties files:

- [ ] The workflow filename of CI workflows should be the name of the language or platform, in lower case.  Special characters should be removed or replaced with words as appropriate (for example, "dotnet" instead of ".NET").

  The workflow filename of publishing workflows should be the name of the language or platform, in lower case, followed by "-publish".
- [ ] Includes a matching `ci/properties/*.properties.json` file.
- [ ] Use sentence case for the names of workflows and steps, for example "Run tests".
- [ ] The name of CI workflows should only be the name of the language or platform: for example "Go" (not "Go CI" or "Go Build")
- [ ] Include comments in the workflow for any parts that are not obvious or could use clarification.
- [ ] CI workflows should run on `push` to `branches: [ $default-branch ]` and `pull_request` to `branches: [ $default-branch ]`.
- [ ] Packaging workflows should run on `release` with `types: [ created ]`.
- [ ] Code Scanning workflows should run on `push` to `branches: [ $default-branch, $protected-branches ]` and `pull_request` to `branches: [ $default-branch ]`. We also recommend a `schedule` trigger of `cron: $cron-weekly`.

Some general notes:

- [ ] This workflow must only use actions that are produced by GitHub, [in the `actions` organization](https://github.com/actions), **or**
- [ ] This workflow must only use actions that are produced by the language or ecosystem that the workflow supports.  These actions must be [published to the GitHub Marketplace](https://github.com/marketplace?type=actions).  We recommend that these actions be referenced using the full 40 character hash of the action's commit instead of a tag.  Additionally, workflows must include the following comment at the top of the workflow file:
    ```
    # This workflow uses actions that are not certified by GitHub.
    # They are provided by a third-party and are governed by
    # separate terms of service, privacy policy, and support
    # documentation.
    ```
- [ ] Automation and CI workflows should not send data to any 3rd party service except for the purposes of installing dependencies.
- [ ] Automation and CI workflows cannot be dependent on a paid service or product.
