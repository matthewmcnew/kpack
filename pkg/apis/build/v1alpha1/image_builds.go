package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

const (
	BuildNumberLabel = "image.build.pivotal.io/buildNumber"
	ImageLabel       = "image.build.pivotal.io/image"

	BuildReasonAnnotation = "image.build.pivotal.io/reason"
	BuildReasonConfig     = "CONFIG"
	BuildReasonCommit     = "COMMIT"
	BuildReasonBuildpack  = "BUILDPACK"
	BuildReasonStack      = "STACK"
)

func (im *Image) buildNeeded(lastBuild *Build, sourceResolver *SourceResolver, builder BuilderResource) ([]string, bool, error) {
	if !sourceResolver.Ready() {
		return []string{}, false, nil
	}

	if !builder.Ready() {
		return []string{}, false, nil
	}

	if lastBuild == nil {
		return []string{BuildReasonConfig}, true, nil
	}

	if im.Spec.Tag != lastBuild.Tag() {
		return []string{BuildReasonConfig}, true, nil
	}

	var reasons []string

	if sourceResolver.ConfigChanged(lastBuild) ||
		!equality.Semantic.DeepEqual(im.Spec.Build.Env, lastBuild.Spec.Env) ||
		!equality.Semantic.DeepEqual(im.Spec.Build.Resources, lastBuild.Spec.Resources) {
		reasons = append(reasons, BuildReasonConfig)
	}

	if sourceResolver.RevisionChanged(lastBuild) {
		reasons = append(reasons, BuildReasonCommit)
	}

	if !lastBuildBuiltWithBuilderBuildpacks(builder, lastBuild) {
		reasons = append(reasons, BuildReasonBuildpack)
	}

	if lastBuild.Status.RunImage != "" {
		lastBuildRunImageRef, err := name.ParseReference(lastBuild.Status.RunImage)
		if err != nil {
			return reasons, false, err
		}

		builderRunImageRef, err := name.ParseReference(builder.RunImage())
		if err != nil {
			return reasons, false, err
		}

		if lastBuildRunImageRef.Identifier() != builderRunImageRef.Identifier() {
			reasons = append(reasons, BuildReasonStack)
		}
	}

	return reasons, len(reasons) > 0, nil
}

func lastBuildBuiltWithBuilderBuildpacks(builder BuilderResource, build *Build) bool {
	for _, bp := range build.Status.BuildMetadata {
		if !builder.BuildpackMetadata().Include(bp) {
			return false
		}
	}

	return true
}

func (im *Image) build(sourceResolver *SourceResolver, builder BuilderResource, reasons []string, nextBuildNumber int64) *Build {
	buildNumber := strconv.Itoa(int(nextBuildNumber))
	return &Build{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    im.Namespace,
			GenerateName: im.generateBuildName(buildNumber),
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(im),
			},
			Labels: im.labels(map[string]string{
				BuildNumberLabel: buildNumber,
				ImageLabel:       im.Name,
			}),
			Annotations: map[string]string{
				BuildReasonAnnotation: strings.Join(reasons, ","),
			},
		},
		Spec: BuildSpec{
			Tags:           im.generateTags(buildNumber),
			Builder:        builder.BuildBuilderSpec(),
			Env:            im.Spec.Build.Env,
			Resources:      im.Spec.Build.Resources,
			ServiceAccount: im.Spec.ServiceAccount,
			Source:         sourceResolver.SourceConfig(),
			CacheName:      im.Status.BuildCacheName,
			LastBuild:      LastBuild{Image: im.Status.LatestImage},
		},
	}
}

func (im *Image) latestForImage(build *Build) string {
	latestImage := im.Status.LatestImage
	if build.IsSuccess() {
		latestImage = build.BuiltImage()
	}
	return latestImage
}

func (im *Image) CacheName() string {
	return kmeta.ChildName(im.Name, "-cache")
}

func (im *Image) NeedCache() bool {
	return im.Spec.CacheSize != nil
}

func (im *Image) BuildCache() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      im.CacheName(),
			Namespace: im.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(im),
			},
			Labels: im.labels(nil),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *im.Spec.CacheSize,
				},
			},
		},
	}
}

func (im *Image) SourceResolverName() string {
	return kmeta.ChildName(im.Name, "-source")
}

func (im *Image) SourceResolver() *SourceResolver {
	return &SourceResolver{
		ObjectMeta: metav1.ObjectMeta{
			Name:      im.SourceResolverName(),
			Namespace: im.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(im),
			},
			Labels: im.labels(nil),
		},
		Spec: SourceResolverSpec{
			ServiceAccount: im.Spec.ServiceAccount,
			Source:         im.Spec.Source,
		},
	}
}

func (im *Image) generateTags(buildNumber string) []string {
	if im.disableAdditionalImageNames() {
		return []string{im.Spec.Tag}
	}
	now := time.Now()

	tag, err := name.NewTag(im.Spec.Tag, name.WeakValidation)
	if err != nil {
		// We assume that if the Image Name cannot be parsed the image will not be successfully built
		// in this case we can just ignore any additional image names
		return nil
	}

	tagName := tag.TagStr() + "-"
	if tagName == "latest-" {
		tagName = ""
	}
	return []string{
		im.Spec.Tag,
		tag.RegistryStr() + "/" + tag.RepositoryStr() + ":" + tagName + "b" + buildNumber + "." + now.Format("20060102") + "." + fmt.Sprintf("%02d%02d%02d", now.Hour(), now.Minute(), now.Second())}
}

func (im *Image) generateBuildName(buildNumber string) string {
	return im.Name + "-build-" + buildNumber + "-"
}

func (im *Image) labels(additionalLabels map[string]string) map[string]string {
	labels := make(map[string]string, len(additionalLabels)+len(im.Labels))

	for k, v := range im.Labels {
		labels[k] = v
	}
	for k, v := range additionalLabels {
		labels[k] = v
	}
	return labels
}

func (im *Image) disableAdditionalImageNames() bool {
	return im.Spec.ImageTaggingStrategy == None
}
