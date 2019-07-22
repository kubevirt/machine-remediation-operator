package components

import (
	"fmt"

	"kubevirt.io/machine-remediation-operator/pkg/consts"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const namespaceOpenshiftMachineAPI = "openshift-machine-api"

func getImage(name string, imageRepository string, imageTag string) string {
	return fmt.Sprintf("%s/%s:%s", imageRepository, name, imageTag)
}

// NewDeployment returns new deployment object
func NewDeployment(name string, namespace string, imageRepository string, imageTag string, pullPolicy corev1.PullPolicy, verbosity string) *appsv1.Deployment {
	template := newPodTemplateSpec(name, namespace, imageRepository, imageTag, pullPolicy, verbosity)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				consts.LabelKubeVirt: name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					consts.LabelKubeVirt: name,
				},
			},
			Template: *template,
		},
	}
}

func newPodTemplateSpec(name string, namespace string, imageRepository string, imageTag string, pullPolicy corev1.PullPolicy, verbosity string) *corev1.PodTemplateSpec {
	containers := newContainers(name, namespace, imageRepository, imageTag, pullPolicy, verbosity)
	tolerations := []corev1.Toleration{
		{
			Key:    "node-role.kubernetes.io/master",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:      "CriticalAddonsOnly",
			Operator: corev1.TolerationOpExists,
		},
		{
			Key:               "node.kubernetes.io/not-ready",
			Effect:            corev1.TaintEffectNoExecute,
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: pointer.Int64Ptr(120),
		},
		{
			Key:               "node.kubernetes.io/unreachable",
			Effect:            corev1.TaintEffectNoExecute,
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: pointer.Int64Ptr(120),
		},
	}

	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				consts.LabelKubeVirt: name,
			},
		},
		Spec: corev1.PodSpec{
			Containers:   containers,
			NodeSelector: map[string]string{"node-role.kubernetes.io/master": ""},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: pointer.BoolPtr(true),
			},
			ServiceAccountName: name,
			Tolerations:        tolerations,
		},
	}
}

func newContainers(name string, namespace string, imageRepository string, imageTag string, pullPolicy corev1.PullPolicy, verbosity string) []corev1.Container {
	resources := corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceMemory: resource.MustParse("20Mi"),
			corev1.ResourceCPU:    resource.MustParse("10m"),
		},
	}
	args := []string{
		"--logtostderr=true",
		fmt.Sprintf("--v=%s", verbosity),
		fmt.Sprintf("--namespace=%s", namespace),
	}

	containers := []corev1.Container{
		{
			Name:            name,
			Image:           getImage(name, imageRepository, imageTag),
			Command:         []string{fmt.Sprintf("/usr/bin/%s", name)},
			Args:            args,
			Resources:       resources,
			ImagePullPolicy: pullPolicy,
		},
	}
	return containers
}
