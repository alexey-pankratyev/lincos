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
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/repo"
)

const outputFlag = "output"
const postRenderFlag = "post-renderer"

func addInstallFlags(f *pflag.FlagSet, client *action.Install) {
	f.BoolVar(&client.DryRun, "dry-run", false, "simulate an install")
	f.BoolVar(&client.CreateNamespace, "create-namespace", false, "create the release namespace if not present")
	f.BoolVar(&client.Replace, "replace", false, "re-use the given name, only if that name is a deleted release which remains in the history. This is unsafe in production")
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVarP(&client.GenerateName, "generate-name", "g", false, "generate the name (and omit the NAME parameter)")
	f.StringVar(&client.NameTemplate, "name-template", "", "specify template used to name the release")
	f.StringVar(&client.Description, "description", "", "add a custom description")
	f.BoolVar(&client.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&client.DependencyUpdate, "dependency-update", false, "run helm dependency update before installing the chart")
}

func addUpgradeFlags(f *pflag.FlagSet, clientUpgrade *action.Upgrade) {
	f.BoolVarP(&clientUpgrade.Install, "install", "i", false, "if a release by this name doesn't already exist, run an install")
	f.BoolVar(&clientUpgrade.Recreate, "recreate-pods", false, "performs pods restart for the resource if applicable")
	f.MarkDeprecated("recreate-pods", "functionality will no longer be updated. Consult the documentation for other methods to recreate pods")
	f.BoolVar(&clientUpgrade.Force, "force", false, "force resource updates through a replacement strategy")
	f.BoolVar(&clientUpgrade.DisableHooks, "no-hooks", false, "disable pre/post upgrade hooks")
	f.BoolVar(&clientUpgrade.DisableOpenAPIValidation, "disable-openapi-validation", false, "if set, the upgrade process will not validate rendered templates against the Kubernetes OpenAPI Schema")
	f.BoolVar(&clientUpgrade.SkipCRDs, "skip-crds", false, "if set, no CRDs will be installed when an upgrade is performed with install flag enabled. By default, CRDs are installed if not already present, when an upgrade is performed with install flag enabled")
	f.BoolVar(&clientUpgrade.ResetValues, "reset-values", false, "when upgrading, reset the values to the ones built into the chart")
	f.BoolVar(&clientUpgrade.ReuseValues, "reuse-values", false, "when upgrading, reuse the last release's values and merge in any overrides from the command line via --set and -f. If '--reset-values' is specified, this is ignored")
	f.BoolVar(&clientUpgrade.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVar(&clientUpgrade.Atomic, "atomic", false, "if set, upgrade process rolls back changes made in case of failed upgrade. The --wait flag will be set automatically if --atomic is used")
	f.IntVar(&clientUpgrade.MaxHistory, "history-max", 10, "limit the maximum number of revisions saved per release. Use 0 for no limit")
	f.BoolVar(&clientUpgrade.CleanupOnFail, "cleanup-on-fail", false, "allow deletion of new resources created in this upgrade when upgrade fails")
	f.BoolVar(&clientUpgrade.SubNotes, "render-subchart-notes", false, "if set, render subchart notes along with the parent")
}

func addValueOptionsFlags(f *pflag.FlagSet, v *values.Options) {
	f.StringSliceVarP(&v.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	f.StringArrayVar(&v.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
}

func addChartPathOptionsFlags(f *pflag.FlagSet, c *action.ChartPathOptions) {
	f.StringVar(&c.Version, "version", "", "specify the exact chart version to use. If this is not specified, the latest version is used")
	f.BoolVar(&c.Verify, "verify", false, "verify the package before using it")
	f.StringVar(&c.RepoURL, "repo", "", "chart repository url where to locate the requested chart")
	f.StringVar(&c.Username, "username", "", "chart repository username where to locate the requested chart")
	f.StringVar(&c.Password, "password", "", "chart repository password where to locate the requested chart")
	f.StringVar(&c.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&c.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.BoolVar(&c.InsecureSkipTLSverify, "insecure-skip-tls-verify", false, "skip tls certificate checks for the chart download")
	f.StringVar(&c.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
}

// bindOutputFlag will add the output flag to the given command and bind the
// value to the given format pointer
func bindOutputFlag(cmd *cobra.Command, varRef *output.Format) {
	cmd.Flags().VarP(newOutputValue(output.Table, varRef), outputFlag, "o",
		fmt.Sprintf("prints the output in the specified format. Allowed values: %s", strings.Join(output.Formats(), ", ")))

	err := cmd.RegisterFlagCompletionFunc(outputFlag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var formatNames []string
		for _, format := range output.Formats() {
			if strings.HasPrefix(format, toComplete) {
				formatNames = append(formatNames, format)
			}
		}
		return formatNames, cobra.ShellCompDirectiveNoFileComp
	})

	if err != nil {
		log.Fatal(err)
	}
}

type outputValue output.Format

func newOutputValue(defaultValue output.Format, p *output.Format) *outputValue {
	*p = defaultValue
	return (*outputValue)(p)
}

func (o *outputValue) String() string {
	// It is much cleaner looking (and technically less allocations) to just
	// convert to a string rather than type asserting to the underlying
	// output.Format
	return string(*o)
}

func (o *outputValue) Type() string {
	return "format"
}

func (o *outputValue) Set(s string) error {
	outfmt, err := output.ParseFormat(s)
	if err != nil {
		return err
	}
	*o = outputValue(outfmt)
	return nil
}

func bindPostRenderFlag(cmd *cobra.Command, varRef *postrender.PostRenderer) {
	cmd.Flags().Var(&postRenderer{varRef}, postRenderFlag, "the path to an executable to be used for post rendering. If it exists in $PATH, the binary will be used, otherwise it will try to look for the executable at the given path")
}

type postRenderer struct {
	renderer *postrender.PostRenderer
}

func (p postRenderer) String() string {
	return "exec"
}

func (p postRenderer) Type() string {
	return "postrenderer"
}

func (p postRenderer) Set(s string) error {
	if s == "" {
		return nil
	}
	pr, err := postrender.NewExec(s)
	if err != nil {
		return err
	}
	*p.renderer = pr
	return nil
}

func compVersionFlag(chartRef string, toComplete string) ([]string, cobra.ShellCompDirective) {
	chartInfo := strings.Split(chartRef, "/")
	if len(chartInfo) != 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	repoName := chartInfo[0]
	chartName := chartInfo[1]

	path := filepath.Join(settings.RepositoryCache, helmpath.CacheIndexFile(repoName))

	var versions []string
	if indexFile, err := repo.LoadIndexFile(path); err == nil {
		for _, details := range indexFile.Entries[chartName] {
			version := details.Metadata.Version
			if strings.HasPrefix(version, toComplete) {
				versions = append(versions, version)
			}
		}
	}

	return versions, cobra.ShellCompDirectiveNoFileComp
}
