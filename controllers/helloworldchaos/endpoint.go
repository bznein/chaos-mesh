package helloworldchaos

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

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
	hellochaos, ok := chaos.(*v1alpha1.HelloWorldChaos)
	if !ok {
		err := errors.New("chaos is not HelloWorldChaos")
		e.Log.Error(err, "chaos is not HelloWorldChaos", "chaos", chaos)
		return err
	}
	// TODO implement GetSelector instead of direct call
	pvs, err := utils.SelectPersistentVolumes(ctx, e.Client, e.Reader, hellochaos.Spec.Selector)
	if err != nil {
		e.Log.Error(err, "fail to select pv")
		return err
	}
	g := errgroup.Group{}
	for _, pv := range pvs {
		g.Go(func() error {
			e.Log.Info("Matched pv", pv.Name)
			return nil
		})
	}

	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.HelloWorldChaos{}
}

func init() {
	router.Register("helloworldchaos", &v1alpha1.HelloWorldChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
