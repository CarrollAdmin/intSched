package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/ykcir/xsched/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

func WaitUntilDeploymentReady(c clientset.Interface, namespace, name string) error {
	ctx := context.TODO()
	return wait.Poll(400*time.Millisecond, 5*time.Minute, func() (bool, error) {
		if util.CheckJobFileExist(name) {
			fmt.Printf("Time: %s, Job file %s already exists when wait\n", util.GetTime(), name)
			return true, nil
		}
		if d, err := c.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{}); err == nil {
			return d.Status.Replicas == d.Status.ReadyReplicas && d.Status.Replicas != 0, nil
		}
		return false, nil
	})
}
