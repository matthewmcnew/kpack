package testhelpers

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"io"
)

type FakeLayer struct {
}

func (f FakeLayer) Digest() (v1.Hash, error) {

	return f.digest, nil
}

func (f FakeLayer) DiffID() (v1.Hash, error) {
	return f.diffid, nil
}

func (f FakeLayer) Compressed() (io.ReadCloser, error) {
	panic("Not implemented")
}

func (f FakeLayer) Uncompressed() (io.ReadCloser, error) {
	panic("Not implemented")
}

func (f FakeLayer) Size() (int64, error) {
}

func (f FakeLayer) MediaType() (types.MediaType, error) {
	return types.DockerLayer, nil
}
