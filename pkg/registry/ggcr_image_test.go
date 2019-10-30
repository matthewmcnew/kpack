package registry

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestGGCRImage(t *testing.T) {
	spec.Run(t, "GGCR Image", testGGCRImage)
}

func testGGCRImage(t *testing.T, when spec.G, it spec.S) {
	when("#CreatedAt", func() {
		it("returns created at from the image", func() {
			image, err := NewGoContainerRegistryImage("cloudfoundry/cnb:bionic@sha256:33c3ad8676530f864d51d78483b510334ccc4f03368f7f5bb9d517ff4cbd630f", authn.DefaultKeychain)
			require.NoError(t, err)

			createdAt, err := image.CreatedAt()
			require.NoError(t, err)

			require.NotEqual(t, time.Time{}, createdAt)
		})
	})

	when("#Label", func() {
		it("returns created at from the image", func() {
			image, err := NewGoContainerRegistryImage("cloudfoundry/cnb:bionic@sha256:33c3ad8676530f864d51d78483b510334ccc4f03368f7f5bb9d517ff4cbd630f", authn.DefaultKeychain)
			require.NoError(t, err)

			metadata, err := image.Label("io.buildpacks.builder.metadata")
			require.NoError(t, err)

			require.NotEmpty(t, metadata)
		})
	})

	when("#Env", func() {
		it("returns created at from the image", func() {
			image, err := NewGoContainerRegistryImage("cloudfoundry/cnb:bionic@sha256:33c3ad8676530f864d51d78483b510334ccc4f03368f7f5bb9d517ff4cbd630f", authn.DefaultKeychain)
			require.NoError(t, err)

			cnbUserId, err := image.Env("CNB_USER_ID")
			require.NoError(t, err)

			require.NotEmpty(t, cnbUserId)
		})
	})

	when("#identifer", func() {
		it("includes digest if repoName does not have a digest", func() {
			image, err := NewGoContainerRegistryImage("cloudfoundry/cnb:bionic", authn.DefaultKeychain)
			require.NoError(t, err)

			identifier, err := image.Identifier()
			require.NoError(t, err)
			require.Len(t, identifier, 104)
			require.Equal(t, identifier[0:40], "index.docker.io/cloudfoundry/cnb@sha256:")
		})

		it("includes digest if repoName already has a digest", func() {
			image, err := NewGoContainerRegistryImage("cloudfoundry/cnb:bionic@sha256:33c3ad8676530f864d51d78483b510334ccc4f03368f7f5bb9d517ff4cbd630f", authn.DefaultKeychain)
			require.NoError(t, err)

			identifier, err := image.Identifier()
			require.NoError(t, err)
			require.Equal(t, identifier, "index.docker.io/cloudfoundry/cnb@sha256:33c3ad8676530f864d51d78483b510334ccc4f03368f7f5bb9d517ff4cbd630f")
		})
	})

	when("#AddLayer", func() {
		it("append layer to image", func() {
			image, err := NewGoContainerRegistryImage("cloudfoundry/cnb:bionic@sha256:33c3ad8676530f864d51d78483b510334ccc4f03368f7f5bb9d517ff4cbd630f", authn.DefaultKeychain)
			require.NoError(t, err)

			buf := bytes.NewBuffer([]byte("some layer"))

			layerToAdd, err := tarball.LayerFromReader(buf)

			changedImage, err := image.AddLayer(layerToAdd)
			require.NoError(t, err)

			ref, err := name.ParseReference("cloudfoundry/cnb:bionic@sha256:33c3ad8676530f864d51d78483b510334ccc4f03368f7f5bb9d517ff4cbd630f", name.WeakValidation)
			require.NoError(t, err)

			expectedImage, err := remote.Image(ref, remote.WithTransport(http.DefaultTransport))
			expectedImage, err = mutate.AppendLayers(expectedImage, layerToAdd)
			require.NoError(t, err)
			expectedDigest, err := expectedImage.Digest()
			require.NoError(t, err)
			digest, err := changedImage.Digest()
			require.NoError(t, err)

			assert.Equal(t, expectedDigest.String(), digest)
		})
	})
}
