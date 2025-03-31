package notify

import (
	"context"
	"fmt"
	"os"
	"time"

	"s3-object-reporter/reports"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// PushToInfluxDB writes backup count data to InfluxDB with timestamp -1 day
func PushToInfluxDB(report reports.Report) error {
	influxURL := os.Getenv("INFLUXDB_URL")
	influxToken := os.Getenv("INFLUXDB_TOKEN")
	org := os.Getenv("INFLUXDB_ORG")
	bucket := os.Getenv("INFLUXDB_BUCKET")

	if influxURL == "" || influxToken == "" || org == "" || bucket == "" {
		return fmt.Errorf("InfluxDB configuration is missing")
	}

	client := influxdb2.NewClient(influxURL, influxToken)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(org, bucket)

	// Get timestamp of yesterday (set time to 00:00:00 UTC)
	yesterday := time.Now().AddDate(0, 0, -1)
	timestamp := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), yesterday.Hour(), yesterday.Minute(), yesterday.Second(), 0, time.UTC)

	// Loop each service and push data into InfluxDB
	for _, service := range report.Services {
		point := influxdb2.NewPoint(
			"backup_count",
			map[string]string{"service_name": service.ServiceName},
			map[string]interface{}{"count": service.Count},
			timestamp,
		)

		err := writeAPI.WritePoint(context.Background(), point)
		if err != nil {
			fmt.Printf("ERROR: Failed to write service %s to InfluxDB: %v\n", service.ServiceName, err)
			continue // Don't stop execution on failure
		}

		fmt.Printf("INFO: Wrote to InfluxDB - Service: %s, Count: %d, Date: %s\n",
			service.ServiceName, service.Count, timestamp.Format("02-01-2006"))
	}

	return nil
}
