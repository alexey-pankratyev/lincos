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
	"fmt"
	"github.com/lincos/pkg/lincoshelm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
	"io"
	"os"
	"strings"
	"time"
)

var (
	valueOpts = &values.Options{}
)

type statusPrinter struct {
	release         *release.Release
	debug           bool
	showDescription bool
}

func (s *statusPrinter) WriteJSON(out io.Writer) error {
	debug("RunDeploy: WriteJSON")
	return output.EncodeJSON(out, s.release)
}

func (s *statusPrinter) WriteYAML(out io.Writer) error {
	debug("RunDeploy: WriteYAML")
	return output.EncodeYAML(out, s.release)
}

func (s *statusPrinter) WriteTable(out io.Writer) error {
	debug("RunDeploy: WriteTable")
	if s.release == nil {
		return nil
	}
	fmt.Fprintf(out, "NAME: %s\n", s.release.Name)
	if !s.release.Info.LastDeployed.IsZero() {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", s.release.Info.LastDeployed.Format(time.ANSIC))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", s.release.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", s.release.Info.Status.String())
	fmt.Fprintf(out, "REVISION: %d\n", s.release.Version)
	if s.showDescription {
		fmt.Fprintf(out, "DESCRIPTION: %s\n", s.release.Info.Description)
	}

	executions := executionsByHookEvent(s.release)
	if tests, ok := executions[release.HookTest]; !ok || len(tests) == 0 {
		fmt.Fprintln(out, "TEST SUITE: None")
	} else {
		for _, h := range tests {
			// Don't print anything if hook has not been initiated
			if h.LastRun.StartedAt.IsZero() {
				continue
			}
			fmt.Fprintf(out, "TEST SUITE:     %s\n%s\n%s\n%s\n",
				h.Name,
				fmt.Sprintf("Last Started:   %s", h.LastRun.StartedAt.Format(time.ANSIC)),
				fmt.Sprintf("Last Completed: %s", h.LastRun.CompletedAt.Format(time.ANSIC)),
				fmt.Sprintf("Phase:          %s", h.LastRun.Phase),
			)
		}
	}

	if s.debug {
		fmt.Fprintln(out, "USER-SUPPLIED VALUES:")
		err := output.EncodeYAML(out, s.release.Config)
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)

		cfg, err := chartutil.CoalesceValues(s.release.Chart, s.release.Config)
		if err != nil {
			return err
		}

		fmt.Fprintln(out, "COMPUTED VALUES:")
		err = output.EncodeYAML(out, cfg.AsMap())
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)
	}

	if strings.EqualFold(s.release.Info.Description, "Dry run complete") || s.debug {
		fmt.Fprintln(out, "HOOKS:")
		for _, h := range s.release.Hooks {
			fmt.Fprintf(out, "---\n# Source: %s\n%s\n", h.Path, h.Manifest)
		}
		fmt.Fprintf(out, "MANIFEST:\n%s\n", s.release.Manifest)
	}

	if len(s.release.Info.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", strings.TrimSpace(s.release.Info.Notes))
	}
	return nil
}

func executionsByHookEvent(rel *release.Release) map[release.HookEvent][]*release.Hook {
	result := make(map[release.HookEvent][]*release.Hook)
	for _, h := range rel.Hooks {
		for _, e := range h.Events {
			executions, ok := result[e]
			if !ok {
				executions = []*release.Hook{}
			}
			result[e] = append(executions, h)
		}
	}
	return result
}

func newDeployCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	clientUpgrade := action.NewUpgrade(cfg)
	client := action.NewInstall(cfg)
	var outfmt output.Format
	cmd := &cobra.Command{
		Use: "deploy",
		//PreRun: Valid,
		Short: "Run Deploy of helm commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			rel, err := RunDeploy(args, cfg, client, clientUpgrade, valueOpts, out)
			if err != nil {
				return err
			}
			return outfmt.Write(out, &statusPrinter{rel, settings.Debug, false})
		},
	}
	addInstallAndUpgradeFlags(cmd.Flags(), client, valueOpts)
	addInstallFlags(cmd.Flags(), client, valueOpts)
	addUpgradeFlags(cmd.Flags(), clientUpgrade, valueOpts)
	addChartPathOptionsFlags(cmd.Flags(), &clientUpgrade.ChartPathOptions)
	addValueOptionsFlags(cmd.Flags(), valueOpts)
	bindOutputFlag(cmd, &outfmt)
	bindPostRenderFlag(cmd, &client.PostRenderer)
	flags := cmd.PersistentFlags()
	settings.AddFlags(flags)

	return cmd
}

func RunDeploy(args []string, cfg *action.Configuration, client *action.Install, clientUpgrade *action.Upgrade, valueOpts *values.Options, out io.Writer) (*release.Release, error) {

	setLogger()

	if err := cfg.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	name, chart, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	debug("Chart name: \"%s\"", chart)

	statusHelmChart, err := lincoshelm.NewStatus(cfg, name, settings)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	infoStatusResult, _ := statusHelmChart.InfoStatus()

	if infoStatusResult == nil {
		log.WithTime(time.Now()).WithFields(log.Fields{
			"chart":       chart,
			"status":      infoStatusResult,
			"Error":       err,
			"Namespace":   settings.Namespace(),
			"KubeContext": settings.KubeContext,
		}).Info("Chart isn't deployed we will install now.")

		installHelmChart, err := lincoshelm.RunInstall(client, cfg, name, chart, settings, valueOpts, out)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		//debug("Install: %s", installHelmChart)

		return installHelmChart, nil
	}

	debug("To check if chart exists: \"%+v\"", infoStatusResult.Info.Status)

	upgradeHelmChart, err := lincoshelm.RunUpgrade(clientUpgrade, cfg, name, chart, settings, valueOpts, out)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	//debug("Upgrade: %s", upgradeHelmChart)

	return upgradeHelmChart, nil

}
