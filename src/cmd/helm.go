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
	"helm.sh/helm/v3/pkg/cli"
)

var (
	settings = cli.New()
)

func newHelmInitCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "helm",
		Short: "Run helm commands",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("helm (install|config)")
		},
	}
	cmd.AddCommand(newDeployCmd())
	return cmd

}

func init() {
	newHelmInitCmd()

	rootCmd.AddCommand(newHelmInitCmd())
}

// Setting up logger
func setLogger() {
	log.SetLevel(log.InfoLevel)
	if settings.Debug {
		log.SetLevel(log.DebugLevel)
	}
}
