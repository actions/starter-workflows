# starter-workflows
Accelerating new GitHub Actions workflows 
# Create-ActionsPRs
The script in this repository creates pull requests to push a GitHub Actions workflow to multiple repositories. 

### Prerequisites
* Install [PowerShell](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell?view=powershell-7)
* Install and configure [git](https://git-scm.com/), and configure your account for access to GitHub. [Configuring SSH Keys](https://help.github.com/en/github/authenticating-to-github/connecting-to-github-with-ssh)
* Install [GitHub CLI](https://cli.github.com/)

### Instructions

#### Option 1: Use GitHub API to automatically target everything
1. Clone this repository to your local machine
2. Open powershell and navigate to the directory you cloned this into 
3. Rename the file called `.env-example` to `.env` and add a [GitHub token](https://github.com/settings/tokens) with the `repo` scope. If your organization requires SSO, authorize the token for SSO. 
4. Confirm that the `workflows` directory has all of the workflow files you want in it
5. Run `./Create-ActionsPRs.ps1` to install the script. 
6. Run `CreatePullRequestForRepositories -Repositories (GetReposFromOrganization -Organization orgname) -CommitMessage "message" -PRBody "prbody" -BranchName "branch_name"` 

This will create PRs in every repository that you have push permisisons in. 

#### Option 2: Specify a list to target 
1. Clone this repository to your local machine
2. Open powershell and navigate to the directory you cloned this into 
3. Confirm that you have all the workflow files you want in the `workflows` directory
4. Run `./Create-ActionsPRs.ps1` to install the script. 
5. Create a file with a list of repository URLs, with one repository per line. Save it somewhere you can access. 
6. Confirm that the `workflows` directory has all of the workflow files you want in it
7. Run `CreatePullRequestsFromFile -FileName file.txt -CommitMessage "message" -PRBody "prbody" -BranchName "branch_name"`. Replace file.txt with the path to your file, and update the CommitMessage and PRBody as appropriate. 

#### Option 3: Choose all CodeQL eligible repositories
1. Clone this repository to your local machine
2. Open powershell and navigate to the directory you cloned this into 
3. Rename the file called `.env-example` to `.env` and add a [GitHub token](https://github.com/settings/tokens) with the `repo` scope. If your organization requires SSO, authorize the token for SSO. 
4. Confirm that the `workflows` directory has all of the workflow files you want in it; the CodeQL workflow file is already included in this repo
5. Run `CreatePullRequestsForCodeQLLanguages -Organization orgname` with the organization you want to target instead of orgname. You may optionally override the commit message or PR body by adding `-CommitMessage` or `-PRBody` as appropriate.

<p align="center">
  <img src="https://avatars0.githubusercontent.com/u/44036562?s=100&v=4"/> 
</p>

## Starter Workflows

These are the workflow files for helping people get started with GitHub Actions.  They're presented whenever you start to create a new GitHub Actions workflow.

**If you want to get started with GitHub Actions, you can use these starter workflows by clicking the "Actions" tab in the repository where you want to create a workflow.**

<img src="https://d3vv6lp55qjaqc.cloudfront.net/items/353A3p3Y2x3c2t2N0c01/Image%202019-08-27%20at%203.25.07%20PM.png" max-width="75%"/>

### Directory structure

* [ci](ci): solutions for Continuous Integration workflows
* [deployments](deployments): solutions for Deployment workflows
* [automation](automation): solutions for automating workflows
* [code-scanning](code-scanning): solutions for [Code Scanning](https://github.com/features/security)
* [pages](pages): solutions for Pages workflows
* [icons](icons): svg icons for the relevant template

Each workflow must be written in YAML and have a `.yml` extension. They also need a corresponding `.properties.json` file that contains extra metadata about the workflow (this is displayed in the GitHub.com UI).

For example: `ci/django.yml` and `ci/properties/django.properties.json`.

### Valid properties

* `name`: the name shown in onboarding. This property is unique within the repository.
* `description`: the description shown in onboarding
* `iconName`: the icon name in the relevant folder, for example, `django` should have an icon `icons/django.svg`. Only SVG is supported at this time. Another option is to use [octicon](https://primer.style/octicons/). The format to use an octicon is `octicon <<icon name>>`. Example: `octicon person`
* `creator`: creator of the template shown in onboarding. All the workflow templates from an author will have the same `creator` field.
* `categories`: the categories that it will be shown under. Choose at least one category from the list [here](#categories). Further, choose the categories from the list of languages available [here](https://github.com/github/linguist/blob/master/lib/linguist/languages.yml) and the list of tech stacks available [here](https://github.com/github-starter-workflows/repo-analysis-partner/blob/main/tech_stacks.yml). When a user views the available templates, those templates that match the language and tech stacks will feature more prominently.

### Categories
* continuous-integration
* deployment
* testing
* code-quality
* code-review
* dependency-management
* monitoring
* Automation
* utilities
* Pages
* Hugo

### Variables
These variables can be placed in the starter workflow and will be substituted as detailed below:

* `$default-branch`: will substitute the branch from the repository, for example `main` and `master`
* `$protected-branches`: will substitute any protected branches from the repository
* `$cron-daily`: will substitute a valid but random time within the day

## How to test templates before publishing

### Disable template for public
The template author adds a `labels` array in the template's `properties.json` file with a label `preview`. This will hide the template from users, unless user uses query parameter `preview=true` in the URL.
Example `properties.json` file:
```json
{
    "name": "Node.js",
    "description": "Build and test a Node.js project with npm.",
    "iconName": "nodejs",
    "categories": ["Continuous integration", "JavaScript", "npm", "React", "Angular", "Vue"],
    "labels": ["preview"]
}
```

For viewing the templates with `preview` label, provide query parameter `preview=true` to the  `new workflow` page URL. Eg. `https://github.com/<owner>/<repo_name>/actions/new?preview=true`.

### Enable template for public
Remove the `labels` array from `properties.json` file to publish the template to public

.----

# GitHub Advanced Security - Code Scanning, Secret Scanning & Dependabot Bulk Enablement Tooling

## Purpose

The purpose of this tool is to help enable GitHub Advanced Security (GHAS) across multiple repositories in an automated way. There will be times when you need the ability to enable Code Scanning (CodeQL), Secret Scanning, Dependabot Alerts, and/or Dependabot Security Updates across various repositories, and you don't want to click buttons manually or drop a GitHub Workflow for CodeQL into every repository. Doing this is manual and painstaking. The purpose of this utility is to help automate these manual tasks.

## Context

The primary motivator for this utility is CodeQL. It is incredibly time-consuming to enable CodeQL across multiple repositories. Additionally, no API allows write access to the `.github/workflow/` directory. So this means teams have to write various scripts with varying results. This tool provides a tried and proven way of doing that.

Secret Scanning & Dependabot is also hard to enable if you only want to enable it on specific repositories versus everything. This tool allows you to do that easily.

## What does this tooling do?

There are two main actions this tool does:

**Part One:**

Goes and collects repositories that will have Code Scanning (CodeQL)/Secret Scanning/Dependabot Alerts/Dependabot Security Updates enabled. There are three main ways these repositories are collected.

- Collect the repositories where the primary language matches a specific value. For example, if you provide JavaScript, all repositories will be collected where the primary language is, Javascript.
- Collect the repositories to which a user has administrative access, or a GitHub App has access.

If you select option 1, the script will return all repositories in the language you specify (which you have access to). The repositories collected from this script are then stored within a `repos.json` file. If you specify option 2, the script will return all repositories you are an administrator over. The third option is to define the `repos.json` manually. We don't recommend this, but it's possible. If you want to go down this path, first run one of the above options for collecting repository information automatically, look at the structure, and build your fine of the laid out format.

**Part Two:**

Loops over the repositories found within the `repos.json` file and enables Code Scanning(CodeQL)/Secret Scanning/Dependabot Alerts/Dependabot Security Updates/Secret Scanning Push Protection.

If you pick Code Scanning:

- Loops over the repositories found within the `repos.json` file. A pull request gets created on that repository with the `codeql-analysis-${language}.yml` found in the `bin/workflows` directory. The `${language}` will be replaced at runtime with the primary language of the repository. For convenience, all pull requests made will be stored within the `prs.txt` file, where you can see and manually review the pull requests after the script has run.

If you pick Secret Scanning:

- Loops over the repositories found within the `repos.json` file. Secret Scanning is then enabled on these repositories.

If you pick Dependabot Alerts:

- Loops over the repositories found within the `repos.json` file. Dependabot Alerts is then enabled on these repositories.

If you pick Dependabot Security Updates:

- Loops over the repositories found within the `repos.json` file. Dependabot Security Updates is then enabled on these repositories.

## Prerequisites

- [Node v18](https://nodejs.org/en/download/) or higher installed.
- [Yarn](https://yarnpkg.com/)\*
- [TypeScript](https://www.typescriptlang.org/download)
- [Git](https://git-scm.com/downloads) installed on the (user's) machine running this tool.
- A Personal Access Token (PAT) that has at least admin access over the repositories they want to enable Code Scanning on or GitHub App credentials which have access to the repositories you want to enable Code Scanning on.
- Some basic software development skills, e.g., can navigate their way around a terminal or command prompt.

* You can use `npm` but for the sake of this `README.md`; we are going to standardise the commands on yarn. These are easily replaceable though with `npm` commands.

## Set up Instructions

1.  Clone this repository onto your local machine.

    ```bash
    git clone https://github.com/NickLiffen/ghas-enablement.git
    ```

2.  Change the directory to the repository you have just installed.

    ```bash
    cd ghas-enablement
    ```

3.  Generate your chosen authentication strategy. You are either able to use a [GitHub App](https://docs.github.com/en/developers/apps/getting-started-with-apps/about-apps) or a [Personal Access Token (PAT)](https://github.com/settings/tokens/new). The GitHub App needs to have permissions of `read and write` of `administration`, `Code scanning alerts`, `contents`, `issues`, `pull requests`, `workflows`. The GitHub PAT needs access to `repo`, `workflow` and `read:org` only. (if you are running `yarn run getOrgs` you will also need the `read:enterprise` scope).

4.  Copy the `.env.sample` to `.env`. On a Mac, this can be done via the following terminal command:

    ```bash
    cp .env.sample .env
    ```

5.  Update the `.env` with the required values. Please pick one of the authentication methods for interacting with GitHub. You can either fill in the `GITHUB_API_TOKEN` with a PAT that has access to the Org. OR, fill in all the values required for a GitHub App. **Note**: It is recommended to pick the GitHub App choice if running on thousands of repositories, as this gives you more API requests versus a PAT.

    - If using a GitHub App, either paste in the value as-is in the `APP_PRIVATE_KEY` in the field surrounded by double quotes (the key will take up multiple lines), or convert the private key to a single line surrounded in double quotes by replacing the new line character with `\n` (In VS Code on Mac, you can use `⌃ + Enter` to find/replace the new line character)

6.  Update the `GITHUB_ORG` value found within the `.env`. Remove the `XXXX` and replace that with the name of the GitHub Organisation you would like to use as part of this script. **NOTE**: If you are running this across multiple organisations within an enterprise, you can not set the `GITHUB_ORG` variable and instead set the `GITHUB_ENTERPRISE` one with the name of the enterprise. You can then run `yarn run getOrgs`, which will collect all the organisations dynamically. This will mean you don't have to hardcode one. However, for most use cases, simply hardcoding the specific org within the `GITHUB_ORG` variable where you would like this script to run will work.

7.  Update the `LANGUAGE_TO_CHECK` value found within the `.env`. Remove the `XXXX` and replace that with the language you would like to use as a filter when collecting repositories. **Note**: Please make sure these are lowercase values, such as: `javascript`, `python`, `go`, `ruby`, etc.

8.  Decide what you want to enable. Update the `ENABLE_ON` value to choose what you want to enable on the repositories found within the `repos.json`. This can be one or multiple values. If you are enabling just code scanning (CodeQL) you will need to set `ENABLE_ON=codescanning`, if you are enabling everything, you will need to set `ENABLE_ON=codescanning,secretscanning,pushprotection,dependabot,dependabotupdates`. You can pick one, two or three. The format is a comma-seperated list.

9.  **OPTIONAL**: Update the `CREATE_ISSUE` value to `true/false` depending on if you would like to create an issue explaining the purpose of the PR. We recommend this, as it will help explain why the PR was created; and give some context. However, this is optional. The text which is in the issue can be modified and found here: `./src/utils/text/`.

10. **OPTIONAL**: If you are a GHES customer, then you will need to set the `GHES` env to `true` and then set `GHES_SERVER_BASE_URL` to the URL of your GHES instance. E.G `https://octodemo.com`.

11. If you are enabling Code Scanning (CodeQL), check the `codeql-analysis.yml` file. This is a sample file; please configure this file to suit your repositories needs.

12. Run `yarn install` or `npm install`, which will install the necessary dependencies.

13. Run `yarn run build` or `npm run build`, which will create the JavaScript bundle from TypeScript.

## How to use?

There are two simple steps to run:

### Step One

The first step is collecting the repositories you would like to run this script on. You have three options as mentioned above. Option 1 is automated and finds all the repositories within an organisation you have admin access to. Option 2 is automated and finds all the repositories within an organisation based on the language you specify. Or, Option 3, which is a manual entry of the repositories you would like to run this script on. See more information below.

**OPTION 1** (Preferred)

```bash
yarn run getRepos // In the `.env` set the `LANGUAGE_TO_CHECK=` to the language. E.G `python`, `javascript`, `go`, etc.
```

When using GitHub Actions, we commonly find (especially for non-build languages such as JavaScript) that the `codeql-analysis.yml` file is repeatable and consistent across multiple repositories of the same language. About 80% of the time, teams can reuse the same workflow files for the same language. For Java, C++ that number drops down to about 60% of the time. But the reason why we recommend enabling Code Scanning at bulk via language is the `codeql-analysis.yml` file you propose within the pull request has the highest chance of being most accurate. Even if the file needs changing, the team reviewing the pull request would likely only need to make small changes. We recommend you run this command first to get a list of repositories to enable Code Scanning. After running the command, you are welcome to modify this file. Just make sure it's a valid JSON file if you do edit.

This script only returns repositories where CodeQL results have not already been uploaded to code scanning. If any CodeQL results have been uploaded to a repositories code scanning feature, that repository will not be returned to this list. The motivation behind this is not to raise pull requests on repositories where CodeQL has already been enabled.

**OPTION 2**

```bash
yarn run getRepos // or npm run getRepos
```

Similar to step one, another automated approach is to enable by user access. This approach will be a little less accurate as the file will most certainly need changing between a Python project and a Java project (if you are enabling CodeQL), and the user's PAT you are using will most likely. But the file you propose is going to be a good start. After running the command, you are welcome to modify this file. Just make sure it's a valid JSON file if you do edit.

This script only returns repositories where CodeQL results have not already been uploaded to code scanning. If any CodeQL results have been uploaded to a repositories code scanning feature, that repository will not be returned to this list. The motivation behind this is not to raise pull requests on repositories where CodeQL has already been enabled.

**OPTION 3**

Create a file called `repos.json` within the `./bin/` directory. This file needs to have an array of organization objects, each with its own array of repository objects. The structure of the objects should look like this:

```JSON
[
  {
    "login": "string <org>",
    "repos":
    [
      {
        "createIssue": "boolean",
        "enableCodeScanning": "boolean",
        "enableDependabot": "boolean",
        "enableDependabotUpdates": "boolean",
        "enablePushProtection": "boolean",
        "enableSecretScanning": "boolean",
        "repo": "string <org/repo>",
      }
    ]
  }
]
```

As you can see, the object takes a number of boolean keys: `createIssue`, `enableCodeScanning`, `enableDependabot`, `enableDependabotUpdates`, `enablePushProtection`, and `enableSecretScanning`, along with a single string key, namely, `repo`. Set `repo` to the name of the repository name where you would like to run this script. Set `enableDependabot` to `true` if you would also like to enable Dependabot Alerts on that repo; set it to `false` if you do not want to enable Dependabot Alerts. The same goes for `enableDependabotUpdates` for Dependabot Security Updates, `enableSecretScanning` for Secret Scanning, `pushprotection` for Secret Scanning push protection, and `enableCodeScanning` for Code Scanning (CodeQL). Finally set `createIssue` to `true` if you would like to create an issue on the repository with the text found in the `./src/utils/text/issueText.ts` file to supplement the PR.

**NOTE:** The account that generated the PAT needs to have `write` access or higher over any repository that you include within the `repos` key.

### Step Two

Run the script which enables Code Scanning (and/or Dependabot Alerts/Dependabot Security Updates/Secret Scanning) on your repository by running:

```bash
yarn run start // or npm run start
```

This will run a script, and you should see output text appearing on your screen.

After the script has run, please head to your `~/Desktop` directory and delete the `tempGitLocations` directory that has been automatically created.

## Running this within a Codespace?

There are some key considerations that you will need to put into place if you are running this script within a GitHub Codespace:

1. You will need to add the following snippet to the `.devcontainer/devcontainer.json`:

```json
  "codespaces": {
    "repositories": [
      {
        "name": "<ORG_NAME>/*",
        "permissions": "write-all"
      }
    ]
  }
```

The reason you need this within your `.devcontainer/devcontainer.json` file is the `GITHUB_TOKEN` tied to the Codespace will need to access other repositories within your organisation which this script may interact with. You will need to create a new Codespace **after** you have added the above and pushed it to your repository.

You do not need to do the above if you are not running it from a Codespace.

## Running as a (scheduled) GitHub workflow

Since this tool uses a PAT or GitHub App Authentication wherever authentication is required, it can be run unattended. You can see in the example
below how you could run the tool in a scheduled GitHub workflow. Instead of using the `.env`
file you can configure all the variables from the `.env.sample` directly as environment variables. This will allow you to
(easily) make use of GitHub action secrets for the PAT or GitHub App credentials.

```yaml
on:
  schedule:
    - cron: "5 16 * * 1"

env:
  APP_ID: ${{ secrets.GHAS_ENABLEMENT_APP_ID }}
  APP_CLIENT_ID: ${{ secrets.GHAS_ENABLEMENT_APP_CLIENT_ID }}
  APP_CLIENT_SECRET: ${{ secrets.GHAS_ENABLEMENT_APP_CLIENT_SECRET }}
  APP_PRIVATE_KEY: ${{ secrets.GHAS_ENABLEMENT_APP_PRIVATE_KEY }}
  ENABLE_ON: "codescanning,secretscanning,pushprotection,dependabot,dependabotupdates"
  DEBUG: "ghas:*"
  CREATE_ISSUE: "false"
  GHES: "false"
  # Organization specific variables
  APP_INSTALLATION_ID: "12345678"
  GITHUB_ORG: "my-target-org"

jobs:
  enable-security-javascript:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
        with:
          repository: NickLiffen/ghas-enablement
      - name: Get dependencies and configure
        run: |
          yarn
          git config --global user.name "ghas-enablement"
          git config --global user.email "ghas.enablement@example.com"
      - name: Enable security on organization (javascript)
        run: |
          npm run getRepos
          npm run start
        env:
          LANGUAGE_TO_CHECK: "javascript"
          TEMP_DIR: ${{ github.workspace }}
```

You can duplicate the last step for the other languages commonly used within your enterprise/organisation.
If you didn't configure the tool as a GitHub App, you can remove all the `APP_*` and set `GITHUB_API_TOKEN` instead.
Above we rely on the sample codeql file for javascript included in this repository. Alternatively you could add this workflow to a repository
containing your customized codeql files and use those to overwrite the samples.

## Found an Issue?

Create an issue within the repository and make it to `@nickliffen`. Key things to mention within your issue:

- Windows, Linux, Codespaces or Mac
- What version of NodeJS you are running.
- Add any logs that appeared when you ran into the issue.

## Want to Contribute?

Great! Open an issue, describe what feature you want to create and make sure to `@nickliffen`.

---
# GitHub Advanced Security - Code Scanning, Secret Scanning & Dependabot Bulk Enablement Tooling

## Purpose

The purpose of this tool is to help enable GitHub Advanced Security (GHAS) across multiple repositories in an automated way. There will be times when you need the ability to enable Code Scanning (CodeQL), Secret Scanning, Dependabot Alerts, and/or Dependabot Security Updates across various repositories, and you don't want to click buttons manually or drop a GitHub Workflow for CodeQL into every repository. Doing this is manual and painstaking. The purpose of this utility is to help automate these manual tasks.

## Context

The primary motivator for this utility is CodeQL. It is incredibly time-consuming to enable CodeQL across multiple repositories. Additionally, no API allows write access to the `.github/workflow/` directory. So this means teams have to write various scripts with varying results. This tool provides a tried and proven way of doing that.

Secret Scanning & Dependabot is also hard to enable if you only want to enable it on specific repositories versus everything. This tool allows you to do that easily.

## What does this tooling do?

There are two main actions this tool does:

**Part One:**

Goes and collects repositories that will have Code Scanning (CodeQL)/Secret Scanning/Dependabot Alerts/Dependabot Security Updates enabled. There are three main ways these repositories are collected.

- Collect the repositories where the primary language matches a specific value. For example, if you provide JavaScript, all repositories will be collected where the primary language is, Javascript.
- Collect the repositories to which a user has administrative access, or a GitHub App has access.

If you select option 1, the script will return all repositories in the language you specify (which you have access to). The repositories collected from this script are then stored within a `repos.json` file. If you specify option 2, the script will return all repositories you are an administrator over. The third option is to define the `repos.json` manually. We don't recommend this, but it's possible. If you want to go down this path, first run one of the above options for collecting repository information automatically, look at the structure, and build your fine of the laid out format.

**Part Two:**

Loops over the repositories found within the `repos.json` file and enables Code Scanning(CodeQL)/Secret Scanning/Dependabot Alerts/Dependabot Security Updates/Secret Scanning Push Protection.

If you pick Code Scanning:

- Loops over the repositories found within the `repos.json` file. A pull request gets created on that repository with the `codeql-analysis-${language}.yml` found in the `bin/workflows` directory. The `${language}` will be replaced at runtime with the primary language of the repository. For convenience, all pull requests made will be stored within the `prs.txt` file, where you can see and manually review the pull requests after the script has run.

If you pick Secret Scanning:

- Loops over the repositories found within the `repos.json` file. Secret Scanning is then enabled on these repositories.

If you pick Dependabot Alerts:

- Loops over the repositories found within the `repos.json` file. Dependabot Alerts is then enabled on these repositories.

If you pick Dependabot Security Updates:

- Loops over the repositories found within the `repos.json` file. Dependabot Security Updates is then enabled on these repositories.

## Prerequisites

- [Node v18](https://nodejs.org/en/download/) or higher installed.
- [Yarn](https://yarnpkg.com/)\*
- [TypeScript](https://www.typescriptlang.org/download)
- [Git](https://git-scm.com/downloads) installed on the (user's) machine running this tool.
- A Personal Access Token (PAT) that has at least admin access over the repositories they want to enable Code Scanning on or GitHub App credentials which have access to the repositories you want to enable Code Scanning on.
- Some basic software development skills, e.g., can navigate their way around a terminal or command prompt.

* You can use `npm` but for the sake of this `README.md`; we are going to standardise the commands on yarn. These are easily replaceable though with `npm` commands.

## Set up Instructions

1.  Clone this repository onto your local machine.

    ```bash
    git clone https://github.com/NickLiffen/ghas-enablement.git
    ```

2.  Change the directory to the repository you have just installed.

    ```bash
    cd ghas-enablement
    ```

3.  Generate your chosen authentication strategy. You are either able to use a [GitHub App](https://docs.github.com/en/developers/apps/getting-started-with-apps/about-apps) or a [Personal Access Token (PAT)](https://github.com/settings/tokens/new). The GitHub App needs to have permissions of `read and write` of `administration`, `Code scanning alerts`, `contents`, `issues`, `pull requests`, `workflows`. The GitHub PAT needs access to `repo`, `workflow` and `read:org` only. (if you are running `yarn run getOrgs` you will also need the `read:enterprise` scope).

4.  Copy the `.env.sample` to `.env`. On a Mac, this can be done via the following terminal command:

    ```bash
    cp .env.sample .env
    ```

5.  Update the `.env` with the required values. Please pick one of the authentication methods for interacting with GitHub. You can either fill in the `GITHUB_API_TOKEN` with a PAT that has access to the Org. OR, fill in all the values required for a GitHub App. **Note**: It is recommended to pick the GitHub App choice if running on thousands of repositories, as this gives you more API requests versus a PAT.

    - If using a GitHub App, either paste in the value as-is in the `APP_PRIVATE_KEY` in the field surrounded by double quotes (the key will take up multiple lines), or convert the private key to a single line surrounded in double quotes by replacing the new line character with `\n` (In VS Code on Mac, you can use `⌃ + Enter` to find/replace the new line character)

6.  Update the `GITHUB_ORG` value found within the `.env`. Remove the `XXXX` and replace that with the name of the GitHub Organisation you would like to use as part of this script. **NOTE**: If you are running this across multiple organisations within an enterprise, you can not set the `GITHUB_ORG` variable and instead set the `GITHUB_ENTERPRISE` one with the name of the enterprise. You can then run `yarn run getOrgs`, which will collect all the organisations dynamically. This will mean you don't have to hardcode one. However, for most use cases, simply hardcoding the specific org within the `GITHUB_ORG` variable where you would like this script to run will work.

7.  Update the `LANGUAGE_TO_CHECK` value found within the `.env`. Remove the `XXXX` and replace that with the language you would like to use as a filter when collecting repositories. **Note**: Please make sure these are lowercase values, such as: `javascript`, `python`, `go`, `ruby`, etc.

8.  Decide what you want to enable. Update the `ENABLE_ON` value to choose what you want to enable on the repositories found within the `repos.json`. This can be one or multiple values. If you are enabling just code scanning (CodeQL) you will need to set `ENABLE_ON=codescanning`, if you are enabling everything, you will need to set `ENABLE_ON=codescanning,secretscanning,pushprotection,dependabot,dependabotupdates`. You can pick one, two or three. The format is a comma-seperated list.

9.  **OPTIONAL**: Update the `CREATE_ISSUE` value to `true/false` depending on if you would like to create an issue explaining the purpose of the PR. We recommend this, as it will help explain why the PR was created; and give some context. However, this is optional. The text which is in the issue can be modified and found here: `./src/utils/text/`.

10. **OPTIONAL**: If you are a GHES customer, then you will need to set the `GHES` env to `true` and then set `GHES_SERVER_BASE_URL` to the URL of your GHES instance. E.G `https://octodemo.com`.

11. If you are enabling Code Scanning (CodeQL), check the `codeql-analysis.yml` file. This is a sample file; please configure this file to suit your repositories needs.

12. Run `yarn install` or `npm install`, which will install the necessary dependencies.

13. Run `yarn run build` or `npm run build`, which will create the JavaScript bundle from TypeScript.

## How to use?

There are two simple steps to run:

### Step One

The first step is collecting the repositories you would like to run this script on. You have three options as mentioned above. Option 1 is automated and finds all the repositories within an organisation you have admin access to. Option 2 is automated and finds all the repositories within an organisation based on the language you specify. Or, Option 3, which is a manual entry of the repositories you would like to run this script on. See more information below.

**OPTION 1** (Preferred)

```bash
yarn run getRepos // In the `.env` set the `LANGUAGE_TO_CHECK=` to the language. E.G `python`, `javascript`, `go`, etc.
```

When using GitHub Actions, we commonly find (especially for non-build languages such as JavaScript) that the `codeql-analysis.yml` file is repeatable and consistent across multiple repositories of the same language. About 80% of the time, teams can reuse the same workflow files for the same language. For Java, C++ that number drops down to about 60% of the time. But the reason why we recommend enabling Code Scanning at bulk via language is the `codeql-analysis.yml` file you propose within the pull request has the highest chance of being most accurate. Even if the file needs changing, the team reviewing the pull request would likely only need to make small changes. We recommend you run this command first to get a list of repositories to enable Code Scanning. After running the command, you are welcome to modify this file. Just make sure it's a valid JSON file if you do edit.

This script only returns repositories where CodeQL results have not already been uploaded to code scanning. If any CodeQL results have been uploaded to a repositories code scanning feature, that repository will not be returned to this list. The motivation behind this is not to raise pull requests on repositories where CodeQL has already been enabled.

**OPTION 2**

```bash
yarn run getRepos // or npm run getRepos
```

Similar to step one, another automated approach is to enable by user access. This approach will be a little less accurate as the file will most certainly need changing between a Python project and a Java project (if you are enabling CodeQL), and the user's PAT you are using will most likely. But the file you propose is going to be a good start. After running the command, you are welcome to modify this file. Just make sure it's a valid JSON file if you do edit.

This script only returns repositories where CodeQL results have not already been uploaded to code scanning. If any CodeQL results have been uploaded to a repositories code scanning feature, that repository will not be returned to this list. The motivation behind this is not to raise pull requests on repositories where CodeQL has already been enabled.

**OPTION 3**

Create a file called `repos.json` within the `./bin/` directory. This file needs to have an array of organization objects, each with its own array of repository objects. The structure of the objects should look like this:

```JSON
[
  {
    "login": "string <org>",
    "repos":
    [
      {
        "createIssue": "boolean",
        "enableCodeScanning": "boolean",
        "enableDependabot": "boolean",
        "enableDependabotUpdates": "boolean",
        "enablePushProtection": "boolean",
        "enableSecretScanning": "boolean",
        "repo": "string <org/repo>",
      }
    ]
  }
]
```

As you can see, the object takes a number of boolean keys: `createIssue`, `enableCodeScanning`, `enableDependabot`, `enableDependabotUpdates`, `enablePushProtection`, and `enableSecretScanning`, along with a single string key, namely, `repo`. Set `repo` to the name of the repository name where you would like to run this script. Set `enableDependabot` to `true` if you would also like to enable Dependabot Alerts on that repo; set it to `false` if you do not want to enable Dependabot Alerts. The same goes for `enableDependabotUpdates` for Dependabot Security Updates, `enableSecretScanning` for Secret Scanning, `pushprotection` for Secret Scanning push protection, and `enableCodeScanning` for Code Scanning (CodeQL). Finally set `createIssue` to `true` if you would like to create an issue on the repository with the text found in the `./src/utils/text/issueText.ts` file to supplement the PR.

**NOTE:** The account that generated the PAT needs to have `write` access or higher over any repository that you include within the `repos` key.

### Step Two

Run the script which enables Code Scanning (and/or Dependabot Alerts/Dependabot Security Updates/Secret Scanning) on your repository by running:

```bash
yarn run start // or npm run start
```

This will run a script, and you should see output text appearing on your screen.

After the script has run, please head to your `~/Desktop` directory and delete the `tempGitLocations` directory that has been automatically created.

## Running this within a Codespace?

There are some key considerations that you will need to put into place if you are running this script within a GitHub Codespace:

1. You will need to add the following snippet to the `.devcontainer/devcontainer.json`:

```json
  "codespaces": {
    "repositories": [
      {
        "name": "<ORG_NAME>/*",
        "permissions": "write-all"
      }
    ]
  }
```

The reason you need this within your `.devcontainer/devcontainer.json` file is the `GITHUB_TOKEN` tied to the Codespace will need to access other repositories within your organisation which this script may interact with. You will need to create a new Codespace **after** you have added the above and pushed it to your repository.

You do not need to do the above if you are not running it from a Codespace.

## Running as a (scheduled) GitHub workflow

Since this tool uses a PAT or GitHub App Authentication wherever authentication is required, it can be run unattended. You can see in the example
below how you could run the tool in a scheduled GitHub workflow. Instead of using the `.env`
file you can configure all the variables from the `.env.sample` directly as environment variables. This will allow you to
(easily) make use of GitHub action secrets for the PAT or GitHub App credentials.

```yaml
on:
  schedule:
    - cron: "5 16 * * 1"

env:
  APP_ID: ${{ secrets.GHAS_ENABLEMENT_APP_ID }}
  APP_CLIENT_ID: ${{ secrets.GHAS_ENABLEMENT_APP_CLIENT_ID }}
  APP_CLIENT_SECRET: ${{ secrets.GHAS_ENABLEMENT_APP_CLIENT_SECRET }}
  APP_PRIVATE_KEY: ${{ secrets.GHAS_ENABLEMENT_APP_PRIVATE_KEY }}
  ENABLE_ON: "codescanning,secretscanning,pushprotection,dependabot,dependabotupdates"
  DEBUG: "ghas:*"
  CREATE_ISSUE: "false"
  GHES: "false"
  # Organization specific variables
  APP_INSTALLATION_ID: "12345678"
  GITHUB_ORG: "my-target-org"

jobs:
  enable-security-javascript:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
        with:
          repository: NickLiffen/ghas-enablement
      - name: Get dependencies and configure
        run: |
          yarn
          git config --global user.name "ghas-enablement"
          git config --global user.email "ghas.enablement@example.com"
      - name: Enable security on organization (javascript)
        run: |
          npm run getRepos
          npm run start
        env:
          LANGUAGE_TO_CHECK: "javascript"
          TEMP_DIR: ${{ github.workspace }}
```

You can duplicate the last step for the other languages commonly used within your enterprise/organisation.
If you didn't configure the tool as a GitHub App, you can remove all the `APP_*` and set `GITHUB_API_TOKEN` instead.
Above we rely on the sample codeql file for javascript included in this repository. Alternatively you could add this workflow to a repository
containing your customized codeql files and use those to overwrite the samples.

## Found an Issue?

Create an issue within the repository and make it to `@nickliffen`. Key things to mention within your issue:

- Windows, Linux, Codespaces or Mac
- What version of NodeJS you are running.
- Add any logs that appeared when you ran into the issue.

## Want to Contribute?

Great! Open an issue, describe what feature you want to create and make sure to `@nickliffen`.
