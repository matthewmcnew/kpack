package main

import (
	"context"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/kpack/pkg/dockercreds/k8sdockercreds"
	"github.com/pivotal/kpack/pkg/registry"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

func main() {
	repositoryRef := os.Args[1]
	log.Printf("Attempting to read %s", repositoryRef)

	logs.Debug.SetOutput(os.Stdout)

	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	config, err := k8s.NewForConfig(restConfig)
	if err != nil {
		log.Fatal(err)
	}

	factory, err := k8sdockercreds.NewSecretKeychainFactory(config)
	if err != nil {
		log.Fatal(err)
	}

	keychain, err := factory.KeychainForSecretRef(context.Background(), registry.SecretRef{
		Namespace: "kpack",
		ImagePullSecrets: []v1.LocalObjectReference{{
			Name: "canonical-registry-secret",
		}},
	})
	if err != nil {
		log.Fatal(err)
	}

	compareKeyChain(repositoryRef, keychain)

	reference, err := name.ParseReference(repositoryRef)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("reading with k8s keychain")
	readReference(keychain, reference)

	log.Printf("reading with default keychain")
	readReference(authn.DefaultKeychain, reference)


	log.Printf("writing with k8s keychain")
	write(keychain, reference)

	log.Printf("writing with default keychain")
	write(authn.DefaultKeychain, reference)
}

func readReference(keychain authn.Keychain, reference name.Reference) {
	image, err := remote.Image(reference, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		log.Printf("error reading image %s", err.Error())
		return
	}

	manifest, err := image.RawManifest()
	if err != nil {
		log.Printf("error reading manifest %s", err.Error())
		return
	}

	log.Printf(string(manifest))
}

func write(keychain authn.Keychain, reference name.Reference) {
	tag := reference.Context().Tag("deleteable")

	image, err := random.Image(1, 1)
	if err != nil {
		log.Printf("error generating random image %s", err.Error())
		return
	}

	err = remote.Write(tag, image, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		log.Printf("error writing image %s", err.Error())
		return
	}

}


func compareKeyChain(repositoryRef string, keychain authn.Keychain) {
	reference, err := name.ParseReference(repositoryRef)
	if err != nil {
		log.Fatal(err)
	}

	resolve, err := keychain.Resolve(reference.Context().Registry)
	if err != nil {
		log.Fatal(errors.Wrap(err, "keychain resolve"))
	}

	authorization, err := resolve.Authorization()
	if err != nil {
		log.Fatal(errors.Wrap(err, "authorization on keychain"))
	}
	log.Printf("Resolved Authorization on k8s keychain:")
	log.Printf("%v", authorization)

	resolve, err = authn.DefaultKeychain.Resolve(reference.Context().Registry)
	if err != nil {
		log.Fatal(errors.Wrap(err, "default keychain resolve"))
	}

	defaultAuth, err := resolve.Authorization()
	if err != nil {
		log.Fatal(errors.Wrap(err, "authorization on default keychain"))
	}
	log.Printf("Resolved Authorization on default keychain:")
	log.Printf("%v", defaultAuth)


	if authorization.Username != defaultAuth.Username {
		log.Printf("usernames %s do not match %s", authorization.Username, defaultAuth.Username)
	} else {
		log.Printf("usernames do match")
	}

	if authorization.Password != defaultAuth.Password {
		log.Printf("Passwords %s do not match %s", authorization.Password, defaultAuth.Password)
	} else {
		log.Printf("Passwords do match")
	}

	if authorization.Auth != defaultAuth.Auth {
		log.Printf("Auths %s do not match %s", authorization.Auth, defaultAuth.Auth)
	} else {
		log.Printf("Auths do match")
	}

}