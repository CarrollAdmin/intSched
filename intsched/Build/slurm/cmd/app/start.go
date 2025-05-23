package app

import (
	"github.com/spf13/cobra"
	"github.com/ykcir/xsched/cmd/app/config"
	"github.com/ykcir/xsched/cmd/app/options"
	"github.com/ykcir/xsched/pkg/cmd"
)

func NewPluginCommand(stopCh <-chan struct{}) *cobra.Command {
	pluginOptions := options.NewPluginOptions()

	cmd := &cobra.Command{
		Use: "slurm-plugin",
		RunE: func(c *cobra.Command, args []string) error {
			if err := pluginOptions.Validate(); err != nil {
				return err
			}

			cfg, err := pluginOptions.Config()
			if err != nil {
				return err
			}
			if err := Run(cfg.Complete(), stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	pluginOptions.AddFlags(cmd.Flags())
	return cmd
}

func Run(cfg *config.CompletedConfig, _ <-chan struct{}) error {
	handler := cmd.NewHandler(cfg)
	return handler.Apply()
}
