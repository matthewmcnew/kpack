package cnb

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"sort"
	"time"
)

func NewRemoteBuilderImage(baseImage v1.Image) (*RemoteBuilderImage, error) {
	baseMetadata := &BuilderImageMetadata{}
	err := getLabel(baseImage, buildpackMetadataLabel, baseMetadata)
	if err != nil {
		return nil, err
	}

	return &RemoteBuilderImage{
		baseMetadata:    baseMetadata,
		baseImage:       baseImage,
		buildpackLayers: map[BuildpackInfo]buildpackLayer{},
	}, nil
}

type RemoteBuilderImage struct {
	baseImage       v1.Image
	baseMetadata    *BuilderImageMetadata
	order           []OrderEntry
	buildpackLayers map[BuildpackInfo]buildpackLayer
}

func (rb *RemoteBuilderImage) AddGroup(buildpacks ...RemoteBuildpack) {
	group := make([]BuildpackRef, 0, len(buildpacks))
	for _, b := range buildpacks {
		group = append(group, b.BuildpackRef)

		for _, layer := range b.Layers {
			rb.buildpackLayers[layer.BuildpackInfo] = layer
		}
	}
	rb.order = append(rb.order, OrderEntry{Group: group})
}

func (rb *RemoteBuilderImage) WriteableImage() (v1.Image, error) {
	buildpackLayerMetadata := make(BuildpackLayerMetadata)
	buildpacks := make([]BuildpackInfo, 0, len(rb.buildpackLayers))
	layers := make([]v1.Layer, 0, len(rb.buildpackLayers)+1)

	sortedBuildpacks, err := deterministicSortBySize(rb.buildpackLayers)
	if err != nil {
		return nil, err
	}

	for _, key := range sortedBuildpacks {
		layer := rb.buildpackLayers[key]

		if err := buildpackLayerMetadata.add(layer); err != nil {
			return nil, err
		}

		size, _ := layer.v1Layer.Size()
		var i float64 = float64(size / 1000000)
		fmt.Printf("adding: %s size: %f \n", key, i)

		buildpacks = append(buildpacks, key)
		layers = append(layers, layer.v1Layer)
	}

	orderLayer, err := rb.tomlLayer()
	if err != nil {
		return nil, err
	}

	image, err := mutate.AppendLayers(rb.baseImage, append(layers, orderLayer)...)
	if err != nil {
		return nil, err
	}

	return setLabels(image, map[string]interface{}{
		buildpackOrderLabel:  rb.order,
		buildpackLayersLabel: buildpackLayerMetadata,
		buildpackMetadataLabel: BuilderImageMetadata{
			Description: "Custom Builder built with kpack",
			Stack:       rb.baseMetadata.Stack,
			Lifecycle:   rb.baseMetadata.Lifecycle,
			CreatedBy: CreatorMetadata{
				Name:    "kpack CustomBuilder",
				Version: "",
			},
			Buildpacks: buildpacks,
		},
	})
}

func (rb *RemoteBuilderImage) tomlLayer() (v1.Layer, error) {
	orderBuf := &bytes.Buffer{}
	err := toml.NewEncoder(orderBuf).Encode(TomlOrder{rb.order})
	if err != nil {
		return nil, err
	}

	return singeFileLayer(orderTomlPath, orderBuf.Bytes())
}

func singeFileLayer(file string, contents []byte) (v1.Layer, error) {
	b := &bytes.Buffer{}
	w := tar.NewWriter(b)

	if err := w.WriteHeader(&tar.Header{
		Name:    file,
		Size:    int64(len(contents)),
		Mode:    0644,
		ModTime: time.Time{},
	}); err != nil {
		return nil, err
	}
	if _, err := w.Write(contents); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return tarball.LayerFromReader(b)
}

func setLabels(image v1.Image, labels map[string]interface{}) (v1.Image, error) {
	configFile, err := image.ConfigFile()
	if err != nil {
		return nil, err
	}
	config := *configFile.Config.DeepCopy()
	if config.Labels == nil {
		config.Labels = map[string]string{}
	}
	for k, v := range labels {
		dataBytes, err := json.Marshal(v)
		if err != nil {
			return nil, errors.Wrapf(err, "marshalling data to JSON for label %s", k)
		}

		config.Labels[k] = string(dataBytes)
	}
	return mutate.Config(image, config)
}

func getLabel(image v1.Image, key string, value interface{}) error {
	configFile, err := image.ConfigFile()
	if err != nil {
		return err
	}
	config := configFile.Config.DeepCopy()

	stringValue, ok := config.Labels[key]
	if !ok {
		return errors.Errorf("could not find label %s", key)
	}

	return json.Unmarshal([]byte(stringValue), value)
}

func deterministicSortBySize(layers map[BuildpackInfo]buildpackLayer) ([]BuildpackInfo, error) {
	keys := make([]BuildpackInfo, 0, len(layers))
	sizes := make(map[BuildpackInfo]int64, len(layers))
	for k, layer := range layers {
		keys = append(keys, k)
		size, err := layer.v1Layer.Size()
		if err != nil {
			return nil, err
		}

		sizes[k] = size
	}

	sort.Slice(keys, func(i, j int) bool {
		sizeI := sizes[keys[i]]
		sizeJ := sizes[keys[j]]

		if sizeI == sizeJ {
			return keys[i].String() > keys[j].String()
		}

		return sizeI > sizeJ
	})
	return keys, nil
}
