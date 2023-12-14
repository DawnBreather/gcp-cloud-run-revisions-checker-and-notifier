package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	run "google.golang.org/api/run/v2"
)

type RevisionWithTime struct {
	Revision *run.GoogleCloudRunV2Revision
	Time     time.Time
}

type Revision struct {
	Name         string
	CreationTime time.Time
	Image        string
}

// var activeRevisions []*run.GoogleCloudRunV2Revision
var revisionsWithTime []RevisionWithTime

type configuration struct {
	ProjectID   string
	ServiceName string
	Region      string
}

var config = configuration{
	ProjectID:   "nzdpu-develop",
	ServiceName: "nzdpu-wis",
	Region:      "us-central1",
}

func main() {
	ctx := context.Background()

	activeRevisions, threeMostRecentRevisions := listRevisions(ctx, config)
	submitToFirestore(ctx, config, activeRevisions, threeMostRecentRevisions)
}

func listRevisions(ctx context.Context, config configuration) (activeRevisions, threeMostRecentRevisions []Revision) {

	// Replace with your Google Cloud project ID and service name

	//projectID := os.Getenv("PROJECT_ID")
	//serviceName := os.Getenv("SERVICE_")
	//region := "us-central1"

	// Create a new Cloud Run service client
	// Make sure you have authenticated with appropriate permissions
	srv, err := run.NewService(ctx, option.WithEndpoint("https://"+config.Region+"-run.googleapis.com/"))
	if err != nil {
		log.Fatalf("run.NewService: %v", err)
	}

	// Build the request to list revisions
	revisionsService := run.NewProjectsLocationsServicesRevisionsService(srv)
	call := revisionsService.List(fmt.Sprintf("projects/%s/locations/%s/services/%s", config.ProjectID, config.Region, config.ServiceName))

	// Make the API request to list revisions
	resp, err := call.Do()
	if err != nil {
		log.Fatalf("Failed to list revisions: %v", err)
	}

	//// Output the revisions
	//for _, revision := range resp.Revisions {
	//	if strings.Contains(revision.Name, "nzdpu-wis-00009-8g8") || strings.Contains(revision.Name, "nzdpu-wis-00008-w79") {
	//
	//		fmt.Printf("Revision: %s\n", revision.Name)
	//		jsonBytes, _ := revision.MarshalJSON()
	//		fmt.Printf("%s\n\n\n", string(jsonBytes))
	//	}
	//}

	for _, revision := range resp.Revisions {
		var isActive bool
		var creationTime time.Time
		for _, condition := range revision.Conditions {
			if condition.Type == "Active" && condition.State == "CONDITION_SUCCEEDED" {
				isActive = true
			}
		}
		if isActive {
			creationTime, err = time.Parse(time.RFC3339, revision.CreateTime)
			if err != nil {
				log.Fatalf("Error parsing creation time: %v", err)
			}

			activeRevisions = append(activeRevisions, Revision{
				Name:         revision.Name,
				CreationTime: creationTime,
				Image:        revision.Containers[0].Image,
			})
			//activeRevisions = append(activeRevisions, revision)
		}

		creationTime, err = time.Parse(time.RFC3339, revision.CreateTime)
		if err != nil {
			log.Fatalf("Error parsing creation time: %v", err)
		}
		revisionsWithTime = append(revisionsWithTime, RevisionWithTime{Revision: revision, Time: creationTime})
	}

	// Sort revisions by creation time in descending order
	sort.Slice(revisionsWithTime, func(i, j int) bool {
		return revisionsWithTime[i].Time.After(revisionsWithTime[j].Time)
	})

	// Select the three most recent revisions
	//var threeMostRecentRevisions []*run.GoogleCloudRunV2Revision
	//var threeMostRecentRevisions []Revision
	for i := 0; i < len(revisionsWithTime) && i < 3; i++ {
		//threeMostRecentRevisions = append(threeMostRecentRevisions, revisionsWithTime[i].Revision)
		threeMostRecentRevisions = append(threeMostRecentRevisions, Revision{
			Name:         revisionsWithTime[i].Revision.Name,
			CreationTime: revisionsWithTime[i].Time,
			Image:        revisionsWithTime[i].Revision.Containers[0].Image,
		})
	}

	// Output the active revisions
	//fmt.Println("Active Revisions:")
	//for _, revision := range activeRevisions {
	//	fmt.Println(revision.Name)
	//}

	// Output the three most recent revisions
	//fmt.Println("\nThree Most Recent Revisions:")
	//for _, revision := range threeMostRecentRevisions {
	//	fmt.Println(revision.Name)
	//}

	return
}

// func submitToFirestore(ctx context.Context, projectID string, serviceName string, activeRevisions, threeMostRecentRevisions []*run.GoogleCloudRunV2Revision) {
func submitToFirestore(ctx context.Context, config configuration, activeRevisions, threeMostRecentRevisions []Revision) {
	// Firestore setup
	firestoreClient, err := firestore.NewClient(context.Background(), config.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Push active revisions to Firestore
	activeRevisions[0].Name = activeRevisions[0].Name + "_manual-prefix" + time.Now().String()
	_, err = firestoreClient.Collection("revisions").Doc(fmt.Sprintf("active.revisions.%s", config.ServiceName)).Set(ctx, map[string]interface{}{
		"revisions": activeRevisions,
	})
	if err != nil {
		log.Fatalf("Failed to write active revisions to Firestore: %v", err)
	}

	// Push most recent revisions to Firestore
	_, err = firestoreClient.Collection("revisions").Doc(fmt.Sprintf("mostrecent.revisions.%s", config.ServiceName)).Set(ctx, map[string]interface{}{
		"revisions": threeMostRecentRevisions,
	})
	if err != nil {
		log.Fatalf("Failed to write most recent revisions to Firestore: %v", err)
	}

	fmt.Println("Revisions successfully written to Firestore")
}
