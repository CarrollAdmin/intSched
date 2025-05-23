package main

import (
	"flag"
	"log"

	"github.com/ykcir/xsched/cmd/app"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	cmd := app.NewPluginCommand(wait.NeverStop)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		log.Fatalf("%s failed: %s", "slurm-plugin", err)
	}
}
