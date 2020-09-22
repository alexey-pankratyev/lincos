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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"io"
	"time"
)

var (
	settings = cli.New()
)

func newHelmInitCmd(out io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "helm",
		Short: "Run helm commands",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(" helm (deploy)")
		},
	}

	flags := cmd.PersistentFlags()

	settings.AddFlags(flags)

	actionConfig := new(action.Configuration)

	cmd.AddCommand(newDeployCmd(actionConfig, out))
	debug("RunDeploy: test2")
	return cmd
}

// Setting up logger
func setLogger() {
	log.SetLevel(log.InfoLevel)
	if settings.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

// Debug function
func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.WithTime(time.Now()).Debug(2, fmt.Sprintf(format, v...))
	}
}
