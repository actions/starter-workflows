# Code Scanning Workflows

GitHub code scanning is a developer-first, GitHub-native approach to easily find security vulnerabilities before they reach production. Before you can configure code scanning for a repository, you must enable code scanning by adding a GitHub Actions workflow to the repository. For more information, see [Setting up code scanning for a repository](https://docs.github.com/en/code-security/secure-coding/setting-up-code-scanning-for-a-repository).
stname: GitHub Actions Demo
on: [push]
jobs:
  Explore-GitHub-Actions:
    runs-on: ubuntu-latest
    steps:
      - run: echo "🎉 The job was automatically triggered by a ${{ github.event_name }} event."
      - run: echo "🐧 This job is now running on a ${{ runner.os }} server hosted by GitHub!"
      - run: echo "🔎 The name of your branch is ${{ github.ref }} and your repository is ${{ github.repository }}."
      - name: Check out repository code
        uses: actions/checkout@v2
      - run: echo "💡 The ${{ github.repository }} repository has been cloned to the runner."
      - run: echo "🖥️ The workflow is now ready to test your code on the runner."
      - name: List files in the repository
        run: |
          ls ${{ github.workspace }}
      - run: echo "🍏 This job's status is ${{ job.status }}."
arter-workflows/code-scanning/README.md
