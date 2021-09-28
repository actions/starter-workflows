package worker

import (
	"fmt"
	"io"
	"sort"

	"github.com/travis-ci/worker/backend"
	"gopkg.in/urfave/cli.v1"
)

var (
	cliHelpPrinter = cli.HelpPrinter
)

const (
	providerHelpHeader = `
All provider options must be given as environment variables of the form:

   $[TRAVIS_WORKER_]{UPCASE_PROVIDER_NAME}_{UPCASE_UNDERSCORED_KEY}
     ^------------^
   optional namespace

e.g.:

   TRAVIS_WORKER_DOCKER_HOST='tcp://127.0.0.1:4243'
   TRAVIS_WORKER_DOCKER_PRIVILEGED='true'

`
)

func init() {
	cli.HelpPrinter = helpPrinter
}

func helpPrinter(w io.Writer, templ string, data interface{}) {
	cliHelpPrinter(w, templ, data)

	fmt.Fprint(w, providerHelpHeader)

	margin := 4
	maxLen := 0

	backend.EachBackend(func(b *backend.Backend) {
		for itemKey := range b.ProviderHelp {
			if len(itemKey) > maxLen {
				maxLen = len(itemKey)
			}
		}
	})

	itemFmt := fmt.Sprintf("%%%ds - %%s\n", maxLen+margin)

	backend.EachBackend(func(b *backend.Backend) {
		fmt.Fprintf(w, "\n%s provider help:\n\n", b.HumanReadableName)

		sortedKeys := []string{}
		for key := range b.ProviderHelp {
			sortedKeys = append(sortedKeys, key)
		}

		sort.Strings(sortedKeys)

		for _, key := range sortedKeys {
			fmt.Printf(itemFmt, key, b.ProviderHelp[key])
		}
	})

	fmt.Println("")
}
