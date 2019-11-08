package cnb

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	v12 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

func TestCreate(t *testing.T) {
	logs.Progress.SetOutput(os.Stdout)

	buildPackageStoreFactory := &BuildPackageStoreFactory{
		KeychainFactory: DefaultKeyChainFactory{},
	}

	creator := &RemoteBuilderCreator{
		FetchImage: func(secretRef registry.SecretRef, image string) (v12.Image, error) {
			ref, err := name.ParseReference(image)
			if err != nil {
				return nil, err
			}

			return remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
		},
		NewStore: buildPackageStoreFactory.NewBuildPackageStore,
	}

	builderImage, err := creator.CreateBuilder(&v1alpha1.CustomBuilder{
		ObjectMeta: v1.ObjectMeta{},
		Spec: v1alpha1.CustomBuilderSpec{
			Tag: "registry.default.svc.cluster.local:5000/builder",
			Stack: v1alpha1.Stack{
				BaseBuilderImage: "registry.default.svc.cluster.local:5000/base",
			},
			Store: v1alpha1.Store{
				Image: "registry.default.svc.cluster.local:5000/store",
			},
			Order: []v1alpha1.Group{
				{
					Group: []v1alpha1.Buildpack{
						{
							ID: "org.cloudfoundry.nodejs",
						},
					},
				},
				//{
				//	Group: []v1alpha1.Buildpack{
				//		//{
				//		//	ID: "org.bogus",
				//		//},
				//
				//		{
				//			ID: "org.cloudfoundry.npm",
				//		},
				//	},
				//},
				{
					Group: []v1alpha1.Buildpack{
						{
							ID: "org.cloudfoundry.openjdk",
						},
						{
							ID: "org.cloudfoundry.buildsystem",
							Optional: true,
						},
						{
							ID: "org.cloudfoundry.jvmapplication",
						},
						{
							ID: "org.cloudfoundry.tomcat",
							Optional: true,
						},
						{
							ID: "org.cloudfoundry.springboot",
							Optional: true,
						},
						//{
						//	ID: "org.cloudfoundry.procfile",
						//	Optional: true,
						//},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	ggcrImage, err := builderImage.WriteableImage()
	require.NoError(t, err)

	//expectedDigest, err := ggcrImage.Digest()
	//require.NoError(t, err)
	//
	//for i := 1; i <= 1000; i++ {
	//	newBuilderimage, err := creator.CreateBuilder(&v1alpha1.CustomBuilder{
	//		ObjectMeta: v1.ObjectMeta{},
	//		Spec: v1alpha1.CustomBuilderSpec{
	//			Tag: "registry.default.svc.cluster.local:5000/builder",
	//			Stack: v1alpha1.Stack{
	//				BaseBuilderImage: "registry.default.svc.cluster.local:5000/base",
	//			},
	//			Store: v1alpha1.Store{
	//				Image: "registry.default.svc.cluster.local:5000/store",
	//			},
	//			Order: []v1alpha1.Group{
	//				{
	//					Group: []v1alpha1.Buildpack{
	//						{
	//							ID: "org.cloudfoundry.nodejs",
	//						},
	//					},
	//				},
	//				{
	//					Group: []v1alpha1.Buildpack{
	//						{
	//							ID: "org.cloudfoundry.npm",
	//						},
	//					},
	//				},
	//			},
	//		},
	//	})
	//	require.NoError(t, err)
	//
	//	image1, err := newBuilderimage.WriteableImage()
	//	require.NoError(t, err)
	//	digest, err := image1.Digest()
	//
	//	if expectedDigest != digest {
	//		tag, err := name.ParseReference("registry.default.svc.cluster.local:5000/builder", name.WeakValidation)
	//		require.NoError(t, err)
	//
	//		err = remote.Write(tag, ggcrImage)
	//		require.NoError(t, err)
	//
	//		err = remote.Write(tag, image1)
	//		require.NoError(t, err)
	//
	//		require.Equal(t, expectedDigest, digest, fmt.Sprintf("failure on %d", i))
	//	}
	//
	//}

	tag, err := name.ParseReference("registry.default.svc.cluster.local:5000/builder", name.WeakValidation)
	require.NoError(t, err)

	err = remote.Write(tag, ggcrImage)
	require.NoError(t, err)
}

//func TestWhat(t *testing.T) {
//
//	logs.Progress.SetOutput(os.Stdout)
//	logs.Progress.SetOutput(os.Stdout)
//
//	ref, err := name.ParseReference("registry.default.svc.cluster.local:5000/store", name.WeakValidation)
//	require.NoError(t, err)
//
//	_, err = remote.Image(ref)
//	require.NoError(t, err)
//
//}

type DefaultKeyChainFactory struct {
}

func (d DefaultKeyChainFactory) KeychainForSecretRef(registry.SecretRef) (authn.Keychain, error) {
	return authn.DefaultKeychain, nil
}
