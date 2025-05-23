package kubernetes

import (
	"context"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// CreateClientSet creates a clientset based on the given kubeConfig. If the
// kubeConfig is empty, it will create the clientset based on the in-cluster
// config
func CreateClientSet(kubeConfig string) (*kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func CreateDynamicClient(kubeConfig string) (dynamic.Interface, error) {

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

func GetNameSpace() string {
	var namespaceBytes []byte
	var err error
	if namespaceBytes, err = os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err != nil {
		namespaceBytes = []byte("slurm")
	}
	return string(namespaceBytes)
}

func GetNodeList(clientSet *kubernetes.Clientset, options metav1.ListOptions) (*corev1.NodeList, error) {
	return clientSet.CoreV1().Nodes().List(context.Background(), options)
}

func GetPodsByDep(clientSet *kubernetes.Clientset, namespace string, dep *appsv1.Deployment) ([]corev1.Pod, error) {
	rsIdList := getRsIDsByDeployment(clientSet, dep)
	podsList := make([]corev1.Pod, 0)
	for _, rs := range rsIdList {
		pods := getPodsByReplicaSet(clientSet, rs, namespace)
		podsList = append(podsList, pods...)
	}

	return podsList, nil
}

// getPodsByReplicaSet 根据传入的ReplicaSet查询到需要的pod
func getPodsByReplicaSet(clientSet kubernetes.Interface, rs appsv1.ReplicaSet, ns string) []corev1.Pod {
	pods, err := clientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Error("list pod error: ", err)
		return nil
	}

	ret := make([]corev1.Pod, 0)
	for _, p := range pods.Items {
		// 找到 pod OwnerReferences uid相同的
		if p.OwnerReferences != nil && len(p.OwnerReferences) == 1 {
			if p.OwnerReferences[0].UID == rs.UID {
				ret = append(ret, p)
			}
		}
	}
	return ret

}

// getRsIDsByDeployment 根据传入的dep，获取到相关连的rs列表(滚更后的ReplicaSet就没用了)
func getRsIDsByDeployment(clientSet kubernetes.Interface, dep *appsv1.Deployment) []appsv1.ReplicaSet {
	// 需要使用match labels过滤
	rsList, err := clientSet.AppsV1().ReplicaSets(dep.Namespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels.Set(dep.Spec.Selector.MatchLabels).String(),
		})
	if err != nil {
		klog.Error("list ReplicaSets error: ", err)
		return nil
	}

	ret := make([]appsv1.ReplicaSet, 0)
	for _, rs := range rsList.Items {
		ret = append(ret, rs)
	}
	return ret
}
