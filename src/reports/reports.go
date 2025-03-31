package reports

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"s3-object-reporter/infra"

	aws "github.com/aws/aws-sdk-go-v2/aws"
)

type Report struct {
	ID         int64     `json:"id"`
	ReportName string    `json:"report_name"`
	ReportDate string    `json:"report_date"`
	Services   []Service `json:"services"`
}

type Service struct {
	ID          int64  `json:"id"`
	Count       int64  `json:"count"`
	ServiceName string `json:"service_name"`
	ServicePath string `json:"service_path"`
}

// GenerateReport from object S3 and create services list:
// func GenerateReport(ctx context.Context, awsConfig aws.Config, bucketName string) (*Report, error) {
// 	yesterday := time.Now().AddDate(0, 0, -1)
// 	formattedDate := yesterday.Format("02/01/2006")
// 	year := yesterday.Format("2006")
// 	datePath := yesterday.Format("02012006")

// 	// 1. List all services with prefix "yyyy/"
// 	serviceObjects, err := infra.ListS3ObjectsByDate(ctx, awsConfig, bucketName, year+"/")
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Printf("Warning: serviceObjects list %s\n", serviceObjects)

// 	serviceMap := make(map[string]*Service)

// 	for _, obj := range serviceObjects {
// 		parts := strings.Split(obj, "/")
// 		if len(parts) < 2 {
// 			continue
// 		}
// 		serviceName := parts[1]

// 		// Ensure unique service
// 		if _, exists := serviceMap[serviceName]; !exists {
// 			serviceMap[serviceName] = &Service{
// 				ID:          time.Now().Unix(),
// 				ServiceName: serviceName,
// 				ServicePath: fmt.Sprintf("%s/%s", year, serviceName),
// 				Count:       0,
// 			}
// 		}
// 	}

// 	log.Printf("Warning: serviceMap list %s\n", serviceMap)

// 	report := &Report{
// 		ID:         time.Now().Unix(),
// 		ReportName: fmt.Sprintf("Daily Backup Report: %s", formattedDate),
// 		ReportDate: formattedDate,
// 	}

// 	// 2. List objects for each service under "yyyy/{service}/ddmmyyyy/"
// 	for serviceName, service := range serviceMap {
// 		servicePrefix := fmt.Sprintf("%s/%s/%s/", year, serviceName, datePath)
// 		serviceObjects, err := infra.ListS3ObjectsByDate(ctx, awsConfig, bucketName, servicePrefix)
// 		if err != nil {
// 			log.Printf("Error listing objects for service %s: %v", serviceName, err)
// 			continue
// 		}

// 		for _, obj := range serviceObjects {
// 			parts := strings.Split(obj, "/")
// 			if len(parts) > 4 {
// 				continue // Max depth check
// 			}
// 			service.Count++
// 		}

// 		report.Services = append(report.Services, *service)
// 	}

// 	return report, nil
// }

// GenerateReport from object S3 and create services list:
func GenerateReport(ctx context.Context, awsConfig aws.Config, bucketName string) (*Report, error) {
	yesterday := time.Now().AddDate(0, 0, -1)
	formattedDate := yesterday.Format("02/01/2006")
	year := yesterday.Format("2006")
	datePath := yesterday.Format("02012006")

	// 1. List all service names using delimiter "/"
	servicePrefixes, err := infra.ListS3Prefixes(ctx, awsConfig, bucketName, year+"/", "/")
	if err != nil {
		return nil, err
	}

	log.Printf("Warning: servicePrefixes list %s\n", servicePrefixes)

	serviceMap := make(map[string]*Service)
	for _, servicePrefix := range servicePrefixes {
		parts := strings.Split(strings.TrimSuffix(servicePrefix, "/"), "/")
		if len(parts) < 2 {
			continue
		}
		serviceName := parts[1]
		serviceMap[serviceName] = &Service{
			ID:          time.Now().Unix(),
			ServiceName: serviceName,
			ServicePath: fmt.Sprintf("%s/%s", year, serviceName),
			Count:       0,
		}
	}

	log.Printf("Warning: serviceMap %s\n", serviceMap)

	report := &Report{
		ID:         time.Now().Unix(),
		ReportName: fmt.Sprintf("Daily Backup Report: %s", formattedDate),
		ReportDate: formattedDate,
	}

	// 2. List objects for each service under "yyyy/{service}/ddmmyyyy/" with limited depth
	for serviceName, service := range serviceMap {
		servicePrefix := fmt.Sprintf("%s/%s/%s/", year, serviceName, datePath)
		serviceObjects, err := infra.ListS3ObjectsLimited(ctx, awsConfig, bucketName, servicePrefix, "/")
		log.Printf("Warning: serviceObjects %s\n", serviceObjects)

		if err != nil {
			log.Printf("Error listing objects for service %s: %v", serviceName, err)
			continue
		}

		service.Count = int64(len(serviceObjects))
		report.Services = append(report.Services, *service)
	}

	return report, nil
}

// 2025/03/27 15:26:47 Warning: serviceObjects [2025/growth-rds/26032025/ 2025/growth-rds/26032025/câu chuyện và thực tế.mp4 2025/growth-rds/26032025/muốn và 2025/growth-rds/26032025/racket.mp4]
