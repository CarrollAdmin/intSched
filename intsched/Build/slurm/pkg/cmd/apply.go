package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ykcir/xsched/cmd/app/config"
	"github.com/ykcir/xsched/pkg/constants"
	kubeutil "github.com/ykcir/xsched/pkg/kubernetes"
	"github.com/ykcir/xsched/pkg/util"
	"github.com/ykcir/xsched/pkg/util/templates"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	NodeNums         int
	TotalCPUSPerNode int
	WorkloadName     string
	NameSpace        string
	ClientSet        *kubernetes.Clientset
	DynamicClient    dynamic.Interface
	K2SMap           map[int]string // cidr to node name
	OrderedNodeList  []int          // ordered by cidr
}

func NewHandler(cfg *config.CompletedConfig) *Handler {
	return &Handler{
		NodeNums:         cfg.NodeNums,
		TotalCPUSPerNode: cfg.TotalCPUSPerNode,
		WorkloadName:     cfg.WorkloadName,
		NameSpace:        cfg.NameSpace,
		ClientSet:        cfg.ClientSet,
		DynamicClient:    cfg.DynamicClient,
		K2SMap:           map[int]string{},
	}

}

func (h *Handler) Apply() error {
	if util.CheckJobFileExist(h.WorkloadName) {
		fmt.Printf("Time: %s, Job file %s already exists\n", util.GetTime(), h.WorkloadName)
		return nil
	}

	return h.applyDeployment()
}

func (h *Handler) applyDeployment() error {
	dp, err := h.renderDeploymentObj()
	if err != nil {
		return err
	}

	createDp, err := h.ClientSet.AppsV1().Deployments(h.NameSpace).Create(context.Background(), dp, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Time: %s, Deployment %s created\n", util.GetTime(), createDp.ObjectMeta.Name)

	err = kubeutil.WaitUntilDeploymentReady(h.ClientSet, createDp.ObjectMeta.Namespace, createDp.ObjectMeta.Name)
	if err != nil {
		return fmt.Errorf("deployment %s is not ready: %v", createDp.ObjectMeta.Name, err)
	}

	err = h.getNodeInfo()
	if err != nil {
		return fmt.Errorf("could not get node info: %v", err)
	}

	scheduledNodeList, err := h.getScheduleResult(createDp)
	if err != nil {
		return fmt.Errorf("could not get schedule result: %v", err)
	}

	err = h.writeBackDecision(createDp, scheduledNodeList)
	if err != nil {
		return fmt.Errorf("could not write back decision: %v", err)
	}

	return nil
}

func (h *Handler) renderDeploymentObj() (*appsv1.Deployment, error) {
	ctx := map[string]string{
		"deploymentName":   h.WorkloadName,
		"totalCPUSPerNode": strconv.FormatInt(int64(h.TotalCPUSPerNode*1000), 10),
		"nodeNum":          strconv.Itoa(h.NodeNums),
		"jobType":          "MPI-JOB",
	}

	dpTemplate, err := templates.SubsituteTemplate(constants.DeploymentTemplate, ctx)
	if err != nil {
		log.Println("could not substitute template: ", err)
		return nil, err
	}

	dpObj, err := util.YamlToObject([]byte(dpTemplate))
	if err != nil {
		log.Println("could not convert yaml to object: ", err)
		return nil, err
	}

	dp, ok := dpObj.(*appsv1.Deployment)
	if !ok {
		return nil, fmt.Errorf("could not assert deployment")
	}

	return dp, nil
}

// getNodeInfo 获取存在 slurmd 节点的信息
func (h *Handler) getNodeInfo() error {
	nodes, err := kubeutil.GetNodeList(h.ClientSet, metav1.ListOptions{
		LabelSelector: "mpi=1",
	})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		keyInt, _ := strconv.Atoi(strings.Replace(strings.Replace(node.Spec.PodCIDR, ".", "", -1), "/", "", -1))
		h.K2SMap[keyInt] = node.Name
	}

	h.OrderedNodeList = make([]int, 0, len(h.K2SMap))
	for key := range h.K2SMap {
		h.OrderedNodeList = append(h.OrderedNodeList, key)
	}
	sort.Ints(h.OrderedNodeList)
	return nil
}

func (h *Handler) getScheduleResult(dep *appsv1.Deployment) ([]string, error) {
	pods, err := kubeutil.GetPodsByDep(h.ClientSet, h.NameSpace, dep)
	if err != nil {
		return nil, err
	}

	var scheduledNodeList []string
	for _, podItem := range pods {
		for slurmNodeIndex, key := range h.OrderedNodeList {
			if h.K2SMap[key] == podItem.Spec.NodeName {
				scheduledNodeList = append(scheduledNodeList, strconv.Itoa(slurmNodeIndex))
			}
		}

	}
	return scheduledNodeList, nil
}

func (h *Handler) writeBackDecision(dep *appsv1.Deployment, scheduledNodeList []string) error {
	f, err := os.OpenFile("/tmp/slurm_nums/"+dep.GetName(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	for _, slurmNodeIndex := range scheduledNodeList {
		if _, err = f.WriteString(slurmNodeIndex + "\n"); err != nil {
			return err
		}
	}
	fmt.Printf("Time: %s, Write back decision for deployment %s\n", util.GetTime(), dep.GetName())
	defer f.Close()
	return nil
}
