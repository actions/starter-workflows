package worker

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var jsonPayload = `
{
	"type": "test",
	"vm_type": "default",
	"queue": "builds.docker",
	"config": {
		"language": "rust",
		"rust": "stable",
		"branches": {
			"only": ["master"]
		},
		"os": "linux",
		".result": "configured",
		"global_env": [],
		"group": "stable",
		"dist": "precise",
		"arch": "amd64"
	},
	"env_vars": [],
	"job": {
		"id": 191312240,
		"number": "11.1",
		"commit": "8a7bc35dbeac4d6c93c2ab92db864cdd63684b04",
		"commit_range": "32507736dce8...8a7bc35dbeac",
		"commit_message": "Updates Readme",
		"branch": "master",
		"ref": null,
		"tag": null,
		"pull_request": false,
		"state": "queued",
		"secure_env_enabled": true,
		"debug_options": {},
		"queued_at": "2017-01-12T15:00:00Z"
	},
	"source": {
		"id": 191312239,
		"number": "11",
		"event_type": "push"
	},
	"repository": {
		"id": 11342078,
		"github_id": 76659567,
		"slug": "lukaspustina/axfrnotify",
		"source_url": "https://github.com/lukaspustina/axfrnotify.git",
		"api_url": "https://api.github.com/repos/lukaspustina/axfrnotify",
		"last_build_id": 190973807,
		"last_build_number": "10",
		"last_build_started_at": "2017-01-11T14:38:56Z",
		"last_build_finished_at": "2017-01-11T14:40:41Z",
		"last_build_duration": 284,
		"last_build_state": "passed",
		"default_branch": "master",
		"description": "axfrnotify sends an NOTIFY message to a secondary name server to initiate a zone refresh for a specific domain name."
	},
	"ssh_key": null,
	"timeouts": {
		"hard_limit": null,
		"log_silence": null
	},
	"cache_settings": {}
}
`

func TestUnmarshalJobPayload(t *testing.T) {
	var job JobPayload

	err := json.Unmarshal([]byte(jsonPayload), &job)
	assert.NoError(t, err)

	assert.NotNil(t, job.Job.QueuedAt)
	assert.Exactly(t, time.Unix(1484233200, 0).In(time.UTC), *job.Job.QueuedAt)
}
