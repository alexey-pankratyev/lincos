package helm

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

type Status struct {
	statusClient *action.Status
	releaseName  string
}

func NewStatus(cfg *action.Configuration, releaseName string, set *cli.EnvSettings) (*Status, error) {
	client := action.NewStatus(cfg)
	debug(set, "We use releaseName: \"%s\" to find out status", releaseName)
	return &Status{
		statusClient: client,
		releaseName:  releaseName,
	}, nil
}

func (status *Status) InfoStatus() (*release.Release, error) {
	results, err := status.statusClient.Run(status.releaseName)
	return results, err
}
