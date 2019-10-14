package main

import (
	"flag"
	"github.com/pivotal/kpack/pkg/dockercreds"
	"log"
	"os"
)

var (
	builder       = flag.String("builder", os.Getenv("BUILDER"), "builder")
	previousImage = flag.String("previousImage", os.Getenv("PREVIOUS_IMAGE"), "previous image")

	gitCredentials    credentialsFlags
	dockerCredentials credentialsFlags
)

func init() {
	flag.Var(&gitCredentials, "basic-git", "Basic authentication for git on the form 'secretname=git.domain.com'")
	flag.Var(&dockerCredentials, "basic-docker", "Basic authentication for docker on form 'secretname=git.domain.com'")
}

const (
	buildSecretsDir = "/var/build-secrets"
)

func main() {
	flag.Parse()

	creds, err := dockercreds.ParseMountedAnnotatedSecrets(buildSecretsDir, dockerCredentials)
	if err != nil {
		log.Fatal(err)
	}

	tags := flag.Args()

	err = Rebase(creds, *builder, *previousImage, tags)
	if err != nil {
		log.Fatal(err)
	}

}
