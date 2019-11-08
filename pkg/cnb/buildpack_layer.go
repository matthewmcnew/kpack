package cnb

import v1 "github.com/google/go-containerregistry/pkg/v1"

type Store interface {
	FetchBuildpack(id, version string) (RemoteBuildpackInfo, error)
}

type RemoteBuildpack struct {
	BuildpackRef BuildpackRef
	Layers       []buildpackLayer
}

type RemoteBuildpackInfo struct {
	BuildpackInfo BuildpackInfo
	Layers        []buildpackLayer
}

func (i RemoteBuildpackInfo) Optional(optional bool) RemoteBuildpack {
	return RemoteBuildpack{
		BuildpackRef: BuildpackRef{
			BuildpackInfo: i.BuildpackInfo,
			Optional:      optional,
		},
		Layers: i.Layers,
	}
}

type buildpackLayer struct {
	v1Layer v1.Layer
	BuildpackInfo BuildpackInfo
	Order   Order
}
