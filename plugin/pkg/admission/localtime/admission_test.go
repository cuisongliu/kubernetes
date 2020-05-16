/*
Copyright 2015 The Kubernetes Authors.

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

package localtime

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// TestAdmission verifies all create requests for pods result in every container's image pull policy
// set to Always
func TestAdmission(t *testing.T) {
	namespace := "test"
	handler := &Localtime{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "123", Namespace: namespace, Annotations: map[string]string{"kubernetes.io/localtime": "true"}},
		Spec: api.PodSpec{
			InitContainers: []api.Container{
				{Name: "init1", Image: "image"},
			},
			Containers: []api.Container{
				{Name: "ctr1", Image: "image"},
				{Name: "ctr2", Image: "image"},
			},
		},
	}
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, false, nil), nil)
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	for _, c := range pod.Spec.InitContainers {
		vMount := c.VolumeMounts[len(c.VolumeMounts)-1]
		if vMount.ReadOnly != true {
			t.Errorf("InitContainer %v: expected VolumeMounts, got %v", c, vMount.ReadOnly)
		}
		if vMount.MountPath != localtimePath {
			t.Errorf("InitContainer %v: expected VolumeMounts, got %v", c, vMount.MountPath)
		}
		if vMount.Name != localtimeName {
			t.Errorf("InitContainer %v: expected VolumeMounts, got %v", c, vMount.Name)
		}
	}
	for _, c := range pod.Spec.Containers {
		vMount := c.VolumeMounts[len(c.VolumeMounts)-1]
		if vMount.ReadOnly != true {
			t.Errorf("Container %v: expected VolumeMounts, got %v", c, vMount.ReadOnly)
		}
		if vMount.MountPath != localtimePath {
			t.Errorf("Container %v: expected VolumeMounts, got %v", c, vMount.MountPath)
		}
		if vMount.Name != localtimeName {
			t.Errorf("Container %v: expected VolumeMounts, got %v", c, vMount.Name)
		}
	}
}

func TestValidate(t *testing.T) {
	namespace := "test"
	handler := &Localtime{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "123", Namespace: namespace, Annotations: map[string]string{"kubernetes.io/localtime": "true"}},
		Spec: api.PodSpec{
			Volumes: []api.Volume{
				{
					Name:         localtimeName,
					VolumeSource: api.VolumeSource{HostPath: &api.HostPathVolumeSource{Path: localtimePath}},
				},
			},
			InitContainers: []api.Container{
				{Name: "init1", Image: "image", VolumeMounts: []api.VolumeMount{
					{Name: localtimeName, ReadOnly: true, MountPath: localtimePath},
				}},
			},
			Containers: []api.Container{
				{Name: "ctr1", Image: "image", VolumeMounts: []api.VolumeMount{
					{Name: localtimeName, ReadOnly: true, MountPath: localtimePath},
				}},
			},
		},
	}
	err := handler.Validate(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, false, nil), nil)
	if err == nil {
		t.Log("validate success")
	}
}
