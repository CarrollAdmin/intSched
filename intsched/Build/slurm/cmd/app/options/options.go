package options

import (
	"github.com/spf13/pflag"
	"github.com/ykcir/xsched/cmd/app/config"

	kubeutil "github.com/ykcir/xsched/pkg/kubernetes"
)

type pluginOptions struct {
	NodeNums         int
	TotalCPUSPerNode int
	WorkloadName     string
	KubeConfig       string
}

func NewPluginOptions() *pluginOptions {
	o := &pluginOptions{}
	return o
}

func (o *pluginOptions) Validate() error {
	return nil
}

func (o *pluginOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.NodeNums, "node-nums", 1, "Number of nodes.")
	fs.IntVar(&o.TotalCPUSPerNode, "total-cpus-per-node", 2, "Total number of cpus per node.")
	fs.StringVar(&o.WorkloadName, "workload-name", "test", "Name of the workload.")
	fs.StringVar(&o.KubeConfig, "kube-config", o.KubeConfig, "Path to the kubeconfig file.")
}

func (o *pluginOptions) Config() (*config.Config, error) {
	var err error
	c := &config.Config{
		NodeNums:         o.NodeNums,
		TotalCPUSPerNode: o.TotalCPUSPerNode,
		WorkloadName:     o.WorkloadName,
	}

	c.ClientSet, err = kubeutil.CreateClientSet(o.KubeConfig)
	if err != nil {
		return nil, err
	}
	c.DynamicClient, err = kubeutil.CreateDynamicClient(o.KubeConfig)
	if err != nil {
		return nil, err
	}
	c.NameSpace = kubeutil.GetNameSpace()
	return c, nil
}
