package main

import (
	"fmt"
	"os"

	"k8s.io/klog"

	"github.com/practo/sqsmv/pkg/cmdutil"
	"github.com/practo/sqsmv/pkg/signals"
	"github.com/practo/sqsmv/pkg/sqsmv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	flagNames := []string{}
	for _, flagName := range flagNames {
		if err := v.BindFlag(flagName); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return v.Cmd
}

func (v *runCmd) run(cmd *cobra.Command, args []string) {
	stopCh := signals.SetupSignalHandler()

	var config sqsmv.Config
	err := viper.Unmarshal(&config)
	if err != nil {
		klog.Fatalf("Unable to unmarshal sqsmv config")
	}

	sqsmv.Run(config.Queues, stopCh)
}
