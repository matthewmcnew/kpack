package cnb

import (
	"encoding/json"
	"time"

	"github.com/buildpack/lifecycle/metadata"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry"
)

const (
	BuilderMetadataLabel    = "io.buildpacks.builder.metadata"
	BuildPackageLayersLabel = "io.buildpacks.buildpack.layers"
)

type BuilderImage struct {
	BuilderBuildpackMetadata []BuildpackInfo
	Stack                    Stack
	Identifier               string
}

type RemoteMetadataRetriever struct {
	RemoteImageFactory registry.RemoteImageFactory
}

func (r *RemoteMetadataRetriever) GetBuilderImage(builder v1alpha1.BuilderResource) (BuilderImage, error) {
	img, err := r.RemoteImageFactory.NewRemote(builder.Image(), registry.SecretRef{
		Namespace:        builder.GetObjectMeta().GetNamespace(),
		ImagePullSecrets: builder.ImagePullSecrets(),
	})
	if err != nil {
		return BuilderImage{}, errors.Wrap(err, "unable to fetch remote builder baseImage")
	}

	var metadataJSON string
	metadataJSON, err = img.Label(BuilderMetadataLabel)
	if err != nil {
		return BuilderImage{}, errors.Wrap(err, "builder baseImage metadata label not present")
	}

	stackId, err := img.Label(metadata.StackMetadataLabel)
	if err != nil {
		return BuilderImage{}, err
	}

	var md BuilderImageMetadata
	err = json.Unmarshal([]byte(metadataJSON), &md)
	if err != nil {
		return BuilderImage{}, errors.Wrap(err, "unsupported builder metadata structure")
	}

	identifier, err := img.Identifier()
	if err != nil {
		return BuilderImage{}, errors.Wrap(err, "failed to retrieve builder baseImage SHA")
	}

	runImage, err := r.RemoteImageFactory.NewRemote(md.Stack.RunImage.Image, registry.SecretRef{
		Namespace:        builder.GetObjectMeta().GetNamespace(),
		ImagePullSecrets: builder.ImagePullSecrets(),
	})

	if err != nil {
		return BuilderImage{}, errors.Wrap(err, "unable to fetch remote run baseImage")
	}

	runImageIdentifier, err := runImage.Identifier()
	if err != nil {
		return BuilderImage{}, errors.Wrap(err, "failed to retrieve run baseImage SHA")
	}

	return BuilderImage{
		BuilderBuildpackMetadata: md.Buildpacks,
		Stack: Stack{
			RunImage: runImageIdentifier,
			ID:       stackId,
		},
		Identifier: identifier,
	}, nil
}

func (r *RemoteMetadataRetriever) GetBuiltImage(ref *v1alpha1.Build) (BuiltImage, error) {
	img, err := r.RemoteImageFactory.NewRemote(ref.Tag(), registry.SecretRef{
		ServiceAccount: ref.Spec.ServiceAccount,
		Namespace:      ref.Namespace,
	})
	if err != nil {
		return BuiltImage{}, err
	}

	return readBuiltImage(img)
}

type BuiltImage struct {
	Identifier        string
	CompletedAt       time.Time
	BuildpackMetadata []metadata.BuildpackMetadata
	Stack             Stack
}

func readBuiltImage(img registry.RemoteImage) (BuiltImage, error) {
	var buildMetadataJSON string
	var layerMetadataJSON string

	buildMetadataJSON, err := img.Label(metadata.BuildMetadataLabel)
	if err != nil {
		return BuiltImage{}, err
	}

	layerMetadataJSON, err = img.Label(metadata.LayerMetadataLabel)
	if err != nil {
		return BuiltImage{}, err
	}

	stackId, err := img.Label(metadata.StackMetadataLabel)
	if err != nil {
		return BuiltImage{}, nil
	}

	var buildMetadata metadata.BuildMetadata
	var layerMetadata metadata.LayersMetadata

	err = json.Unmarshal([]byte(buildMetadataJSON), &buildMetadata)
	if err != nil {
		return BuiltImage{}, err
	}

	err = json.Unmarshal([]byte(layerMetadataJSON), &layerMetadata)
	if err != nil {
		return BuiltImage{}, err
	}

	imageCreatedAt, err := img.CreatedAt()
	if err != nil {
		return BuiltImage{}, err
	}

	identifier, err := img.Identifier()
	if err != nil {
		return BuiltImage{}, err
	}

	runImageReferenceStr := layerMetadata.RunImage.Reference
	runImageRef, err := name.ParseReference(runImageReferenceStr)
	if err != nil {
		return BuiltImage{}, err
	}

	baseRunImage := layerMetadata.Stack.RunImage.Image
	baseImageRef, err := name.ParseReference(baseRunImage)
	if err != nil {
		return BuiltImage{}, err
	}

	return BuiltImage{
		Identifier:        identifier,
		CompletedAt:       imageCreatedAt,
		BuildpackMetadata: buildMetadata.Buildpacks,
		Stack: Stack{
			RunImage: baseImageRef.Context().String() + "@" + runImageRef.Identifier(),
			ID:       stackId,
		},
	}, nil
}
