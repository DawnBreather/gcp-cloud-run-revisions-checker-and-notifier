package firestore

import (
	fstore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"log"
	. "revisions-checker/common"
	. "revisions-checker/config"
)

// func submitToFirestore(ctx context.Context, projectID string, serviceName string, activeRevisions, threeMostRecentRevisions []*run.GoogleCloudRunV2Revision) {
func SubmitRevisionsToFirestore(ctx context.Context, config Configuration, documentName string, revisions []Revision) {
	// Firestore setup
	firestoreClient, err := fstore.NewClient(context.Background(), config.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	_, err = firestoreClient.Collection("revisions").Doc(documentName).Set(ctx, map[string]interface{}{
		"revisions": revisions,
	})
	if err != nil {
		log.Fatalf("Failed to write revisions to document { %s }: %v", documentName, err)
	}

	fmt.Println("Revisions successfully written to Firestore")
}
