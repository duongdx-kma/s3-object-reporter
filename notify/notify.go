package notify

import (
	"os"
	"s3-object-reporter/reports"
	"strings"
	"sync"
)

type sendfunc func(reports.Report) error

func Notify(report reports.Report) []error {
	notifyMethods := strings.Split(os.Getenv("NOTIFY_METHODS"), ",")
	adapters := map[string]sendfunc{
		"influxdb": PushToInfluxDB,
		"smtp":     SMTP,
		"teams":    SendToTeamsAlert,
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(notifyMethods)) // Buffered channel for errors

	for _, method := range notifyMethods {
		method = strings.TrimSpace(method)
		if sendFunc, exists := adapters[method]; exists {
			wg.Add(1)

			go func(f sendfunc) {
				defer wg.Done()
				if err := f(report); err != nil {
					errChan <- err // Send error to channel
				}
			}(sendFunc)
		}
	}

	// Close channel once all Goroutines finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors from channel
	var errorList []error
	for err := range errChan {
		errorList = append(errorList, err)
	}

	return errorList
}
