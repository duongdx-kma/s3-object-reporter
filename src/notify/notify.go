package notify

import (
	"os"
	"s3-object-reporter/reports"
	"strings"
)

type sendfunc func(reports.Report) error

func Notify(report reports.Report) []error {
	var errorList []error
	notifyMethods := strings.Split(os.Getenv("NOTIFY_METHODS"), ",")

	adapters := map[string]sendfunc{
		"influxdb": PushToInfluxDB,
		"smtp":     SMTP,
		"teams":    SendToTeamsAlert,
	}

	for _, method := range notifyMethods {
		err := adapters[strings.TrimSpace(method)](report)
		if err != nil {
			if errorList == nil {
				errorList = []error{}
			}
			errorList = append(errorList, err)
		}
	}
	return errorList
}
