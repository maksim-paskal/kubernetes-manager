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
package cleanoldtags

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	utilsgo "github.com/maksim-paskal/utils-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const deletePrefix = "rm -rf "

type RegistryData struct {
	ProjectID     string
	DockerTag     string
	TagsNotDelete []string
}

var exceptions []string

func Execute(rootSpan opentracing.Span) error {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("cleanOldTagsBy", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	var err error

	projectIDs := []string{}
	projectOrigins := []string{}

	exceptions, err = getExceptions(span)
	if err != nil {
		return err
	}

	ingresss, err := api.GetIngress()
	if err != nil {
		return err
	}

	for _, ingress := range ingresss {
		log := log.WithFields(log.Fields{
			"name":      ingress.IngressName,
			"namespace": ingress.Namespace,
		})

		projectID := ingress.IngressAnotations[config.LabelGitProjectID]
		projectOrigin := ingress.IngressAnotations[config.LabelGitProjectOrigin]

		if utilsgo.StringInSlice(projectID, projectIDs) {
			log.Warnf("projectId=%s already in array", projectID)

			continue
		}

		if len(projectOrigin) == 0 {
			log.Warnf("%s is empty", config.LabelGitProjectOrigin)

			continue
		}

		projectIDs = append(projectIDs, ingress.IngressAnotations[config.LabelGitProjectID])
		projectOrigins = append(projectOrigins, ingress.IngressAnotations[config.LabelGitProjectOrigin])
	}

	items := []RegistryData{}

	for i, projectID := range projectIDs {
		dockerTag := strings.Split(projectOrigins[i], ":")[1]
		dockerTag = strings.TrimSuffix(dockerTag, ".git")

		tagsNotDelete, err := cleanOldTagsByProject(rootSpan, projectID)
		if err != nil {
			return err
		}

		item := RegistryData{
			ProjectID:     projectID,
			DockerTag:     dockerTag,
			TagsNotDelete: tagsNotDelete,
		}

		items = append(items, item)
	}

	hub, err := registry.New(*config.Get().RegistryURL, *config.Get().RegistryUser, *config.Get().RegistryPassword)
	if err != nil {
		return errors.Wrap(err, "can not connect to registry")
	}

	hub.Logf = registry.Quiet

	var deleteCommand strings.Builder

	deleteCommand.WriteString("set -ex\n")

	for _, item := range items {
		commands, err := exec(span, hub, fmt.Sprintf("%s/", item.DockerTag), item.TagsNotDelete)
		if err != nil {
			return errors.Wrap(err, "can not get commands")
		}

		for _, command := range commands {
			deleteCommand.WriteString(command)
		}
	}

	const (
		resultFile           = "cleanOldTags.sh"
		resultFilePermission = 0o744
	)

	err = ioutil.WriteFile(resultFile, []byte(deleteCommand.String()), resultFilePermission)
	if err != nil {
		return errors.Wrap(err, "can not write file")
	}

	log.Infof("%s created", resultFile)

	return nil
}

func getExceptions(rootSpan opentracing.Span) ([]string, error) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("getExceptions", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	allExceptions := []string{}

	datas, err := api.GetCleanOldTagsConfig()
	if err != nil {
		return nil, err
	}

	for _, data := range datas {
		for _, row := range strings.Split(data, "\n") {
			data := strings.Split(row, ":")
			if len(data) == config.KeyValueLength {
				if !utilsgo.StringInSlice(row, allExceptions) {
					allExceptions = append(allExceptions, row)
				}
			}
		}
	}

	return allExceptions, nil
}

func cleanOldTagsByProject(rootSpan opentracing.Span, projectID string) ([]string, error) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("cleanOldTagsByProject", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	nonDelete := []string{}

	for _, exc := range exceptions {
		data := strings.Split(exc, ":")
		if data[0] == projectID {
			nonDelete = append(nonDelete, data[1])
		}
	}

	ingresss, err := api.GetIngress()
	if err != nil {
		return nil, err
	}

	for _, ingress := range ingresss {
		if ingress.IngressAnotations[config.LabelGitProjectID] == projectID {
			tag := ingress.IngressAnotations[config.LabelRegistryTag]
			if !utilsgo.StringInSlice(tag, nonDelete) {
				nonDelete = append(nonDelete, tag)
			}
		}
	}

	log.Infof("projectID=%s, tags to not delete=%s", projectID, nonDelete)

	return nonDelete, nil
}

func exec(
	rootSpan opentracing.Span,
	hub *registry.Registry,
	checkRepository string,
	tagsToLeaveArray []string,
) ([]string, error) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("exec", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	var (
		deleteTags []string
		errorTags  []string
		err        error
	)

	repositories, err := hub.Repositories()
	if err != nil {
		return nil, errors.Wrap(err, "can not list repositories")
	}

	releasePattern, err := regexp.Compile(*config.Get().ReleasePatern)

	releaseMaxDate := time.Time{}

	if err != nil {
		return nil, errors.Wrap(err, "can not compile regexp")
	}

	log.Debug("start list")

	for _, repository := range repositories {
		log.Debugf("repository=%s", repository)

		if strings.HasPrefix(repository, checkRepository) {
			tags, err := hub.Tags(repository)
			if err != nil {
				log.
					WithError(err).
					WithField(logrushookopentracing.SpanKey, span).
					Error()
			}

			for _, tag := range tags {
				log.Debugf("repository=%s,tag=%s", repository, tag)
				digest, err := hub.ManifestDigest(repository, tag)

				/* find max release date */
				if releasePattern.MatchString(tag) {
					releaseDate, err := time.Parse("20060102", releasePattern.FindStringSubmatch(tag)[1])

					if err != nil {
						log.
							WithError(err).
							WithField(logrushookopentracing.SpanKey, span).
							Error()
					} else if releaseDate.After(releaseMaxDate) {
						releaseMaxDate = releaseDate
					}
				}

				if err != nil {
					log.
						WithError(err).
						WithField(logrushookopentracing.SpanKey, span).
						Error()

					errorTags = append(errorTags, fmt.Sprintf("%s:%s", repository, tag))
				} else {
					log.Debugf("%s:%s,%s", repository, tag, digest)

					if !utilsgo.StringInSlice(tag, tagsToLeaveArray) {
						deleteTags = append(deleteTags, fmt.Sprintf("%s:%s", repository, tag))
					}
				}
			}
		}
	}

	log.Debugf("finished")

	var releaseNotDelete []string

	if (releaseMaxDate != time.Time{}) {
		for _, tagToDelete := range deleteTags {
			tag := strings.Split(tagToDelete, ":")

			/* find releases in range */
			if releasePattern.MatchString(tag[1]) {
				releaseDate, err := time.Parse("20060102", releasePattern.FindStringSubmatch(tag[1])[1])

				if err != nil {
					log.
						WithError(err).
						WithField(logrushookopentracing.SpanKey, span).
						Error()
				} else {
					releaseDateDiffDays := releaseMaxDate.Sub(releaseDate).Hours() / config.HoursInDay

					if releaseDateDiffDays < float64(*config.Get().ReleaseNotDeleteDays) {
						log.Debugf("image %s date in notDeleteDays", tagToDelete)
						releaseNotDelete = append(releaseNotDelete, tagToDelete)
					}
				}
			}
		}
	}

	log.Infof(
		"checkRepository=%s,errorTags=%d,deleteTags=%d,releaseNotDelete=%d",
		checkRepository,
		len(errorTags),
		len(deleteTags),
		len(releaseNotDelete),
	)

	deleteCommand := make([]string, 0)

	for _, errorTag := range errorTags {
		image := strings.Split(errorTag, ":")
		deleteCommand = append(
			deleteCommand,
			fmt.Sprintf(
				"%s%sdocker/registry/v2/repositories/%s/_manifests/tags/%s\n",
				deletePrefix,
				*config.Get().RegistryDirectory,
				image[0],
				image[1],
			),
		)
	}

	for _, tagToDelete := range deleteTags {
		if !utilsgo.StringInSlice(tagToDelete, releaseNotDelete) {
			image := strings.Split(tagToDelete, ":")
			deleteCommand = append(
				deleteCommand,
				fmt.Sprintf(
					"%s%sdocker/registry/v2/repositories/%s/_manifests/tags/%s\n",
					deletePrefix,
					*config.Get().RegistryDirectory,
					image[0],
					image[1],
				),
			)
		}
	}

	return deleteCommand, nil
}
