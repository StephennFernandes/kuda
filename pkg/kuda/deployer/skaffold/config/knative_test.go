package config

import (
	"testing"

	config "github.com/cyrildiagne/kuda/pkg/kuda/config"
	latest "github.com/cyrildiagne/kuda/pkg/kuda/manifest/latest"
	"github.com/google/go-cmp/cmp"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestGenerateKnativeConfig(t *testing.T) {

	name := "test-name"
	cfg := latest.Config{
		Dockerfile: "test-file",
	}
	userCfg := config.UserConfig{
		Namespace: "test-namespace",
		Deployer: config.DeployerType{
			Skaffold: &config.SkaffoldDeployerConfig{
				DockerRegistry: "test-registry",
			},
		},
	}

	result, err := GenerateKnativeConfig(name, cfg, userCfg)
	if err != nil {
		t.Errorf("err")
	}

	assert.Equal(t, result.APIVersion, "serving.knative.dev/v1")
	assert.Equal(t, result.Kind, "Service")

	meta := metav1.ObjectMeta{
		Name:      "test-name",
		Namespace: "test-namespace",
	}
	if diff := cmp.Diff(result.ObjectMeta, meta); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}

	numGPUs, _ := resource.ParseQuantity("1")
	container := corev1.Container{
		Image: "test-registry/test-name",
		Name:  "test-name",
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceName("nvidia.com/gpu"): numGPUs,
			},
		},
	}
	spec := v1.ServiceSpec{
		ConfigurationSpec: v1.ConfigurationSpec{
			Template: v1.RevisionTemplateSpec{
				Spec: v1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{container},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(result.Spec, spec); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}