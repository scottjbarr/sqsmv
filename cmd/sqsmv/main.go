package main

import (
	"k8s.io/klog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/practo/sqsmv/pkg/cmdutil"
)

var rootCmd = &cobra.Command{
	Use:   "sqsmv",
	Short: "sqsmv moves the jobs between sqs queues",
	Long:  "sqsmv moves the jobs between sqs queues",
}

func localFlags(flags *pflag.FlagSet) {
}

func init() {
	cobra.OnInitialize(func() {
		cmdutil.CheckErr(cmdutil.InitConfig("sqsmv"))
	})

	flags := rootCmd.PersistentFlags()
	localFlags(flags)
}

func main() {
	klog.InitFlags(nil)

	versionCommand := (&versionCmd{}).new()
	runCommand := (&runCmd{}).new()

	// add main commands
	rootCmd.AddCommand(
		versionCommand,
		runCommand,
	)

	cmdutil.CheckErr(rootCmd.Execute())
}
