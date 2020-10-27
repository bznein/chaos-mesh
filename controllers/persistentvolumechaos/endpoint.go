package persistentvolumechaos

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
	"golang.org/x/sync/errgroup"

	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

type endpoint struct {
	ctx.Context
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	pvchaos, ok := chaos.(*v1alpha1.PersistentVolumeChaos)
	if !ok {
		err := errors.New("chaos is not PersistentVolumeChaos")
		e.Log.Error(err, "chaos is not PersistentVolumeChaos", "chaos", chaos)
		return err
	}
	pvs, err := utils.SelectAndFilterPV(ctx, e.Client, e.Reader, &pvchaos.Spec)
	if err != nil {
		e.Log.Error(err, "fail to select pv")
		return err
	}
	g := errgroup.Group{}

	for index := range pvs {
		pv := &pvs[index]
		g.Go(func() error {
			e.Log.Info("Deleting pv", "name", pv.Name)
			e.Delete(ctx, pv, &client.DeleteOptions{})
			return nil
		})
	}

	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.PersistentVolumeChaos{}
}

func init() {
	router.Register("persistentvolumechaos", &v1alpha1.PersistentVolumeChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
