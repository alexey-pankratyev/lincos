package helm

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"os"
	"time"
)

type Status struct {
	statusClient *action.Status
	chart        string
}

func NewStatus(cfg *action.Configuration, chr string) (*Status, error) {
	client := action.NewStatus(cfg)
	log.WithTime(time.Now()).Debug("To check if chart exists: " + "** " + chr + " **")
	return &Status{
		statusClient: client,
		chart:        chr,
	}, nil

}

func (status *Status) InfoStatus() (string, error) {
	results, err := status.statusClient.Run(status.chart)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
	var res = fmt.Sprintf("%s", results.Info.Status)
	return res, nil
}
