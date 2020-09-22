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
	"github.com/spf13/cobra"
	"io"
	"os"
)

type pluginError struct {
	error
	code int
}

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "lincos",
		Version: "0.1",
		Short:   "Run lincos job from command line.",
	}
	// Add subcommands
	cmd.AddCommand(
		newHelmInitCmd(out),
	)
	return cmd, nil
}

func Execute() {

	cmd, err := newRootCmd(os.Stdout, os.Args[1:])
	if err != nil {
		debug("%+v", err)
		os.Exit(1)
	}

	cobra.OnInitialize(func() {
	})

	// Main executor
	if err := cmd.Execute(); err != nil {
		switch e := err.(type) {
		case pluginError:
			os.Exit(e.code)
		default:
			os.Exit(1)
		}
	}
}
