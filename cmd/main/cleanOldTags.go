package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	sentry "github.com/getsentry/sentry-go"
	"github.com/heroku/docker-registry-client/registry"
	utils "github.com/maksim-paskal/utils-go"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const deletePrefix = "rm -rf "

type RegistryData struct {
	ProjectID     string
	DockerTag     string
	TagsNotDelete []string
}

var exceptions []string

func cleanOldTags(rootSpan opentracing.Span) {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("cleanOldTagsBy", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	projectIDs := []string{}
	projectOrigins := []string{}

	exceptions = getExceptions(span)

	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}

	ingresss, _ := clientset.ExtensionsV1beta1().Ingresses("").List(opt)
	for _, ingress := range ingresss.Items {
		projectID := ingress.Annotations[label_gitProjectId]

		if !utils.StringInSlice(projectID, projectIDs) {
			projectIDs = append(projectIDs, ingress.Annotations[label_gitProjectId])
			projectOrigins = append(projectOrigins, ingress.Annotations[label_gitProjectOrigin])
		}
	}

	items := []RegistryData{}

	for i, projectID := range projectIDs {
		dockerTag := strings.Split(projectOrigins[i], ":")[1]
		dockerTag = strings.TrimSuffix(dockerTag, ".git")

		item := RegistryData{
			ProjectID:     projectID,
			DockerTag:     dockerTag,
			TagsNotDelete: cleanOldTagsByProject(rootSpan, projectID),
		}

		items = append(items, item)
	}

	hub, err := registry.New(*appConfig.registryURL, *appConfig.registryUser, *appConfig.registryPassword)
	if err != nil {
		log.Panic(err)
	}

	hub.Logf = registry.Quiet

	var deleteCommand strings.Builder
	deleteCommand.WriteString("set -ex\n")

	for _, item := range items {
		for _, command := range exec(span, hub, fmt.Sprintf("%s/", item.DockerTag), item.TagsNotDelete) {
			deleteCommand.WriteString(command)
		}
	}
	const resultFile = "cleanOldTags.sh"
	err = ioutil.WriteFile(resultFile, []byte(deleteCommand.String()), 0744)
	if err != nil {
		log.Panic(err)
	}

	log.Infof("%s created", resultFile)
}

func getExceptions(rootSpan opentracing.Span) []string {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("getExceptions", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	allExceptions := []string{}

	opt := metav1.ListOptions{
		LabelSelector: "app=cleanoldtags",
	}

	cms, err := clientset.CoreV1().ConfigMaps(os.Getenv("POD_NAMESPACE")).List(opt)

	if err != nil {
		log.Panic(err)
	}

	log.Infof("found exception configmaps=%d", len(cms.Items))
	for _, cm := range cms.Items {
		cleanoldtags, err := clientset.CoreV1().ConfigMaps(os.Getenv("POD_NAMESPACE")).Get(cm.Name, metav1.GetOptions{})

		if err != nil {
			log.Panic(err)
		}

		data := cleanoldtags.Data["exceptions"]
		for _, row := range strings.Split(data, "\n") {
			data := strings.Split(row, ":")
			if len(data) == 2 {
				if !utils.StringInSlice(row, allExceptions) {
					allExceptions = append(allExceptions, row)
				}
			}
		}
	}

	return allExceptions
}
func cleanOldTagsByProject(rootSpan opentracing.Span, projectID string) []string {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("cleanOldTagsByProject", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	nonDelete := []string{}

	for _, exc := range exceptions {
		data := strings.Split(exc, ":")
		if data[0] == projectID {
			nonDelete = append(nonDelete, data[1])
		}
	}

	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}

	ingresss, _ := clientset.ExtensionsV1beta1().Ingresses("").List(opt)
	for _, ingress := range ingresss.Items {
		if ingress.Annotations[label_gitProjectId] == projectID {
			tag := ingress.Annotations[label_registryTag]
			if !utils.StringInSlice(tag, nonDelete) {
				nonDelete = append(nonDelete, tag)
			}
		}
	}

	log.Infof("projectID=%s, tags to not delete=%s", projectID, nonDelete)
	return nonDelete
}

func exec(rootSpan opentracing.Span, hub *registry.Registry, checkRepository string, tagsToLeaveArray []string) []string {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("exec", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	var deleteTags []string
	var errorTags []string
	var err error
	repositories, err := hub.Repositories()

	if err != nil {
		panic(err)
	}

	releasePattern, err := regexp.Compile(*appConfig.releasePatern)
	var releaseMaxDate time.Time

	if err != nil {
		panic(err)
	}

	log.Debug("start list")
	for _, repository := range repositories {
		log.Debugf("repository=%s", repository)
		if strings.HasPrefix(repository, checkRepository) {
			tags, err := hub.Tags(repository)

			if err != nil {
				log.Error(err)
				logError(span, sentry.LevelInfo, nil, nil, err.Error())
			}

			for _, tag := range tags {
				log.Debugf("repository=%s,tag=%s", repository, tag)
				digest, err := hub.ManifestDigest(repository, tag)

				/* find max release date */
				if releasePattern.MatchString(tag) {
					releaseDate, err := time.Parse("20060102", releasePattern.FindStringSubmatch(tag)[1])

					if err != nil {
						log.Error(err)
						logError(span, sentry.LevelInfo, nil, nil, err.Error())
					} else if releaseDate.After(releaseMaxDate) {
						releaseMaxDate = releaseDate
					}
				}

				if err != nil {
					log.Error(err)
					logError(span, sentry.LevelInfo, nil, nil, err.Error())

					errorTags = append(errorTags, fmt.Sprintf("%s:%s", repository, tag))
				} else {
					log.Debugf("%s:%s,%s", repository, tag, digest)

					if !utils.StringInSlice(tag, tagsToLeaveArray) {
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
					log.Error(err)
					logError(span, sentry.LevelInfo, nil, nil, err.Error())
				} else {
					releaseDateDiffDays := releaseMaxDate.Sub(releaseDate).Hours() / 24

					if releaseDateDiffDays < float64(*appConfig.releaseNotDeleteDays) {
						log.Debugf("image %s date in notDeleteDays", tagToDelete)
						releaseNotDelete = append(releaseNotDelete, tagToDelete)
					}
				}
			}
		}
	}

	log.Infof("checkRepository=%s,errorTags=%d,deleteTags=%d,releaseNotDelete=%d", checkRepository, len(errorTags), len(deleteTags), len(releaseNotDelete))

	var deleteCommand []string

	for _, errorTag := range errorTags {
		image := strings.Split(errorTag, ":")
		deleteCommand = append(
			deleteCommand,
			fmt.Sprintf(
				"%s%sdocker/registry/v2/repositories/%s/_manifests/tags/%s\n",
				deletePrefix,
				*appConfig.registryDirectory,
				image[0],
				image[1],
			),
		)
	}

	for _, tagToDelete := range deleteTags {
		if !utils.StringInSlice(tagToDelete, releaseNotDelete) {
			image := strings.Split(tagToDelete, ":")
			deleteCommand = append(
				deleteCommand,
				fmt.Sprintf(
					"%s%sdocker/registry/v2/repositories/%s/_manifests/tags/%s\n",
					deletePrefix,
					*appConfig.registryDirectory,
					image[0],
					image[1],
				),
			)
		}
	}

	return deleteCommand
}