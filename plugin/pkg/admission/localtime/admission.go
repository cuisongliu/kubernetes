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

// Package alwayspullimages contains an admission controller that modifies every new Pod to force
// the image pull policy to Always. This is useful in a multitenant cluster so that users can be
// assured that their private images can only be used by those who have the credentials to pull
// them. Without this admission controller, once an image has been pulled to a node, any pod from
// any user can use it simply by knowing the image's name (assuming the Pod is scheduled onto the
// right node), without any authorization check against the image. With this admission controller
// enabled, images are always pulled prior to starting containers, which means valid credentials are
// required.
package localtime

import (
	"io"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/admission"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// PluginName indicates name of admission plugin.
const PluginName = "Localtime"
const annotationsLocaltime = "kubernetes.io/localtime"
const localtimePath = "/etc/localtime"
const localtimeName = "kubernetes-localtime"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewAlwaysPullImages(), nil
	})
}

// AlwaysPullImages is an implementation of admission.Interface.
// It looks at all new pods and overrides each container's image pull policy to Always.
type Localtime struct {
	*admission.Handler
}

var _ admission.MutationInterface = &Localtime{}
var _ admission.ValidationInterface = &Localtime{}

// Admit makes an admission decision based on the request attributes
func (a *Localtime) Admit(attributes admission.Attributes, o admission.ObjectInterfaces) (err error) {
	// Ignore all calls to subresources or resources other than pods.
	if shouldIgnore(attributes) {
		return nil
	}
	pod, ok := attributes.GetObject().(*api.Pod)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}
	if v, ok := pod.Annotations[annotationsLocaltime]; ok && v == "true" {

		resultVolumes := pod.Spec.Volumes
		//新增数据localtime
		vVolume := api.Volume{
			Name: localtimeName,
		}
		vVolume.HostPath = &api.HostPathVolumeSource{
			Path: localtimePath,
		}
		resultVolumes = append(resultVolumes, vVolume)
		pod.Spec.Volumes = resultVolumes
		//设置pod相关的/etc/localtime
		for i := range pod.Spec.InitContainers {
			vVolumeMounts := pod.Spec.InitContainers[i].VolumeMounts
			vMount := api.VolumeMount{
				Name:      localtimeName,
				ReadOnly:  true,
				MountPath: localtimePath,
			}
			vVolumeMounts = append(vVolumeMounts, vMount)
			pod.Spec.InitContainers[i].VolumeMounts = vVolumeMounts
		}

		for i := range pod.Spec.Containers {
			vVolumeMounts := pod.Spec.Containers[i].VolumeMounts
			vMount := api.VolumeMount{
				Name:      localtimeName,
				ReadOnly:  true,
				MountPath: localtimePath,
			}
			vVolumeMounts = append(vVolumeMounts, vMount)
			pod.Spec.Containers[i].VolumeMounts = vVolumeMounts
		}
	}

	return nil
}

// Validate makes sure that all containers are set to always pull images
func (*Localtime) Validate(attributes admission.Attributes, o admission.ObjectInterfaces) (err error) {
	if shouldIgnore(attributes) {
		return nil
	}

	pod, ok := attributes.GetObject().(*api.Pod)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}
	if v, ok := pod.Annotations[annotationsLocaltime]; ok && v == "true" {
		resultVolumes := pod.Spec.Volumes
		volumesFlag := false
		for _, v := range resultVolumes {
			if v.Name == localtimeName {
				if v.HostPath != nil && v.HostPath.Path == localtimePath {
					volumesFlag = true
					break
				}
			}
		}
		if !volumesFlag {
			//return error
			return admission.NewForbidden(attributes,
				field.NotSupported(field.NewPath("spec", "volumes"),
					pod.Spec.Volumes, []string{"name:kubernetes-localtime", "path:/etc/localtime"},
				),
			)
		}

		for i := range pod.Spec.InitContainers {
			vVolumeMounts := pod.Spec.InitContainers[i].VolumeMounts
			volumesMountFlag := false
			for _, v := range vVolumeMounts {
				if v.Name == localtimeName {
					if v.MountPath == localtimePath && v.ReadOnly == true {
						volumesMountFlag = true
						break
					}
				}
			}
			if !volumesMountFlag {
				return admission.NewForbidden(attributes,
					field.NotSupported(field.NewPath("spec", "initContainers").Index(i).Child("volumeMounts"),
						pod.Spec.InitContainers[i].VolumeMounts, []string{"name:kubernetes-localtime", "mountPath:/etc/localtime", "readOnly:true"},
					),
				)
			}
		}
		for i := range pod.Spec.Containers {
			vVolumeMounts := pod.Spec.Containers[i].VolumeMounts
			volumesMountFlag := false
			for _, v := range vVolumeMounts {
				if v.Name == localtimeName {
					if v.MountPath == localtimePath && v.ReadOnly == true {
						volumesMountFlag = true
						break
					}
				}
			}
			if !volumesMountFlag {
				return admission.NewForbidden(attributes,
					field.NotSupported(field.NewPath("spec", "containers").Index(i).Child("volumeMounts"),
						pod.Spec.Containers[i].VolumeMounts, []string{"name:kubernetes-localtime", "mountPath:/etc/localtime", "readOnly:true"},
					),
				)
			}
		}
	}
	return nil
}

func shouldIgnore(attributes admission.Attributes) bool {
	// Ignore all calls to subresources or resources other than pods.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != api.Resource("pods") {
		return true
	}

	return false
}

// NewAlwaysPullImages creates a new always pull images admission control handler
func NewAlwaysPullImages() *Localtime {
	return &Localtime{
		//pod updates may not change fields other than `spec.containers[*].image`, `spec.initContainers[*].image`, `spec.activeDeadlineSeconds` or `spec.tolerations` (only additions to existing tolerations)
		Handler: admission.NewHandler(admission.Create),
	}
}
