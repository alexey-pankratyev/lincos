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

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	_ "github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/values"
	"lincos/pkg/helm"
	"os"
	"time"
)

const repo = "https://artifactory.wgdp.io/wtp-helm"

var (
	valueOpts = &values.Options{}
)

func newDeployCmd(cfg *action.Configuration) *cobra.Command {
	client := action.NewInstall(cfg)
	cmd := &cobra.Command{
		Use: "deploy",
		//PreRun: Valid,
		Short: "Run Deploy of helm commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := RunDeploy(args, cfg, client, valueOpts)
			if err != nil {
				return err
			}
			return nil
		},
	}
	addInstallFlags(cmd.Flags(), client, valueOpts)

	flags := cmd.PersistentFlags()
	settings.AddFlags(flags)

	return cmd
}

func addInstallFlags(f *pflag.FlagSet, client *action.Install, valueOpts *values.Options) {
	f.BoolVar(&client.CreateNamespace, "create-namespace", false, "create the release namespace if not present")
	f.BoolVar(&client.DryRun, "dry-run", false, "simulate an install")
	f.BoolVar(&client.DisableHooks, "no-hooks", false, "prevent hooks from running during install")
	f.BoolVar(&client.Replace, "replace", false, "re-use the given name, only if that name is a deleted release which remains in the history. This is unsafe in production")
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&client.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVarP(&client.GenerateName, "generate-name", "g", false, "generate the name (and omit the NAME parameter)")
	f.StringVar(&client.NameTemplate, "name-template", "", "specify template used to name the release")
	f.StringVar(&client.Description, "description", "", "add a custom description")
	f.BoolVar(&client.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&client.DependencyUpdate, "dependency-update", false, "run helm dependency update before installing the chart")
	f.BoolVar(&client.DisableOpenAPIValidation, "disable-openapi-validation", false, "if set, the installation process will not validate rendered templates against the Kubernetes OpenAPI Schema")
	f.BoolVar(&client.Atomic, "atomic", false, "if set, the installation process deletes the installation on failure. The --wait flag will be set automatically if --atomic is used")
	f.BoolVar(&client.SkipCRDs, "skip-crds", false, "if set, no CRDs will be installed. By default, CRDs are installed if not already present")
	f.BoolVar(&client.SubNotes, "render-subchart-notes", false, "if set, render subchart notes along with the parent")
}

func RunDeploy(args []string, cfg *action.Configuration, client *action.Install, valueOpts *values.Options) (string, error) {

	setLogger()

	if err := cfg.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	name, chart, err := client.NameAndChart(args)
	if err != nil {
		return "nil", err
	}
	debug("CHART NAME: %s\n", chart)

	statusHelmChart, err := helm.NewStatus(cfg, name, settings)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	infoStatusResult, err := statusHelmChart.InfoStatus()

	if infoStatusResult == nil {
		log.WithTime(time.Now()).WithFields(log.Fields{
			"chart":       chart,
			"status":      infoStatusResult,
			"Error":       err,
			"Namespace":   settings.Namespace(),
			"KubeContext": settings.KubeContext,
		}).Info("Chart isn't deployed we will install now.")

		installHelmChart, err := helm.RunInstall(client, cfg, name, chart, settings, valueOpts)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		debug("Install: %s", installHelmChart)

		os.Exit(0)
	}

	debug("To check if chart exists:  %+v", infoStatusResult.Info.Status)

	return "ok", nil
}
