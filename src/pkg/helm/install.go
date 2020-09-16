/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"time"
)

type Install struct {
	installClient *action.Install
	releaseName   string
}

//runInstall
func RunInstall(
	client *action.Install,
	cfg *action.Configuration,
	releaseName string,
	chart string,
	set *cli.EnvSettings,
	valueOpts *values.Options,
) (*release.Release, error) {
	debug(set, "We use chart name for deployment: %s", releaseName)
	client.ReleaseName = releaseName
	client.Namespace = set.Namespace()
	debug(set, "RunInstall Namespace:", client.Namespace)

	cp, err := client.ChartPathOptions.LocateChart(chart, set)
	if err != nil {
		return nil, err
	}

	debug(set, "CHART PATH:", cp)

	p := getter.All(set)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}
	return client.Run(chartRequested, vals)

}

// Debug function
func debug(settings *cli.EnvSettings, format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.WithTime(time.Now()).Debug(2, fmt.Sprintf(format, v...))
	}
}
