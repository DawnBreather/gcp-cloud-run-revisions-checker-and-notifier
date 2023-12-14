// Package hellofirestore contains a Cloud Event Function triggered by a Cloud Firestore event.
package hellofirestore

//package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/golang/protobuf/proto"
	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"log"
	"net/http"
	"time"
)

// Define structs to match the nested JSON structure.
type FirestoreTimestamp struct {
	Seconds int64 `json:"seconds"`
	Nanos   int32 `json:"nanos"`
}

type FirestoreValueType struct {
	StringValue    *string              `json:"StringValue,omitempty"`
	TimestampValue *FirestoreTimestamp  `json:"TimestampValue,omitempty"`
	MapValue       *FirestoreMapValue   `json:"MapValue,omitempty"`
	ArrayValue     *FirestoreArrayValue `json:"ArrayValue,omitempty"`
}

type FirestoreValue struct {
	ValueType *FirestoreValueType `json:"ValueType,omitempty"`
}

type FirestoreMapValue struct {
	Fields map[string]FirestoreValue `json:"fields"`
}

type FirestoreArrayValue struct {
	Values []FirestoreValue `json:"values"`
}

type FirestoreRevisions struct {
	Revisions FirestoreValue `json:"revisions"`
}

type Revision struct {
	Name         string    `json:"name"`
	CreationTime time.Time `json:"creationTime"`
	Image        string    `json:"image"`
}

func init() {
	functions.CloudEvent("HelloFirestore", helloFirestore)
}

// helloFirestore is triggered by a change to a Firestore document.
func helloFirestore(ctx context.Context, event event.Event) error {
	// Unmarshal the event data into a firestore.DocumentEventData structure.
	var data firestoredata.DocumentEventData
	//if err := e.DataAs(&data); err != nil {
	//	log.Printf("Error unmarshalling data: %v", err)
	//	return err
	//}

	if err := proto.Unmarshal(event.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	// Log the change for debugging purposes.
	log.Printf("Function triggered by change to: %v", event.Subject())
	log.Printf("Old value: %+v", data.OldValue)
	log.Printf("New value: %+v", data.Value)

	// Marshal the Firestore 'Fields' into a JSON string.
	revisionJson, err := json.Marshal(data.Value.Fields)
	if err != nil {
		log.Printf("Error marshalling Firestore 'Fields': %v", err)
		return err
	}

	var fsRevisions FirestoreRevisions
	err = json.Unmarshal(revisionJson, &fsRevisions)
	if err != nil {
		log.Printf("Error unmarshalling Firestore document: %v\n", err)
		return err
	}

	// Convert FirestoreDocument to a slice of Revision.
	var revisions []Revision
	if fsRevisions.Revisions.ValueType.ArrayValue != nil {
		for _, revValue := range fsRevisions.Revisions.ValueType.ArrayValue.Values {
			if revValue.ValueType.MapValue != nil {
				revision, err := processFirestoreValue(revValue)
				if err != nil {
					log.Printf("Error processing Firestore value: %v", err)
					continue
				}
				revisions = append(revisions, revision)
			}
		}
	}

	// Now you have the revisions as a []Revision slice.
	// You can now handle the revisions as needed by your application logic.

	// Example: Print the revisions
	for _, rev := range revisions {
		fmt.Printf("Revision: %+v\n", rev)
		webhookURL := "https://hooks.slack.com/services/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" // Replace with your Slack webhook URL
		message := fmt.Sprintf("Revision: %+v\n", rev)
		err := SendSlackNotification(webhookURL, message)
		if err != nil {
			fmt.Printf("Error sending notification to Slack: %s\n", err)
		}
	}

	return nil
}

// SlackRequestBody is the request payload that Slack expects for an Incoming Webhook.
type SlackRequestBody struct {
	Text string `json:"text"`
}

// SendSlackNotification sends a message to a Slack channel.
func SendSlackNotification(webhookURL string, msg string) error {
	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return fmt.Errorf("Non-ok response returned from Slack")
	}
	defer resp.Body.Close()

	return nil
}

func main() {
	test()
	//webhookURL := "https://hooks.slack.com/services/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" // Replace with your Slack webhook URL
	//message := "Hello, this is a test notification from my Go program!"
	//err := SendSlackNotification(webhookURL, message)
	//if err != nil {
	//	fmt.Printf("Error sending notification to Slack: %s\n", err)
	//}
}

var jsonData = `{"revisions":{"ValueType":{"ArrayValue":{"values":[{"ValueType":{"MapValue":{"fields":{"CreationTime":{"ValueType":{"TimestampValue":{"seconds":1701374149,"nanos":453921000}}},"Image":{"ValueType":{"StringValue":"us-central1-docker.pkg.dev/nzdpu-develop/nzdpu-2547di-docker-internal/nzdpu-wis@sha256:4d73f72835c8477600f3a1fd7d0b3c0abe868b9f5f733e3cf470d2f3283beb3b"}},"Name":{"ValueType":{"StringValue":"projects/nzdpu-develop/locations/us-central1/services/nzdpu-wis/revisions/nzdpu-wis-00001-6fh_manual-prefix2023-12-05 01:25:24.557249 +0600 +06 m=+2.590374376"}}}}}}]}}}}`

func test() (err error) {
	// Unmarshal the Firestore data into FirestoreRevisions.
	var fsRevisions FirestoreRevisions
	err = json.Unmarshal([]byte(jsonData), &fsRevisions)
	if err != nil {
		log.Printf("Error unmarshalling Firestore document: %v\n", err)
		return err
	}

	//fmt.Println(fsRevisions.Revisions)

	//fmt.Printf("revisions: %+v", fsDoc)
	//fmt.Printf("revisions: %+v", fsDoc.Revisions)

	// Convert FirestoreDocument to a slice of Revision.
	// Process the FirestoreValue to extract Revision data.
	var revisions []Revision
	if fsRevisions.Revisions.ValueType.ArrayValue != nil {
		for _, revValue := range fsRevisions.Revisions.ValueType.ArrayValue.Values {
			if revValue.ValueType.MapValue != nil {
				revision, err := processFirestoreValue(revValue)
				if err != nil {
					log.Printf("Error processing Firestore value: %v", err)
					continue
				}
				revisions = append(revisions, revision)
			}
		}
	}

	// Now you have the revisions as a []Revision slice.
	// You can now handle the revisions as needed by your application logic.

	// Example: Print the revisions
	for _, rev := range revisions {
		fmt.Printf("Revision: %+v\n", rev)
	}

	return nil
}

func processFirestoreValue(fv FirestoreValue) (Revision, error) {
	var revision Revision

	for key, value := range fv.ValueType.MapValue.Fields {
		switch key {
		case "Name":
			if value.ValueType.StringValue != nil {
				revision.Name = *value.ValueType.StringValue
			}
		case "Image":
			if value.ValueType.StringValue != nil {
				revision.Image = *value.ValueType.StringValue
			}
		case "CreationTime":
			if value.ValueType.TimestampValue != nil {
				revision.CreationTime = time.Unix(value.ValueType.TimestampValue.Seconds, int64(value.ValueType.TimestampValue.Nanos))
			}
		}
	}

	return revision, nil
}
