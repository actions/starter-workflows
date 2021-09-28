package worker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/require"
	"github.com/travis-ci/worker/config"
)

func TestBuildScriptGenerator(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	gen := NewBuildScriptGenerator(&config.Config{BuildAPIURI: ts.URL})

	script, err := gen.Generate(context.TODO(), &fakeJob{
		rawPayload: simplejson.New(),
	})
	require.Nil(t, err)
	require.Equal(t, []byte("Hello, client\n"), script)
}
