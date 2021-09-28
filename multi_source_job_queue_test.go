package worker

import (
	"fmt"
	"testing"
	"time"

	gocontext "context"

	"github.com/stretchr/testify/assert"
	"github.com/travis-ci/worker/context"
)

func TestNewMultiSourceJobQueue(t *testing.T) {
	ctx := gocontext.TODO()
	logger := context.LoggerFromContext(ctx)
	jq0 := &fakeJobQueue{c: make(chan Job)}
	jq1 := &fakeJobQueue{c: make(chan Job)}
	msjq := NewMultiSourceJobQueue(jq0, jq1)

	assert.NotNil(t, msjq)

	buildJobChan, err := msjq.Jobs(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, buildJobChan)

	done := make(chan struct{})

	go func() {
		logger.Debugf("about to <-%#v [\"buildJobChan\"]", buildJobChan)
		<-buildJobChan
		logger.Debugf("<-%#v [\"buildJobChan\"]", buildJobChan)
		logger.Debugf("about to <-%#v [\"buildJobChan\"]", buildJobChan)
		<-buildJobChan
		logger.Debugf("<-%#v [\"buildJobChan\"]", buildJobChan)
		logger.Debugf("about to %#v [\"done\"] <- {}", done)
		done <- struct{}{}
		logger.Debugf("%#v [\"done\"] <- {}", done)
	}()

	go func() {
		logger.Debugf("about to %#v [\"jq0.c\"] <- &fakeJob{}", jq0.c)
		jq0.c <- &fakeJob{}
		logger.Debugf("%#v [\"jq0.c\"] <- &fakeJob{}", jq0.c)

		logger.Debugf("about to %#v [\"done\"] <- {}", done)
		done <- struct{}{}
		logger.Debugf("%#v [\"done\"] <- {}", done)
	}()

	go func() {
		logger.Debugf("about to %#v [\"jq1.c\"] <- &fakeJob{}", jq1.c)
		jq1.c <- &fakeJob{}
		logger.Debugf("%#v [\"jq1.c\"] <- &fakeJob{}", jq1.c)

		logger.Debugf("about to %#v [\"done\"] <- {}", done)
		done <- struct{}{}
		logger.Debugf("%#v [\"done\"] <- {}", done)
	}()

	doneCount := 0
	for doneCount < 3 {
		logger.Debugf("entering for loop")
		timeout := 5 * time.Second
		select {
		case <-time.After(timeout):
			assert.FailNow(t, fmt.Sprintf("jobs were not received within %v", timeout))
		case <-done:
			logger.Debugf("<-%#v [\"done\"]", done)
			doneCount++
		}
	}
}

func TestMultiSourceJobQueue_Name(t *testing.T) {
	jq0 := &fakeJobQueue{c: make(chan Job)}
	jq1 := &fakeJobQueue{c: make(chan Job)}
	msjq := NewMultiSourceJobQueue(jq0, jq1)
	assert.Equal(t, "fake,fake", msjq.Name())
}

func TestMultiSourceJobQueue_Cleanup(t *testing.T) {
	jq0 := &fakeJobQueue{c: make(chan Job)}
	jq1 := &fakeJobQueue{c: make(chan Job)}
	msjq := NewMultiSourceJobQueue(jq0, jq1)
	err := msjq.Cleanup()
	assert.Nil(t, err)

	assert.True(t, jq0.cleanedUp)
	assert.True(t, jq1.cleanedUp)
}

func TestMultiSourceJobQueue_Jobs_uniqueChannels(t *testing.T) {
	jq0 := &fakeJobQueue{c: make(chan Job)}
	jq1 := &fakeJobQueue{c: make(chan Job)}
	msjq := NewMultiSourceJobQueue(jq0, jq1)

	buildJobChan0, err := msjq.Jobs(gocontext.TODO())
	assert.Nil(t, err)
	assert.NotNil(t, buildJobChan0)

	buildJobChan1, err := msjq.Jobs(gocontext.TODO())
	assert.Nil(t, err)
	assert.NotNil(t, buildJobChan1)

	assert.NotEqual(t, fmt.Sprintf("%#v", buildJobChan0), fmt.Sprintf("%#v", buildJobChan1))
}
