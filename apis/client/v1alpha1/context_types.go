/*


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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ContextSpec defines the desired state of Context
type ContextSpec struct {
	// OrgID is the organization ID.
	OrgID string `json:"orgId" yaml:"orgId"`

	// GlobalEndpoint to issue global requests.
	GlobalAuth GlobalAuthConfig `json:"globalAuth,omitempty" yaml:"globalAuth"`

	// RegionalEndpoints are the paths to the regional kubeconfig files by name.
	RegionalEndpoints map[string]string `json:"regionalEndpoints,omitempty" yaml:"regionalEndpoints"`
}

// GlobalAuthConfig configuration.
type GlobalAuthConfig struct {
	// Endpoint to issue global requests.
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Issuer url for IDP, compliant with OIDC Metadata Discovery.
	Issuer string `json:"issuer" yaml:"issuer"`
}

// ContextStatus defines the observed state of Context.
type ContextStatus struct {
	// GlobalAuth status.
	GlobalAuth GlobalAuthStatus `json:"globalAuth" yaml:"globalAuth"`
}

// GlobalAuthStatus is the global auth status.
type GlobalAuthStatus struct {
	// UserName is the authorized user the token is assigned to.
	UserName string `json:"userName" yaml:"userName"`

	// Permissions are roles assigned to the user.
	Permissions []string `json:"permissions" yaml:"permissions"`

	// AccessToken is the current access token based on the context.
	AccessToken string `json:"token" yaml:"token"`

	// IDToken is the current id token based on the context scoped to the CLI.
	IDToken string `json:"IDToken" yaml:"IDToken"`

	// RefreshToken will be stored only in case of api-token login flow.
	RefreshToken string `json:"refresh_token" yaml:"refresh_token"`

	// Expiration times of the token.
	Expiration *metav1.Time `json:"expiration" yaml:"expiration"`

	// Type of the token (user or client).
	Type string `json:"type" yaml:"type"`
}

// +kubebuilder:object:root=true

// Context is the Schema for the contexts API
type Context struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContextSpec   `json:"spec,omitempty"`
	Status ContextStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ContextList contains a list of Context
type ContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Context `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Context{}, &ContextList{})
}
