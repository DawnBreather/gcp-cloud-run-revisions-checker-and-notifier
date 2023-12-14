package cloudrun

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/api/run/v2"
	"log"
	. "revisions-checker/common"
	. "revisions-checker/config"
)

func ListServices(ctx context.Context, config Configuration) ([]Service, error) {
  // Create a new Cloud Run service client
  srv, err := run.NewService(ctx, option.WithEndpoint("https://"+config.Region+"-run.googleapis.com/"))
  if err != nil {
    log.Fatalf("run.NewService: %v", err)
    return nil, err
  }

  // Build the request to list services
  servicesService := run.NewProjectsLocationsServicesService(srv)
  call := servicesService.List(fmt.Sprintf("projects/%s/locations/%s", config.ProjectID, config.Region))

  // Make the API request to list services
  resp, err := call.Do()
  if err != nil {
    log.Fatalf("Failed to list services: %v", err)
    return nil, err
  }

  var services []Service
  for _, service := range resp.Services {
    services = append(services, Service{
      Name:         service.Name,
      CreationTime: service.CreateTime,
      URL:          service.Uri,
    })
  }

  return services, nil
}
