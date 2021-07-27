package pulsarcluster

import (
	"context"
	"fmt"
	"github.com/monimesl/operator-helper/k8s/annotation"
	"github.com/monimesl/operator-helper/k8s/pod"
	"github.com/monimesl/operator-helper/k8s/statefulset"
	"github.com/monimesl/operator-helper/reconciler"
	"github.com/monimesl/pulsar-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

const (
	setupVolume          = "broker-setup"
	setupVolumeMouthPath = "/broker-setup"
)

// ReconcileStatefulSet reconcile the statefulset of the specified cluster
func ReconcileStatefulSet(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	sts := &v1.StatefulSet{}
	return ctx.GetResource(types.NamespacedName{
		Name:      cluster.StatefulSetName(),
		Namespace: cluster.Namespace,
	}, sts,
		// Found
		func() error {
			if *cluster.Spec.Size != *sts.Spec.Replicas {
				if err := updateStatefulset(ctx, sts, cluster); err != nil {
					return err
				}
			}
			return nil
		},
		// Not Found
		func() error {
			sts = createStatefulSet(cluster)
			if err := ctx.SetOwnershipReference(cluster, sts); err != nil {
				return err
			}
			ctx.Logger().Info("Creating the zookeeper statefulset.",
				"StatefulSet.Name", sts.GetName(),
				"StatefulSet.Namespace", sts.GetNamespace())
			if err := ctx.Client().Create(context.TODO(), sts); err != nil {
				return err
			}
			ctx.Logger().Info("StatefulSet creation success.",
				"StatefulSet.Name", sts.GetName(),
				"StatefulSet.Namespace", sts.GetNamespace())
			return nil
		})
}

func updateStatefulset(ctx reconciler.Context, sts *v1.StatefulSet, cluster *v1alpha1.PulsarCluster) error {
	sts.Spec.Replicas = cluster.Spec.Size
	ctx.Logger().Info("Updating the zookeeper statefulset.",
		"StatefulSet.Name", sts.GetName(),
		"StatefulSet.Namespace", sts.GetNamespace(), "NewReplicas", cluster.Spec.Size)
	return ctx.Client().Update(context.TODO(), sts)
}

func createStatefulSet(c *v1alpha1.PulsarCluster) *v1.StatefulSet {
	labels := c.CreateLabels(true, nil)
	templateSpec := createPodTemplateSpec(c, labels)
	spec := statefulset.NewSpec(*c.Spec.Size, c.HeadlessServiceName(), labels, nil, templateSpec)
	sts := statefulset.New(c.Namespace, c.StatefulSetName(), labels, spec)
	annotations := c.Spec.Annotations
	if c.Spec.MonitoringConfig.Enabled &&
		(c.Spec.Ports.Web > 0 || c.Spec.Ports.WebTLS > 0) {
		metricPort := c.Spec.Ports.Web
		if metricPort <= 0 {
			metricPort = c.Spec.Ports.WebTLS
		}
		annotations = annotation.DecorateForPrometheus(
			annotations, true, int(metricPort))
	}
	sts.Annotations = annotations
	return sts
}

func createPodTemplateSpec(c *v1alpha1.PulsarCluster, labels map[string]string) v12.PodTemplateSpec {
	return pod.NewTemplateSpec("", c.StatefulSetName(), labels, nil, createPodSpec(c))
}

func createPodSpec(c *v1alpha1.PulsarCluster) v12.PodSpec {
	setupEnv := []v12.EnvVar{
		{Name: "PULSAR_VERSION", Value: c.Spec.PulsarVersion},
		{Name: "PULSAR_CONNECTORS", Value: generateConnectorString(c)},
		{Name: "PULSAR_SETUP_DIRECTORY", Value: setupVolumeMouthPath},
	}
	envs := processEnvVars(c.Spec.Env)
	volumeMounts := []v12.VolumeMount{
		{Name: setupVolume, MountPath: setupVolumeMouthPath},
	}
	image := c.Image()
	initContainers := []v12.Container{
		{
			Name: "broker-setup",
			Image: fmt.Sprintf("%s:%s",
				v1alpha1.ConnectSetupImageRepository,
				v1alpha1.DefaultConnectorsSetupImageVersion),
			ImagePullPolicy: image.PullPolicy,
			VolumeMounts:    volumeMounts,
			Env:             setupEnv,
		},
	}
	containers := []v12.Container{
		{
			Name:            "pulsar-broker",
			Ports:           createContainerPorts(c),
			Image:           image.ToString(),
			ImagePullPolicy: image.PullPolicy,
			VolumeMounts:    volumeMounts,
			Lifecycle:       &v12.Lifecycle{PreStop: createPreStopHandler()},
			Env:             pod.DecorateContainerEnvVars(true, envs...),
			EnvFrom: []v12.EnvFromSource{
				{
					ConfigMapRef: &v12.ConfigMapEnvSource{
						LocalObjectReference: v12.LocalObjectReference{
							Name: c.ConfigMapName(),
						},
					},
				},
			},
			Command: []string{"sh", "-c"},
			Args: []string{
				strings.Join([]string{
					"rm -rf /pulsar/connectors",
					"cp -r \"$PULSAR_SETUP_DIRECTORY/connectors\" /pulsar",
					"bin/apply-config-from-env.py conf/broker.conf",
					"sleep infinity",
				}, "; "),
			},
		},
	}
	volumes := []v12.Volume{
		{
			Name: setupVolume,
			VolumeSource: v12.VolumeSource{
				EmptyDir: &v12.EmptyDirVolumeSource{},
			},
		},
	}
	spec := pod.NewSpec(c.Spec.PodConfig, volumes, initContainers, containers)
	spec.TerminationGracePeriodSeconds = c.Spec.PodConfig.TerminationGracePeriodSeconds
	return spec
}

func generateConnectorString(c *v1alpha1.PulsarCluster) string {
	formats := make([]string, len(c.Spec.Connectors))
	for i, connector := range c.Spec.Connectors {
		if connector.Builtin != "" {
			formats[i] = connector.Builtin
		} else {
			headers := ""
			for k, v := range connector.Custom.Headers {
				if headers == "" {
					headers += ";"
				}
				headers += fmt.Sprintf("%s:%s", k, v)
			}
			formats[i] = fmt.Sprintf("%s;%s", connector.Custom.URL, headers)
		}
	}
	return strings.Join(formats, " ")
}

func createContainerPorts(c *v1alpha1.PulsarCluster) []v12.ContainerPort {
	ports := c.Spec.Ports
	containerPorts := []v12.ContainerPort{{Name: v1alpha1.ClientPortName, ContainerPort: ports.Client}}
	if ports.ClientTLS > 0 {
		containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.ClientTLSPortName, ContainerPort: ports.ClientTLS})
	}
	if ports.Web > 0 {
		containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.WebPortName, ContainerPort: ports.Web})
	}
	if ports.WebTLS > 0 {
		containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.WebTLSPortName, ContainerPort: ports.WebTLS})
	}
	kop := c.Spec.KOP
	if kop.Enabled {
		if kop.PlainTextPort > 0 {
			containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.KopPlainTextPortName, ContainerPort: kop.PlainTextPort})
		}
		if kop.SecuredPort > 0 {
			containerPorts = append(containerPorts, v12.ContainerPort{Name: v1alpha1.KopSecuredPortName, ContainerPort: kop.SecuredPort})
		}
	}
	return containerPorts
}

func createStartupProbe(probe *pod.Probe) *v12.Probe {
	return probe.ToK8sProbe(v12.Handler{
		Exec: &v12.ExecAction{Command: []string{"/scripts/probeStartup.sh"}},
	})
}
func createReadinessProbe(probe *pod.Probe) *v12.Probe {
	return probe.ToK8sProbe(v12.Handler{
		Exec: &v12.ExecAction{Command: []string{"/scripts/probeReadiness.sh"}},
	})
}

func createLivenessProbe(probe *pod.Probe) *v12.Probe {
	return probe.ToK8sProbe(v12.Handler{
		Exec: &v12.ExecAction{Command: []string{"/scripts/probeLiveness.sh"}},
	})
}

func createPreStopHandler() *v12.Handler {
	return &v12.Handler{Exec: &v12.ExecAction{Command: []string{"/scripts/stop.sh"}}}
}
