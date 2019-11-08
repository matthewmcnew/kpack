package cnb

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry"
)

type RemoteBuilderCreator struct {
	FetchImage func(secretRef registry.SecretRef, image string) (v1.Image, error) //is secret ref the right abstraction?
	NewStore   func(secretRef registry.SecretRef, storeImage string) (Store, error)
}

func (r *RemoteBuilderCreator) CreateBuilder(customBuilder *v1alpha1.CustomBuilder) (*RemoteBuilderImage, error) {
	var secretRef = registry.SecretRef{
		ServiceAccount: customBuilder.Spec.ServiceAccount,
		Namespace:      customBuilder.Namespace,
	}
	baseImage, err := r.FetchImage(secretRef, customBuilder.Spec.Stack.BaseBuilderImage)
	if err != nil {
		return nil, err
	}

	remoteBuilder, err := NewRemoteBuilderImage(baseImage)
	if err != nil {
		return nil, err
	}

	store, err := r.NewStore(secretRef, customBuilder.Spec.Store.Image)
	if err != nil {
		return nil, err
	}

	for _, group := range customBuilder.Spec.Order {
		buildpacks := make([]RemoteBuildpack, 0, len(group.Group))

		for _, buildpack := range group.Group {
			remoteBuildpack, err := store.FetchBuildpack(buildpack.ID, buildpack.Version)
			if err != nil {
				return nil, err
			}

			buildpacks = append(buildpacks, remoteBuildpack.Optional(buildpack.Optional))
		}
		remoteBuilder.AddGroup(buildpacks...)
	}

	return remoteBuilder, nil
}
