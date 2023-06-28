/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	tagNamespace                = config.Namespace + "-namespace"
	tagCluster                  = config.Namespace + "-cluster"
	azureTypeDBforMySQL         = "Microsoft.DBforMySQL/servers"
	azureTypeDBforMySQLFlexible = "Microsoft.DBforMySQL/flexibleServers"
	azureTypeCompute            = "Microsoft.Compute/virtualMachines"
)

type providerConfig struct {
	SubscriptionID string
	ClientID       string
	ClientSecret   string
	TenantID       string
}

type Provider struct {
	config                providerConfig
	cred                  *azidentity.ClientSecretCredential
	condition             config.WebHook
	message               types.WebhookMessage
	virtualMachinesClient *armcompute.VirtualMachinesClient
	serverClient          *armmysql.ServersClient
	serverClientFlexible  *armmysqlflexibleservers.ServersClient
}

func (provider *Provider) Init(condition config.WebHook, message types.WebhookMessage) error {
	log.Info("init azure provider")

	configBytes, err := json.Marshal(condition.Config)
	if err != nil {
		return errors.Wrap(err, "invalid condition config")
	}

	err = json.Unmarshal(configBytes, &provider.config)
	if err != nil {
		return errors.Wrap(err, "invalid config")
	}

	cred, err := azidentity.NewClientSecretCredential(
		provider.config.TenantID,
		provider.config.ClientID,
		provider.config.ClientSecret,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "invalid auth")
	}

	virtualMachinesClient, err := armcompute.NewVirtualMachinesClient(provider.config.SubscriptionID, cred, nil)
	if err != nil {
		return errors.Wrap(err, "can not create virtual machines client")
	}

	serverClient, err := armmysql.NewServersClient(provider.config.SubscriptionID, cred, nil)
	if err != nil {
		return errors.Wrap(err, "can not create mysql databases client")
	}

	serverClientFlexible, err := armmysqlflexibleservers.NewServersClient(provider.config.SubscriptionID, cred, nil)
	if err != nil {
		return errors.Wrap(err, "can not create mysql databases client")
	}

	provider.cred = cred
	provider.condition = condition
	provider.message = message
	provider.virtualMachinesClient = virtualMachinesClient
	provider.serverClient = serverClient
	provider.serverClientFlexible = serverClientFlexible

	return nil
}

func (provider *Provider) Process(ctx context.Context) error {
	log.Info("process azure provider")

	newClient, err := armresources.NewClient(provider.config.SubscriptionID, provider.cred, nil)
	if err != nil {
		return errors.Wrap(err, "can not create resources client")
	}

	filter := fmt.Sprintf("resourceType eq '%s' or resourceType eq '%s' or resourceType eq '%s'",
		azureTypeDBforMySQL,
		azureTypeDBforMySQLFlexible,
		azureTypeCompute,
	)
	log.Debugf("filter: %s", filter)

	pager := newClient.NewListPager(&armresources.ClientListOptions{
		Filter: &filter,
	})

	resources := make([]*arm.ResourceID, 0)

	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return errors.Wrap(err, "can not get next page")
		}

		for _, v := range nextResult.Value {
			if v.Tags == nil || v.Tags[tagNamespace] == nil || v.Tags[tagCluster] == nil {
				continue
			}

			if *v.Tags[tagNamespace] == provider.message.Namespace && *v.Tags[tagCluster] == provider.message.Cluster {
				resource, err := arm.ParseResourceID(*v.ID)
				if err != nil {
					return errors.Wrapf(err, "can not parse resource id %s", *v.ID)
				}

				log.Debug("add resource: ", resource.String())

				resources = append(resources, resource)
			}
		}
	}

	if len(resources) == 0 {
		log.Warnf("no resources found for filter %s", filter)

		return nil
	}

	for _, resource := range resources {
		name := fmt.Sprintf("%s/%s/%s", resource.ResourceType, resource.ResourceGroupName, resource.Name)

		log.Debugf("%s (%s)", name, resource.String())

		switch provider.message.Event {
		case types.EventStart:
			log.Infof("Starting resource %s", name)

			switch resource.ResourceType.String() {
			case azureTypeDBforMySQL:
				_, err := provider.serverClient.BeginStart(ctx, resource.ResourceGroupName, resource.Name, nil)
				if err != nil {
					return errors.Wrapf(err, "can start resource %s", name)
				}
			case azureTypeDBforMySQLFlexible:
				_, err := provider.serverClientFlexible.BeginStart(ctx, resource.ResourceGroupName, resource.Name, nil)
				if err != nil {
					return errors.Wrapf(err, "can start resource %s", name)
				}
			case azureTypeCompute:
				_, err := provider.virtualMachinesClient.BeginStart(ctx, resource.ResourceGroupName, resource.Name, nil)
				if err != nil {
					return errors.Wrapf(err, "can start resource %s", name)
				}
			default:
				return errors.Errorf("unknown resource type %s", resource.ResourceType.String())
			}
		case types.EventStop:
			log.Infof("Stoping resource %s", name)

			switch resource.ResourceType.String() {
			case azureTypeDBforMySQL:
				_, err := provider.serverClient.BeginStop(ctx, resource.ResourceGroupName, resource.Name, nil)
				if err != nil {
					return errors.Wrapf(err, "can stop resource %s", name)
				}
			case azureTypeDBforMySQLFlexible:
				_, err := provider.serverClientFlexible.BeginStop(ctx, resource.ResourceGroupName, resource.Name, nil)
				if err != nil {
					return errors.Wrapf(err, "can stop resource %s", name)
				}
			case azureTypeCompute:
				_, err := provider.virtualMachinesClient.BeginDeallocate(ctx, resource.ResourceGroupName, resource.Name, nil)
				if err != nil {
					return errors.Wrapf(err, "can stop resource %s", name)
				}
			default:
				return errors.Errorf("unknown resource type %s", resource.ResourceType.String())
			}
		default:
			return errors.Errorf("unknown event %s", provider.message.Event)
		}
	}

	return nil
}
