/*
Copyright 2020 The TKG Contributors.

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

package region

// ManagerFactory provides interface for region manager factory
type ManagerFactory interface {
	CreateManager(configPath string) (Manager, error)
}

type managerFactory struct{}

// NewFactory creates new manager factory
func NewFactory() ManagerFactory {
	return &managerFactory{}
}

func (mf *managerFactory) CreateManager(configPath string) (Manager, error) {
	return New(configPath)
}
