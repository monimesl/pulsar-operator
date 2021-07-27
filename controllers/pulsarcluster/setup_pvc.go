package pulsarcluster

import (
	"context"
	pvc1 "github.com/monimesl/operator-helper/k8s/pvc"
	"github.com/monimesl/operator-helper/reconciler"
	"github.com/monimesl/pulsar-operator/api/v1alpha1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

const (
	brokerSetupPvcSize = "2Gi"
)

// ReconcileSetupPVC reconcile the setup shared volume between the brokers
func ReconcileSetupPVC(ctx reconciler.Context, cluster *v1alpha1.PulsarCluster) error {
	pvc := &v12.PersistentVolumeClaim{}
	return ctx.GetResource(types.NamespacedName{
		Name:      cluster.BrokersSetupPvcName(),
		Namespace: cluster.Namespace,
	}, pvc,
		nil,
		// Not Found
		func() (err error) {
			p := createPersistentVolumeClaim(cluster)
			pvc = &p
			if err = ctx.SetOwnershipReference(cluster, pvc); err == nil {
				ctx.Logger().Info("Creating the pulsar-broker setup PVC",
					"PVC.Name", pvc.GetName(),
					"PVC.Namespace", pvc.GetNamespace())
				if err = ctx.Client().Create(context.TODO(), pvc); err == nil {
					ctx.Logger().Info("PVC creation success.",
						"PVC.Name", pvc.GetName(),
						"PVC.Namespace", pvc.GetNamespace())
				}
			}
			return
		})
}

func createPersistentVolumeClaim(c *v1alpha1.PulsarCluster) v12.PersistentVolumeClaim {
	return pvc1.New(c.Namespace, c.BrokersSetupPvcName(),
		c.CreateLabels(false, nil),
		v12.PersistentVolumeClaimSpec{
			Resources: v12.ResourceRequirements{
				Requests: map[v12.ResourceName]resource.Quantity{
					v12.ResourceStorage: resource.MustParse(brokerSetupPvcSize),
				}},
			AccessModes: []v12.PersistentVolumeAccessMode{v12.ReadWriteMany},
		})
}
