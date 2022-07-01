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
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type providerConfig struct {
	AccessKeyID     string
	AccessSecretKey string
	Region          string
}

type Provider struct {
	config    providerConfig
	sess      *session.Session
	condition config.WebHook
	message   types.WebhookMessage
}

func (provider *Provider) Init(condition config.WebHook, message types.WebhookMessage) error {
	var err error

	configBytes, err := yaml.Marshal(condition.Config)
	if err != nil {
		return errors.Wrap(err, "invalid condition config")
	}

	err = yaml.Unmarshal(configBytes, &provider.config)
	if err != nil {
		return errors.Wrap(err, "invalid config")
	}

	provider.condition = condition
	provider.message = message

	provider.sess, err = session.NewSession(&aws.Config{
		Region: aws.String(provider.config.Region),
		Credentials: credentials.NewStaticCredentials(
			provider.config.AccessKeyID,
			provider.config.AccessSecretKey,
			"",
		),
	},
	)
	if err != nil {
		return errors.Wrap(err, "error while creating session")
	}

	if len(provider.config.Region) == 0 {
		return errors.New("region is not defined")
	}

	return nil
}

func (provider *Provider) Process() error {
	processInstances := make(chan error)
	processDatabases := make(chan error)

	go func() {
		processInstances <- provider.processInstances()
	}()

	go func() {
		processDatabases <- provider.processDatabases()
	}()

	type Result struct {
		ErrProcessInstances string
		ErrProcessDatases   string
	}

	result := Result{}
	hasError := false

	if err := <-processInstances; err != nil {
		hasError = true
		result.ErrProcessInstances = err.Error()
	}

	if err := <-processDatabases; err != nil {
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

func (provider *Provider) processInstances() error {
	svc := ec2.New(provider.sess, &aws.Config{Region: aws.String(provider.config.Region)})

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + config.TagCluster),
				Values: []*string{
					aws.String(provider.message.Cluster),
				},
			},
			{
				Name: aws.String("tag:" + config.TagNamespace),
				Values: []*string{
					aws.String(provider.message.Namespace),
				},
			},
		},
	}

	log.Debug(params)

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return errors.Wrap(err, "error while getting instances")
	}

	if len(resp.Reservations) == 0 {
		return errors.New("reservations not found")
	}

	instances := make([]*string, 0)

	for _, reservations := range resp.Reservations {
		for _, instance := range reservations.Instances {
			// exclude terminated instances
			if *instance.State.Name != "shutting-down" && *instance.State.Name != "terminated" {
				instances = append(instances, instance.InstanceId)
			}
		}
	}

	if len(instances) == 0 {
		return errors.New("instances not found")
	}

	switch provider.message.Event {
	case types.EventStart:
		result, err := svc.StartInstances(&ec2.StartInstancesInput{
			DryRun:      aws.Bool(false),
			InstanceIds: instances,
		})
		if err != nil {
			return errors.Wrap(err, "error while starting database")
		}

		log.Debug(result.String())
	case types.EventStop:
		result, err := svc.StopInstances(&ec2.StopInstancesInput{
			DryRun:      aws.Bool(false),
			InstanceIds: instances,
		})
		if err != nil {
			return errors.Wrap(err, "error while stoping database")
		}

		log.Debug(result.String())
	default:
		log.Warn("unknown event " + provider.message.Event)
	}

	return nil
}

func (provider *Provider) processDatabases() error {
	resources := resourcegroupstaggingapi.New(provider.sess, &aws.Config{Region: aws.String(provider.config.Region)})

	// list databases by tags, rds.DescribeDBInstances do not use tags for filtering
	dbs, err := resources.GetResources(&resourcegroupstaggingapi.GetResourcesInput{
		ResourceTypeFilters: aws.StringSlice([]string{"rds:db"}),
		TagFilters: []*resourcegroupstaggingapi.TagFilter{
			{
				Key:    aws.String(config.TagCluster),
				Values: aws.StringSlice([]string{provider.message.Cluster}),
			},
			{
				Key:    aws.String(config.TagNamespace),
				Values: aws.StringSlice([]string{provider.message.Namespace}),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "error getting resources databases")
	}

	svc := rds.New(provider.sess, &aws.Config{Region: aws.String(provider.config.Region)})

	for _, resource := range dbs.ResourceTagMappingList {
		database, err := svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: resource.ResourceARN,
		})
		if err != nil {
			return errors.Wrap(err, "error getting database")
		}

		status := database.DBInstances[0].DBInstanceStatus

		switch provider.message.Event {
		case types.EventStart:
			// start instance only if it is stopped or inaccessible-encryption-credentials-recoverable
			if *status != "stopped" && *status != "inaccessible-encryption-credentials-recoverable" {
				log.Warn("database has invalid status=" + *status + " to start")

				continue
			}

			result, err := svc.StartDBInstance(&rds.StartDBInstanceInput{
				DBInstanceIdentifier: database.DBInstances[0].DBInstanceIdentifier,
			})
			if err != nil {
				return errors.Wrap(err, "error while starting instances")
			}

			log.Debug(result.String())

		case types.EventStop:
			// stop instance only if it is available
			if *status != "available" {
				log.Warn("database has invalid status=" + *status + " to stop")

				continue
			}

			result, err := svc.StopDBInstance(&rds.StopDBInstanceInput{
				DBInstanceIdentifier: database.DBInstances[0].DBInstanceIdentifier,
			})
			if err != nil {
				return errors.Wrap(err, "error while stoping instances")
			}

			log.Debug(result.String())

		default:
			log.Warn("unknown event " + provider.message.Event)
		}
	}

	return nil
}
