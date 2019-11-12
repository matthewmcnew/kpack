package cnb

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	eV1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry"
	"github.com/pivotal/kpack/pkg/registry/registryfakes"
	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCreateBuilder(t *testing.T) {
	spec.Run(t, "Create Builder", testCreateBuilder)
}

func testCreateBuilder(t *testing.T, when spec.G, it spec.S) {
	const (
		tag         = "custom/example"
		storeImage  = "store/image"
		baseBuilder = "base/builder"
	)

	var (
		fakeClient       = registryfakes.NewFakeClient()
		fakeStore        = &fakeStore{buildpacks: map[string][]buildpackLayer{}}
		expectedKeychain = authn.NewMultiKeychain(authn.DefaultKeychain)
	)

	fakeClient.ExpectedKeychain(expectedKeychain)
	remoteBuilderCreator := RemoteBuilderCreator{
		RemoteImageClient: fakeClient,
		NewStore: func(keychain authn.Keychain, image string) (Store, error) {
			if keychain != expectedKeychain {
				return nil, errors.New("invalid keychain")
			}
			if image != storeImage {
				return nil, errors.New("invalid store image")
			}

			return fakeStore, nil
		},
	}

	clusterBuilder := &eV1alpha1.CustomBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-name",
		},
		Spec: eV1alpha1.CustomBuilderSpec{
			Tag: "custom/example",
			Stack: eV1alpha1.Stack{
				BaseBuilderImage: baseBuilder,
			},
			Store: eV1alpha1.Store{
				Image: storeImage,
			},
			Order: []eV1alpha1.Group{
				{
					Group: []eV1alpha1.Buildpack{
						{
							ID:      "io.buildpack.1",
							Version: "v1",
						},
						{
							ID:      "io.buildpack.2",
							Version: "v2",
						},
					},
				},
			},
		},
	}

	var (
		buildpack1Layer v1.Layer
		buildpack2Layer v1.Layer
		buildpack3Layer v1.Layer
	)

	it.Before(func() {
		var err error
		buildpack1Layer, err = crane.Layer(map[string][]byte{
			"/cnb/io.buildpack.1/v1/bp":        []byte("io.buildpack.1"),
			"/cnb/io.buildpack.1/v1/layerSize": []byte("sm"),
		})
		require.NoError(t, err)

		fakeStore.AddBP("io.buildpack.1", "v1", []buildpackLayer{
			{
				v1Layer: buildpack1Layer,
				BuildpackInfo: BuildpackInfo{
					ID:      "io.buildpack.1",
					Version: "v1",
				},
			},
		})

		buildpack2Layer, err = crane.Layer(map[string][]byte{
			"/cnb/io.buildpack.2/v1/bp":        []byte("io.buildpack.2"),
			"/cnb/io.buildpack.2/v1/layerSize": []byte("medium"),
		})
		require.NoError(t, err)

		buildpack3Layer, err = crane.Layer(map[string][]byte{
			"/cnb/io.buildpack.3/v1/bp":        []byte("io.buildpack.3"),
			"/cnb/io.buildpack.3/v1/layerSize": []byte("THE-Largest"),
		})
		require.NoError(t, err)

		fakeStore.AddBP("io.buildpack.2", "v2", []buildpackLayer{
			{
				v1Layer: buildpack3Layer,
				BuildpackInfo: BuildpackInfo{
					ID:      "io.buildpack.3",
					Version: "v2",
				},
			},
			{
				v1Layer: buildpack2Layer,
				BuildpackInfo: BuildpackInfo{
					ID:      "io.buildpack.2",
					Version: "v1",
				},
				Order: Order{
					{
						Group: []BuildpackRef{
							{
								BuildpackInfo: BuildpackInfo{
									ID:      "io.buildpack.3",
									Version: "v2",
								},
								Optional: false,
							},
						},
					},
				},
			},
		})
	})

	when("CreateBuilder", func() {
		var baseImage v1.Image
		it.Before(func() {
			var err error
			baseImage, err = random.Image(10, 10)
			require.NoError(t, err)

			baseImage, err := registry.SetStringLabel(baseImage, map[string]string{
				stackMetadataLabel: "io.buildpacks.stack",
			})
			require.NoError(t, err)

			baseImage, err = registry.SetLabels(baseImage, map[string]interface{}{
				buildpackMetadataLabel: BuilderImageMetadata{
					Stack: StackMetadata{
						RunImage: RunImageMetadata{
							Image: "kpack/run",
						},
					},
					Lifecycle: LifecycleMetadata{
						LifecycleInfo: LifecycleInfo{
							Version: "0.5.0",
						},
						API: LifecycleAPI{
							BuildpackVersion: "0.2",
							PlatformVersion:  "0.1",
						},
					},
				},
			})
			require.NoError(t, err)

			fakeClient.AddImage("base/builder", baseImage)
		})

		it("creates a with build layers from the store", func() {
			builderRecord, err := remoteBuilderCreator.CreateBuilder(expectedKeychain, clusterBuilder)
			require.NoError(t, err)

			assert.Len(t, builderRecord.Buildpacks, 3)
			assert.Contains(t, builderRecord.Buildpacks, v1alpha1.BuildpackMetadata{ID: "io.buildpack.1", Version: "v1"})
			assert.Contains(t, builderRecord.Buildpacks, v1alpha1.BuildpackMetadata{ID: "io.buildpack.2", Version: "v1"})
			assert.Contains(t, builderRecord.Buildpacks, v1alpha1.BuildpackMetadata{ID: "io.buildpack.3", Version: "v2"})
			assert.Equal(t, v1alpha1.BuildStack{RunImage: "kpack/run", ID: "io.buildpacks.stack"}, builderRecord.Stack)

			assert.Len(t, fakeClient.SavedImages(), 1)
			savedImage := fakeClient.SavedImages()[tag]

			hash, err := savedImage.Digest()
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("%s@%s", tag, hash), builderRecord.Image)

			layers, err := savedImage.Layers()
			require.NoError(t, err)
			assert.Contains(t, layers, buildpack1Layer)
			assert.Contains(t, layers, buildpack2Layer)
			assert.Contains(t, layers, buildpack3Layer)

			buildpackOrder, err := registry.GetStringLabel(savedImage, buildpackOrderLabel)
			assert.NoError(t, err)
			assert.JSONEq(t, `[{"group":[{"id":"io.buildpack.1","version":"v1"},{"id":"io.buildpack.2","version":"v2"}]}]`, buildpackOrder)

			buildpackMetadata, err := registry.GetStringLabel(savedImage, buildpackMetadataLabel)
			assert.NoError(t, err)
			assert.JSONEq(t, `{
  "description": "Custom Builder built with kpack",
  "stack": {
    "runImage": {
      "image": "kpack/run",
      "mirrors": null
    }
  },
  "lifecycle": {
    "version": "0.5.0",
    "api": {
      "buildpack": "0.2",
      "platform": "0.1"
    }
  },
  "createdBy": {
    "name": "kpack CustomBuilder",
    "version": ""
  },
  "buildpacks": [
    {
      "id": "io.buildpack.3",
      "version": "v2"
    },
    {
      "id": "io.buildpack.2",
      "version": "v1"
    },
    {
      "id": "io.buildpack.1",
      "version": "v1"
    }
  ]
}`, buildpackMetadata)

			buildpackLayers, err := registry.GetStringLabel(savedImage, buildpackLayersLabel)
			assert.NoError(t, err)
			assert.JSONEq(t, `{
  "io.buildpack.1": {
    "v1": {
      "layerDigest": "sha256:05e924557699a861c0f009f64b9bbcaa1912374bdded6e20d2eaedba32d56a5c",
      "layerDiffID": "sha256:0aca161a19383e3ad7c3d21cd650dd982d5b6f4a3f1c26b37c059e5c884692b1"
    }
  },
  "io.buildpack.2": {
    "v1": {
      "layerDigest": "sha256:e0fa2148e4506e3d9910aa813cfdef6d088863565e30e9944e43ceb431fae774",
      "layerDiffID": "sha256:3859a85bbe16b42427048313b98822df666112ab4cd123a5775f104acba2d45c",
      "order": [
        {
          "group": [
            {
              "id": "io.buildpack.3",
              "version": "v2"
            }
          ]
        }
      ]
    }
  },
  "io.buildpack.3": {
    "v2": {
      "layerDigest": "sha256:05781c7ea28573ed96d88bc12685c6fb76ba0af19866499f4202658cbd090df7",
      "layerDiffID": "sha256:e5f3087b7180a8ef3f6c0b2d55590011279991e38be88daa041a0a53806ee9c3"
    }
  }
}
`, buildpackLayers)

		})

		it("creates images deterministically ", func() {
			original, err := remoteBuilderCreator.CreateBuilder(expectedKeychain, clusterBuilder)
			require.NoError(t, err)

			for i := 1; i <= 50; i++ {
				other, err := remoteBuilderCreator.CreateBuilder(expectedKeychain, clusterBuilder)
				require.NoError(t, err)

				require.Equal(t, original.Image, other.Image)
				require.Equal(t, original.Buildpacks, other.Buildpacks)
			}
		})
	})
}

type fakeStore struct {
	buildpacks map[string][]buildpackLayer
}

func (f *fakeStore) FetchBuildpack(id, version string) (RemoteBuildpackInfo, error) {
	layers, ok := f.buildpacks[fmt.Sprintf("%s@%s", id, version)]
	if !ok {
		return RemoteBuildpackInfo{}, errors.New("buildpack not found")
	}

	return RemoteBuildpackInfo{
		BuildpackInfo: BuildpackInfo{
			ID:      id,
			Version: version,
		},
		Layers: layers,
	}, nil
}

func (f *fakeStore) AddBP(id, version string, layers []buildpackLayer) {
	f.buildpacks[fmt.Sprintf("%s@%s", id, version)] = layers
}
