package revisions

import (
	"google.golang.org/api/run/v2"
	"time"
)

type RevisionWithTime struct {
	Revision *run.GoogleCloudRunV2Revision
	Time     time.Time
}

var activeRevisions []*run.GoogleCloudRunV2Revision
var revisionsWithTime []RevisionWithTime
