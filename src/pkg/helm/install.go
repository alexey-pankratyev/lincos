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
	"helm.sh/helm/v3/pkg/release"
	"time"
)

type Instal struct {
	instalClient *action.Install
	releaseName  string
}

func NewInstall(cfg *action.Configuration, releaseName string, set *cli.EnvSettings) (*Instal, error) {
	client := action.NewInstall(cfg)
	log.WithTime(time.Now()).Debug("We use chart name for deployment: " + "** " + releaseName + "\nsettings: " + fmt.Sprintf("%+v", set) + " **")
	return &Instal{
		instalClient: client,
		releaseName:  releaseName,
	}, nil
}

//runInstall
func (instal *Instal) RunInstall() (*release.Release, error) {
	instal.instalClient.ReleaseName = instal.releaseName
	instal.instalClient.Namespace = "dataplatform"

	chartPath := "/tmp/traefik-1.87.2.tgz"
	chart, err := loader.Load(chartPath)
	if err != nil {
		panic(err)
	}
	log.WithTime(time.Now()).Debug("after load chart")
	res, error := instal.instalClient.Run(chart, nil)
	log.WithTime(time.Now()).Info(fmt.Sprintf("Successfully installed release: %+v ", res.Name))
	return res, error
}
