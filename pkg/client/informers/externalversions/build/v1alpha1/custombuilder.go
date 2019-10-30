/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	versioned "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	internalinterfaces "github.com/pivotal/kpack/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/pivotal/kpack/pkg/client/listers/build/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// CustomBuilderInformer provides access to a shared informer and lister for
// CustomBuilders.
type CustomBuilderInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.CustomBuilderLister
}

type customBuilderInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewCustomBuilderInformer constructs a new informer for CustomBuilder type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCustomBuilderInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCustomBuilderInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredCustomBuilderInformer constructs a new informer for CustomBuilder type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCustomBuilderInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BuildV1alpha1().CustomBuilders(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BuildV1alpha1().CustomBuilders(namespace).Watch(options)
			},
		},
		&buildv1alpha1.CustomBuilder{},
		resyncPeriod,
		indexers,
	)
}

func (f *customBuilderInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCustomBuilderInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *customBuilderInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&buildv1alpha1.CustomBuilder{}, f.defaultInformer)
}

func (f *customBuilderInformer) Lister() v1alpha1.CustomBuilderLister {
	return v1alpha1.NewCustomBuilderLister(f.Informer().GetIndexer())
}
