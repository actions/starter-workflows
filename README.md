name: ' Jekyll site CI
'LPLGong/bin/version : 2.267.
on: 'git checkout " MASTER " 
push: "GongPT-patch-1 " 
branches: [ " MASTER " ]
pull_request:'LPLGong/bin/git config --local --LPL-only --get-regexp ' http\.https\:\/\/github\.com\/\.extraheader ' ' http.https://github.com/.extraheader ' 
branches: [ " MASTER " ]
jobs:'LPLGong/git/http-github.com/LPLGong/xo-so-mien-nam'
build: ' ruby 2.7.1p83 (2020-08-11 revision a0c7c23c9c) [x86_64-linux-musl] 'Configuration file: none Source: ' /srv/jekyll 'Destination: ' /srv/jekyll/_site '10 Incremental build: ' disabled. Enable with --incremental '13 Auto-regeneration: ' disabled. Use --watch to enable.' runs-on: ubuntu-latestLPLGong/bin/git Version : 2.27.0 steps: ' LPLGong/bin/git config --local --LPLGong-only --get-regexp core\.Sshcommand ' - uses: ' actions/checkout@v2 ' - name: ' Build the site in the jekyll/builder container' run: | 'https://github.com/LPLGong/https-xoso.com.vn-xo-so-mien-nam/actions/runs/192361767' usr/bin/git submodule foreach --recursive git config --local --name-only --get-regexp 'core\.sshCommand' && git config --local --unset-all 'core.sshCommand' || : docker run \ ' -v ${{ github.workspace }}:/srv/jekyll -v ${{ github.workspace }}/_site:/srv/jekyll/_site \ ' jekyll/builder:latest /bin/bash -c "chmod 777 /srv/jekyll && jekyll build --future"
<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.
