/*
* Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

// Package v1alpha1 is the v1alpha1 version of the API.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceRecyclePolicy  to get
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ResourceRecyclePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceRecyclePolicySpec `json:"spec"`
}

// ResourceRecyclePolicySpec for k8s generation
type ResourceRecyclePolicySpec struct {

	// The configuration of the target type. If not set, the pluralName and
	// groupName fields will be set from the metadata.name of this resource. The
	// kind field must be set.
	Target APIResource `json:"target"` // Target to get

	// Whether or not the target type is namespaced.
	Namespaced bool `json:"namespaced"` // Namespace to get
	// If this resource is namespaced,
	Namespace       []string             `json:"namespace"`       // Namespace to get
	Selector        metav1.LabelSelector `json:"selector"`        // Selector to get
	MonitorInterval int                  `json:"monitorInterval"` // MonitorInterval to get

	// Rules map[string] string `json:"rules"`
	Extras map[string]int `json:"extras"` // Extras to get
}

// APIResource for k8s generation
type APIResource struct {
	// Group of the resource.
	Group string `json:"group,omitempty"`
	// Version of the resource.
	Version string `json:"version,omitempty"`
	// Camel-cased singular name of the resource (e.g. ConfigMap)
	Kind string `json:"kind"`
	// Lower-cased plural name of the resource (e.g. configmaps).  If
	// not provided, it will be computed by lower-casing the kind and
	// suffixing an 's'.
	PluralName string `json:"pluralName,omitempty"`
}

func apiResourceToMeta(apiResource APIResource, namespaced bool) metav1.APIResource {
	return metav1.APIResource{
		Group:      apiResource.Group,
		Version:    apiResource.Version,
		Kind:       apiResource.Kind,
		Name:       apiResource.PluralName,
		Namespaced: namespaced,
	}
}

// ResourceRecyclePolicyList +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ResourceRecyclePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ResourceRecyclePolicy `json:"items"`
}
