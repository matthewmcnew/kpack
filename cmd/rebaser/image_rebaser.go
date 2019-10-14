package main

import (
	"context"
	"encoding/json"
	"github.com/buildpack/imgutil/remote"
	"github.com/google/go-containerregistry/pkg/authn"
	"io"
	"time"

	"github.com/buildpack/imgutil"
	"github.com/buildpack/lifecycle"
	lcyclemd "github.com/buildpack/lifecycle/metadata"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

func Rebase(keychain authn.Keychain, builder, previousImage string, tags []string) error {
	builderImage, err := newRemote(builder, builder, keychain)
	if err != nil {
		return err
	}

	appImage, err := newRemote(tags[0], previousImage, keychain)
	if err != nil {
		return err
	}

	metadataJSON, err := builderImage.Label("io.buildpacks.builder.metadata")
	if err != nil {
		return err
	}

	var metadata BuilderImageMetadata
	err = json.Unmarshal([]byte(metadataJSON), &metadata)
	if err != nil {
		return err
	}

	newBaseImage, err := newRemote(metadata.Stack.RunImage.Image, metadata.Stack.RunImage.Image, keychain)
	if err != nil {
		return err
	}

	rebaser := lifecycle.Rebaser{
		Logger: wrappedLogger{logging.FromContext(context.TODO())},
	}
	err = rebaser.Rebase(appImage, newBaseImage, tags[1:])
	if err != nil {
		return err
	}

	return nil
}

type remoteImageWrapper struct {
	remoteImgUtilImage imgutil.Image
}

func (r remoteImageWrapper) CreatedAt() (time.Time, error) {
	return r.remoteImgUtilImage.CreatedAt()
}

func (r remoteImageWrapper) Identifier() (string, error) {
	i, err := r.remoteImgUtilImage.Identifier()
	if err != nil {
		return "", err
	}
	return i.String(), nil
}

func (r remoteImageWrapper) Label(labelName string) (string, error) {
	return r.remoteImgUtilImage.Label(labelName)
}

func (r remoteImageWrapper) Env(key string) (string, error) {
	return r.remoteImgUtilImage.Env(key)
}

type wrappedLogger struct {
	logger *zap.SugaredLogger
}

func (w wrappedLogger) Debug(msg string) {
	w.logger.Debug(msg)
}

func (w wrappedLogger) Debugf(fmt string, v ...interface{}) {
	w.logger.Debugf(fmt, v...)
}

func (w wrappedLogger) Info(msg string) {
	w.logger.Info(msg)
}

func (w wrappedLogger) Infof(fmt string, v ...interface{}) {
	w.logger.Infof(fmt, v...)
}

func (w wrappedLogger) Warn(msg string) {
	w.logger.Warn(msg)
}

func (w wrappedLogger) Warnf(fmt string, v ...interface{}) {
	w.logger.Warnf(fmt, v...)
}

func (w wrappedLogger) Error(msg string) {
	w.logger.Error(msg)
}

func (w wrappedLogger) Errorf(fmt string, v ...interface{}) {
	w.logger.Errorf(fmt, v...)
}

func (w wrappedLogger) Writer() io.Writer {
	panic("not implemented")
}

func (w wrappedLogger) WantLevel(level string) {
	panic("not implemented")
}

func newRemote(imageName string, baseImage string, keychain authn.Keychain) (imgutil.Image, error) {
	image, err := remote.NewImage(imageName, keychain, remote.FromBaseImage(baseImage))
	return image, errors.WithStack(err)
}

type BuildpackMetadata struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type BuilderImageMetadata struct {
	Buildpacks []BuildpackMetadata    `json:"buildpacks"`
	Stack      lcyclemd.StackMetadata `json:"stack"`
}

type BuilderImage struct {
	BuilderBuildpackMetadata BuilderMetadata
	RunImage                 string
	Identifier               string
}

type BuilderMetadata []BuildpackMetadata
