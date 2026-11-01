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
package aws

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	rgtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type providerConfig struct {
	AccessKeyID     string
	AccessSecretKey string
	Region          string
}

type Provider struct {
	config    providerConfig
	awsConfig aws.Config
	condition config.WebHook
	message   types.WebhookMessage
}

func (provider *Provider) Init(condition config.WebHook, message types.WebhookMessage) error {
	var err error

	configBytes, err := json.Marshal(condition.Config)
	if err != nil {
		return errors.Wrap(err, "invalid condition config")
	}

	err = json.Unmarshal(configBytes, &provider.config)
	if err != nil {
		return errors.Wrap(err, "invalid config")
	}

	provider.condition = condition
	provider.message = message

	if len(provider.config.Region) == 0 {
		provider.config.Region = os.Getenv("AWS_REGION")
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(provider.config.Region),
	}

	if len(provider.config.AccessKeyID) > 0 && len(provider.config.AccessSecretKey) > 0 {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				provider.config.AccessKeyID,
				provider.config.AccessSecretKey,
				"",
			),
		))
	}

	provider.awsConfig, err = awsconfig.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return errors.Wrap(err, "error while creating aws config")
	}

	return nil
}

func (provider *Provider) validEvent() bool {
	switch provider.message.Event { //nolint:exhaustive
	case types.EventStart:
		return true
	case types.EventStop:
		return true
	default:
		return false
	}
}

func (provider *Provider) Process(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "webhook.aws.Process")
	defer span.End()

	if !provider.validEvent() {
		log.Warn("this event not supported")

		return nil
	}

	processInstances := make(chan error)
	processDatabases := make(chan error)

	go func() {
		processInstances <- provider.processInstances(ctx)
	}()

	go func() {
		processDatabases <- provider.processDatabases(ctx)
	}()

	type Result struct {
		ErrProcessInstances string
		ErrProcessDatases   string
	}

	result := Result{}
	hasError := false

	err := <-processInstances
	if err != nil {
		hasError = true
		result.ErrProcessInstances = err.Error()
	}

	err = <-processDatabases
	if err != nil {
		hasError = true
		result.ErrProcessDatases = err.Error()
	}

	if hasError {
		resultText, err := json.Marshal(result)
		if err != nil {
			log.WithError(err).Error("error while marshaling result")
		}

		return errors.New(string(resultText))
	}

	return nil
}

func (provider *Provider) processInstances(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "webhook.aws.processInstances")
	defer span.End()

	svc := ec2.NewFromConfig(provider.awsConfig)

	params := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name: aws.String("tag:" + config.TagCluster),
				Values: []string{
					provider.message.Cluster,
				},
			},
			{
				Name: aws.String("tag:" + config.TagNamespace),
				Values: []string{
					provider.message.Namespace,
				},
			},
		},
	}

	log.Debug(params)

	resp, err := svc.DescribeInstances(ctx, params)
	if err != nil {
		return errors.Wrap(err, "error while getting instances")
	}

	if len(resp.Reservations) == 0 {
		return errors.New("reservations not found")
	}

	instances := make([]string, 0)

	for _, reservations := range resp.Reservations {
		for _, instance := range reservations.Instances {
			// exclude terminated instances
			if instance.State.Name != ec2types.InstanceStateNameShuttingDown && instance.State.Name != ec2types.InstanceStateNameTerminated {
				instances = append(instances, aws.ToString(instance.InstanceId))
			}
		}
	}

	if len(instances) == 0 {
		return errors.New("instances not found")
	}

	switch provider.message.Event { //nolint:exhaustive
	case types.EventStart:
		result, err := svc.StartInstances(ctx, &ec2.StartInstancesInput{
			DryRun:      new(false),
			InstanceIds: instances,
		})
		if err != nil {
			return errors.Wrap(err, "error while starting database")
		}

		log.Debug(result)
	case types.EventStop:
		result, err := svc.StopInstances(ctx, &ec2.StopInstancesInput{
			DryRun:      new(false),
			InstanceIds: instances,
		})
		if err != nil {
			return errors.Wrap(err, "error while stoping database")
		}

		log.Debug(result)
	default:
		log.Warn("unknown event " + provider.message.Event)
	}

	return nil
}

func (provider *Provider) processDatabases(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "webhook.aws.processDatabases")
	defer span.End()

	resources := resourcegroupstaggingapi.NewFromConfig(provider.awsConfig)

	// list databases by tags, rds.DescribeDBInstances do not use tags for filtering
	dbs, err := resources.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		ResourceTypeFilters: []string{"rds:db"},
		TagFilters: []rgtypes.TagFilter{
			{
				Key:    aws.String(config.TagCluster),
				Values: []string{provider.message.Cluster},
			},
			{
				Key:    aws.String(config.TagNamespace),
				Values: []string{provider.message.Namespace},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "error getting resources databases")
	}

	svc := rds.NewFromConfig(provider.awsConfig)

	for _, resource := range dbs.ResourceTagMappingList {
		database, err := svc.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: resource.ResourceARN,
		})
		if err != nil {
			return errors.Wrap(err, "error getting database")
		}

		status := database.DBInstances[0].DBInstanceStatus

		switch provider.message.Event { //nolint:exhaustive
		case types.EventStart:
			// start instance only if it is stopped or inaccessible-encryption-credentials-recoverable
			if *status != "stopped" && *status != "inaccessible-encryption-credentials-recoverable" {
				log.Warn("database has invalid status=" + *status + " to start")

				continue
			}

			result, err := svc.StartDBInstance(ctx, &rds.StartDBInstanceInput{
				DBInstanceIdentifier: database.DBInstances[0].DBInstanceIdentifier,
			})
			if err != nil {
				return errors.Wrap(err, "error while starting instances")
			}

			log.Debug(result)

		case types.EventStop:
			// stop instance only if it is available
			if *status != "available" {
				log.Warn("database has invalid status=" + *status + " to stop")

				continue
			}

			result, err := svc.StopDBInstance(ctx, &rds.StopDBInstanceInput{
				DBInstanceIdentifier: database.DBInstances[0].DBInstanceIdentifier,
			})
			if err != nil {
				return errors.Wrap(err, "error while stoping instances")
			}

			log.Debug(result)
		default:
			log.Warn("unknown event " + provider.message.Event)
		}
	}

	return nil
}
