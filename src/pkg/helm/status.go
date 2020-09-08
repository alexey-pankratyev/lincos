package helm

import (
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"time"
)

type Status struct {
	statusClient *action.Status
	chart        string
}

func NewStatus(cfg *action.Configuration, chr string) (*Status, error) {
	client := action.NewStatus(cfg)
	log.WithTime(time.Now()).Debug("We use chart name for deployment: " + "** " + chr + " **")
	return &Status{
		statusClient: client,
		chart:        chr,
	}, nil

}

func (status *Status) InfoStatus() (*release.Release, error) {
	results, err := status.statusClient.Run(status.chart)
	return results, err
}
