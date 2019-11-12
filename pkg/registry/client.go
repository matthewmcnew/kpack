package registry

import (
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Client interface {
	Fetch(keychain authn.Keychain, repoName string) (v1.Image, error)
	Save(keychain authn.Keychain, tag string, image v1.Image) (string, error)
}
