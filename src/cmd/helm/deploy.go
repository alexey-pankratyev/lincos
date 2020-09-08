/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	_ "fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"lincos/pkg/helm"
	"os"
	"time"
)

const repo = "https://artifactory.wgdp.io/wtp-helm"

var (
	version   string
	namespace string
	chart     string
	DeployCmd = &cobra.Command{
		Use: "deploy",
		//PreRun: Valid,
		Short: "Run Deploy of helm commands",
		Run:   RunDeploy,
	}
)

func Valid(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Error("Nothing todo. Exit")
		os.Exit(1)
	}
}

func init() {

	DeployCmd.PersistentFlags().StringVarP(&version, "version", "v", "", "version for installation")
	DeployCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "set up namespace")
	DeployCmd.PersistentFlags().StringVarP(&chart, "chart", "c", "", "chart name")
	DeployCmd.Flags().StringP("values", "f", "", "the value file")

	//DeployCmd.PersistentFlags().StringP("context", "c", "", "version for installation")
	DeployCmd.Flags().Bool("dry-run", false, "It's boolean parameters bu default is false it performs without deployment, only a trial run")

}

func RunDeploy(cmd *cobra.Command, args []string) {

	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	statusHelmChart, err := helm.NewStatus(actionConfig, chart)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	infoStatusResult, err := statusHelmChart.InfoStatus()

	if infoStatusResult == nil {
		log.WithTime(time.Now()).WithFields(log.Fields{
			"chart":  chart,
			"status": infoStatusResult,
			"Error":  err,
		}).Info("Chart isn't deployed we will install now.")

	}

	log.WithTime(time.Now()).Debug("To check if chart exists: " + "** " + fmt.Sprintf("%+v", infoStatusResult.Info.Status) + " **")

}
