---
name: Code Scanning onboarding
about: Captures all the information and tasks required to onboard a 3rd party project into Code Scanning
title: 'Code Scanning Partner: '
labels: 'code scanning'
assignees: ''

---

:wave: Thanks for your interest in integrating with Code Scanning! To ensure a swift onboarding of your integration, please provide the following `Requested information` and complete the `Action items` below:

## Requested information
- [ ] Name of your integration:
- [ ] Name of your product / company:
- [ ] Description of your integration:
- [ ] Languages supported by your integration:
- [ ] [For integrations leveraging GitHub Actions] PR for your proposed workflow:
- [ ] URL to an SVG logo representing your integration / product / company:

## Action items
- [ ] Apply to join the GitHub Technology Partner Program: [partner.github.com/apply](https://partner.github.com/apply?partnershipType=Technology+Partner)
- [ ] Develop your integration, by _either_ [following this guide for GitHub Actions](https://docs.github.com/en/github/finding-security-vulnerabilities-and-errors-in-your-code/uploading-a-sarif-file-to-github#uploading-a-code-scanning-analysis-with-github-actions), or [integrating directly with the REST API](https://docs.github.com/en/rest/reference/code-scanning#upload-a-sarif-file)
- [ ] [For integrations leveraging GitHub Actions] Submit a PR in this repo for your proposed starter workflow. The workflow should:
	- [ ] Live in [the `code-scanning` directory](https://github.com/actions/starter-workflows/tree/main/code-scanning)
	- [ ] Have a filename that is in accordance with your product / service / business name, in [_kebab-cased_ format](https://en.wikipedia.org/wiki/Kebab_case), with a `.yml` file extension
	- [ ] Include comments describing the workflow’s behavior ([example](https://github.com/actions/starter-workflows/blob/c59b62dee0eae1f9f368b7011cf05c2fc42cf084/code-scanning/codeql.yml#L1-L11))
	- [ ] Trigger on push, pull_request, and schedule events ([example](https://github.com/actions/starter-workflows/blob/c59b62dee0eae1f9f368b7011cf05c2fc42cf084/code-scanning/codeql.yml#L14-L21))
	- [ ] Reference your GitHub Action using a 40-char commit SHA (e.g. `uses: github/codeql-action@a3a8231e64d3db0e7da0f3b56b9521dcccdfe412`)
- [ ] Update the `Requested information` above, ensuring all details are correct
- [ ] When ready, please ping `@actions/advanced-security-code-scanning` in a comment below, for a review :bow:
