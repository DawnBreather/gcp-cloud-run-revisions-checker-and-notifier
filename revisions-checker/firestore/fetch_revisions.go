package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"log"
	. "revisions-checker/common"
)

type RevisionsDocument struct {
	Revisions []Revision `json:"revisions"`
}

func FetchFirestoreDocument(ctx context.Context, projectID, collectionName, documentID string) ([]Revision, error) {
	// Create a Firestore client
	//client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile("path/to/your/service-account-file.json"))
	client, err := firestore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
		return nil, err
	}
	defer client.Close()

	// Reference to the document
	docRef := client.Collection(collectionName).Doc(documentID)

	// Get the document
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("Failed to get document: %v", err)
		return nil, err
	}

	//fmt.Printf("%+v", docSnapshot.Data()["revisions"])
	//jsonBytes, err := json.Marshal(docSnapshot.Data())
	//fmt.Printf("%s", string(jsonBytes))
	//
	//os.Exit(0)

	// Map data to your struct
	var revisionsDoc RevisionsDocument
	err = docSnapshot.DataTo(&revisionsDoc)
	if err != nil {
		log.Printf("Failed to decode document: %v", err)
		return nil, err
	}

	return revisionsDoc.Revisions, nil
}
