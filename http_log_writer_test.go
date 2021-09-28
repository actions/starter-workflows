package worker

import (
	"testing"
	"time"

	gocontext "context"

	"github.com/stretchr/testify/assert"
)

func buildTestHTTPLogWriter() (gocontext.CancelFunc, *httpLogWriter, error) {
	ctx, cancel := gocontext.WithCancel(gocontext.TODO())
	hlw, err := newHTTPLogWriter(
		ctx,
		"https://jobs.example.org/foo",
		"fafafaf",
		1337,
		time.Second)

	hlw.SetMaxLogLength(100)
	hlw.SetCancelFunc(noCancel)

	return cancel, hlw, err
}

func TestNewHTTPLogWriter(t *testing.T) {
	cancel, hlw, err := buildTestHTTPLogWriter()
	defer cancel()

	assert.Nil(t, err)
	assert.NotNil(t, hlw)
	assert.Implements(t, (*LogWriter)(nil), hlw)
}

func TestHTTPLogWriter_Write(t *testing.T) {
	cancel, hlw, _ := buildTestHTTPLogWriter()
	defer cancel()

	assert.NotNil(t, hlw)
	n, err := hlw.Write([]byte("it's a hot one out there"))
	assert.Nil(t, err)
	assert.True(t, n > 0)
}

func TestHTTPLogWriter_Write_HitsMaxLogLength(t *testing.T) {
	cancel, hlw, _ := buildTestHTTPLogWriter()
	defer cancel()

	assert.NotNil(t, hlw)

	hlw.bytesWritten = 1000

	n, err := hlw.Write([]byte("there's a strong wind blowing"))
	assert.Nil(t, err)
	assert.True(t, hlw.MaxLengthReached())
	assert.Equal(t, 0, n)
}

func TestHTTPLogWriter_Write_HitsMaxLogLength_CannotWriteAndClose(t *testing.T) {
	cancel, hlw, _ := buildTestHTTPLogWriter()
	defer cancel()

	assert.NotNil(t, hlw)

	hlw.bytesWritten = 1000
	mbs := hlw.lps.maxBufferSize
	hlw.lps.maxBufferSize = 0
	defer func() { hlw.lps.maxBufferSize = mbs }()

	n, err := hlw.Write([]byte("looks like rain mmm hmm"))
	assert.Nil(t, err)
	assert.True(t, hlw.MaxLengthReached())
	assert.Equal(t, 0, n)
}

func TestHTTPLogWriter_Write_HitsMaxLogLength_CannotWriteAndClose_LogClosed(t *testing.T) {
	cancel, hlw, _ := buildTestHTTPLogWriter()
	defer cancel()

	assert.NotNil(t, hlw)

	hlw.bytesWritten = 1000
	err := hlw.Close()
	assert.Nil(t, err)

	n, err := hlw.Write([]byte("a storm is a comin"))
	assert.NotNil(t, err)
	assert.Equal(t, 0, n)
}
