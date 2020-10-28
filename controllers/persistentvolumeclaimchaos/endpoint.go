package persistentvolumeclaimchaos

import (
	"context"
	"encoding/json"
	"errors"

	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

type endpoint struct {
	ctx.Context
}

type patch struct {
	Op   string `json:"op"`
	Path string `json:"path"`
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	pvcchaos, ok := chaos.(*v1alpha1.PersistentVolumeClaimChaos)
	if !ok {
		err := errors.New("chaos is not PersistentVolumeClaimChaos")
		e.Log.Error(err, "chaos is not PersistentVolumeClaimChaos", "chaos", chaos)
		return err
	}
	pvcs, err := utils.SelectAndFilterPVC(ctx, e.Client, e.Reader, &pvcchaos.Spec)
	if err != nil {
		e.Log.Error(err, "fail to select pv")
		return err
	}

	g := errgroup.Group{}
	payload := []patch{{
		Op:   "remove",
		Path: "/metadata/finalizers",
	}}
	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		e.Log.Error(err, "failure creating patch")
		return err
	}
	for index := range pvcs {
		pvc := &pvcs[index]
		g.Go(func() error {
			e.Log.Info("Deleting pvc", "name", pvc.Name)
			err := e.Delete(ctx, pvc, &client.DeleteOptions{})
			if err != nil {
				e.Log.Error(err, "Can't delete PVC!")
			}
			if pvcchaos.Spec.RemoveFinalizers {
				e.Log.Info("Removing finalizers")
				err := e.Client.Patch(context.TODO(), pvc, client.ConstantPatch(types.JSONPatchType, payloadBytes))
				if err != nil {
					e.Log.Error(err, "Failed to patch - Finalizers will run")
				}
			}

			return nil
		})
	}

	return nil

}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.PersistentVolumeClaimChaos{}
}

func init() {
	router.Register("persistentvolumeclaimchaos", &v1alpha1.PersistentVolumeClaimChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
