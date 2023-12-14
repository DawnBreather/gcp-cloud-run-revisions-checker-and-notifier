// Package helloworld provides a set of Cloud Functions samples.
package main

import (
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"log"
	. "revisions-checker/cloudrun"
	. "revisions-checker/common"
	. "revisions-checker/config"
	. "revisions-checker/firestore"
	"revisions-checker/slack"
	"revisions-checker/utils"
	"sync"
)

// var activeRevisions []*run.GoogleCloudRunV2Revision
var revisionsWithTime []RevisionWithTime

func init() {
	functions.CloudEvent("HelloPubSub", helloPubSub)
}

// helloPubSub consumes a CloudEvent message and extracts the Pub/Sub message.
func helloPubSub(ctx context.Context, e event.Event) error {
	// var msg MessagePublishedData
	// if err := e.DataAs(&msg); err != nil {
	//   return fmt.Errorf("event.DataAs: %v", err)
	// }

	// name := string(msg.Message.Data) // Automatically decoded from base64.
	// if name == "" {
	//   name = "World"
	// }
	// log.Printf("Hello, %s!", name)
	// return nil
	execute(Configuration{
		ProjectID:   "nzdpu-develop",
		ServiceName: "fmlqjj-wis",
		Region:      "us-central1",
		MostRecentRevisionsFirebaseDocumentPrefix: "mostrecent.revisions.",
		ActiveRevisionsFirebaseDocumentPrefix:     "active.revisions.",
		SlackWebhookURL:                           "https://hooks.slack.com/services/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",
	})

	return nil
}

func main() {
	execute(Configuration{
		ProjectID:   "nzdpu-develop",
		ServiceName: "fmlqjj-wis",
		Region:      "us-central1",
		MostRecentRevisionsFirebaseDocumentPrefix: "mostrecent.revisions.",
		ActiveRevisionsFirebaseDocumentPrefix:     "active.revisions.",
		SlackWebhookURL:                           "https://hooks.slack.com/services/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",
	})
}

func execute(config Configuration) {
	ctx := context.Background()

	services, err := ListServices(ctx, config)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var wg sync.WaitGroup // Declare a WaitGroup

	for _, service := range services {
		wg.Add(1)            // Increment the WaitGroup counter
		go func(s Service) { // Pass 'service' as a parameter to avoid capturing the loop variable
			defer wg.Done() // Decrement the counter when the goroutine completes

			fmt.Printf("> Processing %s\n", s.Name)

			previousActiveRevisions, previousThreeMostRecentRevisions, err := FetchPreviousRevisions(ctx, config, utils.ExtractShortServiceName(s.Name))
			if err != nil {
				log.Fatalf("Error fetching last state of revisions from Firebase: %v", err)
			}

			activeRevisions, threeMostRecentRevisions := ListRevisions(ctx, utils.ExtractShortServiceName(s.Name), config)

			deltaActiveRevisions := findNewRevisions(previousActiveRevisions, activeRevisions)
			deltaThreeMostRecentRevisions := findNewRevisions(previousThreeMostRecentRevisions, threeMostRecentRevisions)

			if len(deltaActiveRevisions) > 0 {
				fmt.Println("Identified ", len(deltaActiveRevisions), " delta of active revisions")
				for _, activeRevision := range deltaActiveRevisions {
					err := slack.SendSlackNotification(activeRevision, s, config)
					if err != nil {
						log.Printf("Error sending slack notification: %v", err)
					}
				}
				SubmitActiveRevisionsToFirestore(ctx, config, s, activeRevisions)
			}

			if len(deltaThreeMostRecentRevisions) > 0 {
				fmt.Println("Identified ", len(deltaThreeMostRecentRevisions), " delta of the most recent revisions")
				SubmitThreeMostRecentRevisionsToFirestore(ctx, config, s, threeMostRecentRevisions)
			}

		}(service)

	}

	wg.Wait() // Wait for all goroutines to finish

}

func SubmitActiveRevisionsToFirestore(ctx context.Context, config Configuration, service Service, revisions []Revision) {
	SubmitRevisionsToFirestore(ctx, config, fmt.Sprintf("%s%s", config.ActiveRevisionsFirebaseDocumentPrefix, utils.ExtractShortServiceName(service.Name)), revisions)
}

func SubmitThreeMostRecentRevisionsToFirestore(ctx context.Context, config Configuration, service Service, revisions []Revision) {
	SubmitRevisionsToFirestore(ctx, config, fmt.Sprintf("%s%s", config.MostRecentRevisionsFirebaseDocumentPrefix, utils.ExtractShortServiceName(service.Name)), revisions)
}

func FetchPreviousRevisions(ctx context.Context, config Configuration, serviceName string) (activeRevisions, threeMostRecentRevisions []Revision, err error) {
	threeMostRecentRevisions, err = FetchFirestoreDocument(ctx, config.ProjectID, "revisions", fmt.Sprintf("%s%s", config.MostRecentRevisionsFirebaseDocumentPrefix, serviceName))
	if err != nil {
		return
	}

	activeRevisions, err = FetchFirestoreDocument(ctx, config.ProjectID, "revisions", fmt.Sprintf("%s%s", config.ActiveRevisionsFirebaseDocumentPrefix, serviceName))
	if err != nil {
		return
	}

	return
}

func contains(revisions []Revision, rev Revision) bool {
	for _, r := range revisions {
		if r.Name == rev.Name {
			return true
		}
	}
	return false
}

// Function to find new revisions
func findNewRevisions(previous, current []Revision) []Revision {
	var newRevisions []Revision
	//fmt.Println(len(previous), "(previous) ==? (current)", len(current))
	for _, currRev := range current {
		if !contains(previous, currRev) {
			newRevisions = append(newRevisions, currRev)
		}
	}
	return newRevisions
}
