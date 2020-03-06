package main

import (
	"fmt"
	"os"

	"github.com/practo/sqsmv/pkg/cmdutil"
	"github.com/practo/sqsmv/pkg/signals"
	"github.com/practo/sqsmv/pkg/sqsmv"
	"github.com/spf13/cobra"
)

type runCmd struct {
	cmdutil.BaseCmd
}

var (
	runLong    = `Run the sqsmv daemon`
	runExample = `  sqsmv run`
)

func (v *runCmd) new() *cobra.Command {
	v.Init("sqsmv", &cobra.Command{
		Use:     "run",
		Short:   "Run the sqsmv",
		Long:    runLong,
		Example: runExample,
		Run:     v.run,
	})

	// flags := v.Cmd.Flags()

	flagNames := []string{}
	// flagNames := []string{
	// 	"from-aws-region",
	// 	"to-aws-region",
	// }
	// flags.String("from-aws-region", "ap-southeast-1", "aws region from which the queues are being moved")
	// flags.String("to-aws-region", "ap-south-1", "aws region to which the queues are being moved")

	for _, flagName := range flagNames {
		if err := v.BindFlag(flagName); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return v.Cmd
}

func (v *runCmd) run(cmd *cobra.Command, args []string) {
	// fromRegion := v.Viper.GetString("from-aws-region")
	// toRegion := v.Viper.GetString("from-aws-region")

	// // set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	sqsmv.Run(stopCh)
}
