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
	version    string
	namespace  string
	InstallCmd = &cobra.Command{
		Use: "install",
		//PreRun: Valid,
		Short: "Run install of helm commands",
		Run:   RunInstall,
	}
)

func Valid(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Error("Nothing todo. Exit")
		os.Exit(1)
	}
}

func init() {
	InstallCmd.PersistentFlags().StringVarP(&version, "version", "v", "", "version for installation")
	InstallCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "set up namespace")
	InstallCmd.Flags().StringP("values", "f", "", "the value file")
	//InstallCmd.PersistentFlags().StringP("context", "c", "", "version for installation")
	InstallCmd.Flags().Bool("dry-run", false, "It's boolean parameters bu default is false it performs without deployment, only a trial run")
}

func RunInstall(cmd *cobra.Command, args []string) {

	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	statusHelmChart, err := helm.Status(actionConfig, "nps-dptool")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.WithTime(time.Now()).Debug("To check if chart exists: " + "** " + statusHelmChart + " **")
}
