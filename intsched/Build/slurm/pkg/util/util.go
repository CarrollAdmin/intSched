package util

import (
	"context"
	"os"
	"os/exec"
	"time"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

const ShellToUse = "/bin/sh"

func Shellout(command string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, ShellToUse, "-c", command).Run(); err != nil {

	}
}

// YamlToObject deserializes object in yaml format to a runtime.Object
func YamlToObject(yamlContent []byte) (k8sruntime.Object, error) {
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode(yamlContent, nil, nil)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// GetTime return timestamp in string
func GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// CheckJobFileExist checks if a file exists
func CheckJobFileExist(fileName string) bool {
	filePath := "/tmp/slurm_nums/" + fileName
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
