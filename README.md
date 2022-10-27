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
* `categories`: the categories that it will be shown under. Choose at least one category from the list [here](#categories). Further, choose the categories from the list of languages available [here](https://github.com/github/linguist/blob/master/lib/linguist/languages.yml). When a user views the available templates, those templates that match the same language will feature more prominently.

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
