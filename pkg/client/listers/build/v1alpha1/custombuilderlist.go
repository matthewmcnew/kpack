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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CustomBuilderListLister helps list CustomBuilderLists.
type CustomBuilderListLister interface {
	// List lists all CustomBuilderLists in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.CustomBuilderList, err error)
	// CustomBuilderLists returns an object that can list and get CustomBuilderLists.
	CustomBuilderLists(namespace string) CustomBuilderListNamespaceLister
	CustomBuilderListListerExpansion
}

// customBuilderListLister implements the CustomBuilderListLister interface.
type customBuilderListLister struct {
	indexer cache.Indexer
}

// NewCustomBuilderListLister returns a new CustomBuilderListLister.
func NewCustomBuilderListLister(indexer cache.Indexer) CustomBuilderListLister {
	return &customBuilderListLister{indexer: indexer}
}

// List lists all CustomBuilderLists in the indexer.
func (s *customBuilderListLister) List(selector labels.Selector) (ret []*v1alpha1.CustomBuilderList, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CustomBuilderList))
	})
	return ret, err
}

// CustomBuilderLists returns an object that can list and get CustomBuilderLists.
func (s *customBuilderListLister) CustomBuilderLists(namespace string) CustomBuilderListNamespaceLister {
	return customBuilderListNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CustomBuilderListNamespaceLister helps list and get CustomBuilderLists.
type CustomBuilderListNamespaceLister interface {
	// List lists all CustomBuilderLists in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.CustomBuilderList, err error)
	// Get retrieves the CustomBuilderList from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.CustomBuilderList, error)
	CustomBuilderListNamespaceListerExpansion
}

// customBuilderListNamespaceLister implements the CustomBuilderListNamespaceLister
// interface.
type customBuilderListNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CustomBuilderLists in the indexer for a given namespace.
func (s customBuilderListNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.CustomBuilderList, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CustomBuilderList))
	})
	return ret, err
}

// Get retrieves the CustomBuilderList from the indexer for a given namespace and name.
func (s customBuilderListNamespaceLister) Get(name string) (*v1alpha1.CustomBuilderList, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("custombuilderlist"), name)
	}
	return obj.(*v1alpha1.CustomBuilderList), nil
}
