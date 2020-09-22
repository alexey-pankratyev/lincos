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

package cmd

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"io"
)

//Upgrade release
func RunUpgrade(
	clientUpgrade *action.Upgrade,
	cfg *action.Configuration,
	releaseName string,
	chart string,
	valueOpts *values.Options,
	out io.Writer,
) (*release.Release, error) {
	debug( "We use chart name for upgrade: %s", releaseName)

	clientUpgrade.Namespace = settings.Namespace()
	debug("RunInstall Namespace:", clientUpgrade.Namespace)

	chartPath, err := clientUpgrade.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}

	debug( "CHART PATH:", chartPath)

	vals, err := valueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return nil, err
		}
	}

	if ch.Metadata.Deprecated {
		warning("This chart is deprecated")
	}

	return clientUpgrade.Run(releaseName, ch, vals)

}
