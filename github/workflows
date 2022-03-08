name: first-workflow
on: [push]
jobs:
  check-python-version:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Python 3.6.9
        uses: actions/setup-python@v2
        with:
          python-version: 3.6.9
      - name: Check Version
        run: python --version
