package helm

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"time"
)

type Status struct {
	statusClient *action.Status
	releaseName  string
}

func NewStatus(cfg *action.Configuration, releaseName string, set *cli.EnvSettings) (*Status, error) {
	client := action.NewStatus(cfg)
	log.WithTime(time.Now()).Debug("We use chart name for deployment: " + "** " + releaseName + "\nsettings: " + fmt.Sprintf("%+v", set) + " **")
	return &Status{
		statusClient: client,
		releaseName:  releaseName,
	}, nil

}

func (status *Status) InfoStatus() (*release.Release, error) {
	results, err := status.statusClient.Run(status.releaseName)
	return results, err
}
