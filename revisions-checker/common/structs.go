package common

import (
	"google.golang.org/api/run/v2"
	"time"
)

// MessagePublishedData contains the full Pub/Sub message
// See the documentation for more details:
// https://cloud.google.com/eventarc/docs/cloudevents#pubsub
type MessagePublishedData struct {
	Message PubSubMessage
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// User defined structures

type RevisionWithTime struct {
	Revision *run.GoogleCloudRunV2Revision
	Time     time.Time
}

type Revision struct {
	Name         string
	CreationTime time.Time
	Image        string
}

type Service struct {
	Name         string
	CreationTime string
	URL          string
}
