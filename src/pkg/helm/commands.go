package helm

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"os"
)

func Status(config *action.Configuration, chart string) (string, error) {
	client := action.NewStatus(config)
	results, err := client.Run(chart)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
	res := fmt.Sprintf("%s", results.Info.Status)
	return res, nil
}
