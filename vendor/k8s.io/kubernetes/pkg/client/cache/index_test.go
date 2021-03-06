/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

func testIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*api.Pod)
	return []string{pod.Labels["foo"]}, nil
}

func TestGetIndexFuncValues(t *testing.T) {
	index := NewIndexer(MetaNamespaceKeyFunc, Indexers{"testmodes": testIndexFunc})

	pod1 := &api.Pod{ObjectMeta: api.ObjectMeta{Name: "one", Labels: map[string]string{"foo": "bar"}}}
	pod2 := &api.Pod{ObjectMeta: api.ObjectMeta{Name: "two", Labels: map[string]string{"foo": "bar"}}}
	pod3 := &api.Pod{ObjectMeta: api.ObjectMeta{Name: "tre", Labels: map[string]string{"foo": "biz"}}}

	index.Add(pod1)
	index.Add(pod2)
	index.Add(pod3)

	keys := index.ListIndexFuncValues("testmodes")
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys but got %v", len(keys))
	}

	for _, key := range keys {
		if key != "bar" && key != "biz" {
			t.Errorf("Expected only 'bar' or 'biz' but got %s", key)
		}
	}
}

func testUsersIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*api.Pod)
	usersString := pod.Annotations["users"]

	return strings.Split(usersString, ","), nil
}

func TestMultiIndexKeys(t *testing.T) {
	index := NewIndexer(MetaNamespaceKeyFunc, Indexers{"byUser": testUsersIndexFunc})

	pod1 := &api.Pod{ObjectMeta: api.ObjectMeta{Name: "one", Annotations: map[string]string{"users": "ernie,bert"}}}
	pod2 := &api.Pod{ObjectMeta: api.ObjectMeta{Name: "two", Annotations: map[string]string{"users": "bert,oscar"}}}
	pod3 := &api.Pod{ObjectMeta: api.ObjectMeta{Name: "tre", Annotations: map[string]string{"users": "ernie,elmo"}}}

	index.Add(pod1)
	index.Add(pod2)
	index.Add(pod3)

	erniePods, err := index.ByIndex("byUser", "ernie")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(erniePods) != 2 {
		t.Errorf("Expected 2 pods but got %v", len(erniePods))
	}

	bertPods, err := index.ByIndex("byUser", "bert")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(bertPods) != 2 {
		t.Errorf("Expected 2 pods but got %v", len(bertPods))
	}

	oscarPods, err := index.ByIndex("byUser", "oscar")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(oscarPods) != 1 {
		t.Errorf("Expected 1 pods but got %v", len(erniePods))
	}

	ernieAndBertKeys, err := index.Index("byUser", pod1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ernieAndBertKeys) != 3 {
		t.Errorf("Expected 3 pods but got %v", len(ernieAndBertKeys))
	}

	index.Delete(pod3)
	erniePods, err = index.ByIndex("byUser", "ernie")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(erniePods) != 1 {
		t.Errorf("Expected 1 pods but got %v", len(erniePods))
	}
	elmoPods, err := index.ByIndex("byUser", "elmo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(elmoPods) != 0 {
		t.Errorf("Expected 0 pods but got %v", len(elmoPods))
	}

	obj, err := api.Scheme.DeepCopy(pod2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	copyOfPod2 := obj.(*api.Pod)
	copyOfPod2.Annotations["users"] = "oscar"
	index.Update(copyOfPod2)
	bertPods, err = index.ByIndex("byUser", "bert")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(bertPods) != 1 {
		t.Errorf("Expected 1 pods but got %v", len(bertPods))
	}

}
