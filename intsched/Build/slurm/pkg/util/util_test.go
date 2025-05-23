package util

import (
	"testing"

	"github.com/ykcir/xsched/pkg/constants"
	"github.com/ykcir/xsched/pkg/util/templates"
	appsv1 "k8s.io/api/apps/v1"
)

func TestYamlToObject(t *testing.T) {
	ctx := map[string]string{
		"deploymentName":   "default",
		"totalCPUSPerNode": "2000",
		"nodeNum":          "1",
		"jobType":          "MPI-JOB",
	}

	dpTemplate, err := templates.SubsituteTemplate(constants.DeploymentTemplate, ctx)
	if err != nil {
		t.Fatalf("could not substitute template: %v", err)
	}

	t.Logf("dpTemplate: %v", dpTemplate)
	dpObj, err := YamlToObject([]byte(dpTemplate))
	if err != nil {
		t.Fatalf("could not convert yaml to object: %v", err)
	}

	dp, ok := dpObj.(*appsv1.Deployment)
	if !ok {
		t.Fatalf("could not assert deployment")
	}

	t.Logf("dp: %v", dp)
}
