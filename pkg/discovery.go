// Copyright 2018
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crdimporter

import (
	"strings"

	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type SchemaPuller interface {
	PullCRDs(resourceNames... string) (map[string]*apiextensionsv1.CustomResourceDefinition, error)
}

type schemaPuller struct {
	discoveryClient *discovery.DiscoveryClient
	openapiSchema *openapi_v2.Document
}
var _ SchemaPuller = &schemaPuller{}

func (sp *schemaPuller) PullCRDs(resourceNames... string) (map[string]*apiextensionsv1.CustomResourceDefinition, error) {
	crds := map[string]*apiextensionsv1.CustomResourceDefinition{}
	apiResourcesLists, err := sp.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, apiResourcesList := range apiResourcesLists {
		for _, apiResource := range apiResourcesList.APIResources {
			for _, resourceName := range resourceNames {
				if resourceName != apiResource.Name {
					continue
				}
				CRDName := apiResource.Name
				if apiResource.Group == "" {
					CRDName = CRDName + ".core"
				} else {
					CRDName = CRDName + "." + apiResource.Group
				}
				var resourceScope apiextensionsv1.ResourceScope
				if apiResource.Namespaced {
					resourceScope = apiextensionsv1.NamespaceScoped
				} else {
					resourceScope = apiextensionsv1.ClusterScoped
				}
				swaggerSpecDefinitionName := apiResource.Group
				if swaggerSpecDefinitionName == "" {
					swaggerSpecDefinitionName = "core"	
				}
				if ! strings.Contains(swaggerSpecDefinitionName, ".") {
					swaggerSpecDefinitionName = "io.k8s.api." + swaggerSpecDefinitionName
				}
				swaggerSpecDefinitionName = swaggerSpecDefinitionName + "." + apiResource.Version + "." + apiResource.Kind
				config := Config {
					SpecDefinitionName: swaggerSpecDefinitionName,
					Group: apiResource.Group,
					Version: apiResource.Version,
					Kind: apiResource.Kind,
					ResourceScope: string(resourceScope),
					EnableValidation: true,
					Plural: apiResource.Name,
					Categories: apiResource.Categories,
					ShortNames: apiResource.ShortNames,
					SpecReplicasPath: "spec.replicas",
					StatusReplicasPath: "status.replicas",
				}
				crds[apiResource.Name] = NewCustomResourceDefinition(config)
				break
			}
		}		
	}
	return crds, nil
}

func NewSchemaPuller(config *rest.Config) (SchemaPuller, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	openapiSchema, err := discoveryClient.OpenAPISchema()
	if err != nil {
		return nil, err
	}
	return &schemaPuller{
		discoveryClient: discoveryClient,
		openapiSchema: openapiSchema, 
	}, nil
}