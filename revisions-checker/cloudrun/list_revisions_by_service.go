package cloudrun

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/api/run/v2"
	"log"
	. "revisions-checker/common"
	. "revisions-checker/config"
	"revisions-checker/utils"
	"sort"
	"time"
)

func ListRevisions(ctx context.Context, serviceName string, config Configuration) (activeRevisions, threeMostRecentRevisions []Revision) {

	// Create a new Cloud Run service client
	// Make sure you have authenticated with appropriate permissions
	srv, err := run.NewService(ctx, option.WithEndpoint("https://"+config.Region+"-run.googleapis.com/"))
	if err != nil {
		log.Fatalf("run.NewService: %v", err)
	}

	var serviceNameToSearch = config.ServiceName
	if serviceName != "" {
		serviceNameToSearch = serviceName
	}

	// Build the request to list revisions
	revisionsService := run.NewProjectsLocationsServicesRevisionsService(srv)
	call := revisionsService.List(fmt.Sprintf("projects/%s/locations/%s/services/%s", config.ProjectID, config.Region, serviceNameToSearch))

	// Make the API request to list revisions
	resp, err := call.Do()
	if err != nil {
		log.Fatalf("Failed to list revisions: %v", err)
	}

	var revisionsWithTime = []RevisionWithTime{}

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
				Name:         utils.ExtractShortServiceName(revision.Name),
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
			Name:         utils.ExtractShortServiceName(revisionsWithTime[i].Revision.Name),
			CreationTime: revisionsWithTime[i].Time,
			Image:        revisionsWithTime[i].Revision.Containers[0].Image,
		})
	}

	return
}
