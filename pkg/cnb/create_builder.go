package cnb

import (
	"github.com/buildpack/imgutil/remote"
	"github.com/google/go-containerregistry/pkg/authn"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry"
)

// Given a list of blobs
//inputs
//buildpackage image gcr.io/build/package
//order
//secrets
//tag

type PackageMetadata struct {
	Buildpacks map[string]struct {
		Version map[string]struct {
			DiffID string `json:"layerDiffID"`
		}
	}
}

type PackageMetadaFetcher interface {
	GetBuildpackage(string, authn.Keychain) (PackageMetadata, error)
}

type RemoteBuilderCreator struct {
	KeychainFactory registry.KeychainFactory
	MetadataFetcher PackageMetadaFetcher
}

func (r *RemoteBuilderCreator) createBuilder(customBuilder v1alpha1.CustomBuilder) (v1alpha1.BuilderStatus, error) {
	keychain, err := r.KeychainFactory.KeychainForSecretRef(registry.SecretRef{
		ServiceAccount: customBuilder.Spec.ServiceAccount,
		Namespace:      customBuilder.Namespace,
	})
	if err != nil {
		return v1alpha1.BuilderStatus{}, err
	}

	//fetch metadatadata
	metadata, err := r.MetadataFetcher.GetBuildpackage(customBuilder.Spec.Store.Image, keychain)
	if err != nil {
		return v1alpha1.BuilderStatus{}, err
	}

	image, err := remote.NewImage(customBuilder.Spec.Tag, keychain,
		remote.FromBaseImage(customBuilder.Spec.Stack.BaseBuilderImage),
		remote.WithPreviousImage(customBuilder.Spec.Store.Image))
	if err != nil {
		return v1alpha1.BuilderStatus{}, err
	}

	// filter the metadata based on order
	for _, group := range customBuilder.Spec.Order {
		for _, buildpack := range group.Group {

		}
		//image.ReuseLayer(group.DiffID)
	}

	image.AddLayer()


	return v1alpha1.BuilderStatus{}, nil
}

// Create remote image
// Attach the blobs to the image
// Write some metadata
// push image to registry
//
//pb buildpack upload io.wells.oracle
//
//Store {
//	Images {
//		pivotal.io/buildpackage - > digest
//		gcr.io/wells/oracle-buildpack@sha256:digest ->
//	}
//}
//
//
//
//nodejs-cnb // meta 3 buildpacks
//yarn
//nodeengine
//npm
