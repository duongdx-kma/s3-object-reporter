package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"s3-object-reporter/infra"
	"s3-object-reporter/notify"
	"s3-object-reporter/reports"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

// Handler function for AWS Lambda
func handler(ctx context.Context) (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	roleArn := os.Getenv("AWS_ROLE_ARN")
	bucketName := os.Getenv("S3_BUCKET_NAME")

	if roleArn == "" || bucketName == "" {
		log.Fatalf("ERROR: AWS_ROLE_ARN or S3_BUCKET_NAME is not set")
	}

	awsConfig, err := infra.AssumeRole(ctx, roleArn)
	if err != nil {
		log.Fatalf("ERROR: Failed to assume role: %v", err)
	}

	report, err := reports.GenerateReport(ctx, awsConfig, bucketName)
	fmt.Println("report list:", report)

	if err != nil {
		log.Fatalf("ERROR: Failed to generate report: %v", err)
	}

	errors := notify.Notify(*report)

	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("ERROR: Notify failed - %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("INFO: Report successfully processed and notified.")

	return "Report successfully processed and notified", nil
}

func main() {
	lambda.Start(handler)
}
