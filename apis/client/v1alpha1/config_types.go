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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServerType is the type of server.
type ServerType string

const (
	// ManagementClusterServerType is a management cluster control plane server.
	ManagementClusterServerType ServerType = "managementcluster"

	// GlobalServerType is a global control plane server.
	GlobalServerType ServerType = "global"
)

// Server connection.
type Server struct {
	// Name of the server.
	Name string `json:"name,omitempty" yaml:"name"`

	// Type of the endpoint.
	Type ServerType `json:"type,omitempty" yaml:"type"`

	// GlobalOpts if the server is global.
	GlobalOpts *GlobalServer `json:"globalOpts,omitempty" yaml:"globalOpts"`

	// ManagementClusterOpts if the server is management cluster.
	ManagementClusterOpts *ManagementClusterServer `json:"managementClusterOpts,omitempty" yaml:"managementClusterOpts"`
}

// ManagementClusterServer is the configruation for a management cluster control plane kubeconfig.
type ManagementClusterServer struct {
	// Path to the kubeconfig.
	Path string `json:"path,omitempty" yaml:"path"`

	// The context to use (if required), defaults to current.
	Context string `json:"context,omitempty" yaml:"context"`
}

// GlobalServer is the configuration for a global server.
type GlobalServer struct {
	// Endpoint for the server.
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint"`

	// Auth for the global server.
	Auth GlobalServerAuth `json:"auth,omitempty" yaml:"auth"`
}

// GlobalServerAuth is authentication for a global server.
type GlobalServerAuth struct {
	// Issuer url for IDP, compliant with OIDC Metadata Discovery.
	Issuer string `json:"issuer" yaml:"issuer"`

	// UserName is the authorized user the token is assigned to.
	UserName string `json:"userName" yaml:"userName"`

	// Permissions are roles assigned to the user.
	Permissions []string `json:"permissions" yaml:"permissions"`

	// AccessToken is the current access token based on the context.
	AccessToken string `json:"accessToken" yaml:"accessToken"`

	// IDToken is the current id token based on the context scoped to the CLI.
	IDToken string `json:"IDToken" yaml:"IDToken"`

	// RefreshToken will be stored only in case of api-token login flow.
	RefreshToken string `json:"refresh_token" yaml:"refresh_token"`

	// Expiration times of the token.
	Expiration metav1.Time `json:"expiration" yaml:"expiration"`

	// Type of the token (user or client).
	Type string `json:"type" yaml:"type"`
}

// +kubebuilder:object:root=true

// Config is the Schema for the configs API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// KnownServers available.
	KnownServers []Server `json:"servers,omitempty" yaml:"servers"`

	// CurrentServer in use.
	CurrentServer string `json:"current,omitempty" yaml:"current"`
}

// +kubebuilder:object:root=true

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
