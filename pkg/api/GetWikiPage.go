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
package api

import (
	"context"
	"strconv"

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetWikiPageInput struct {
	projectID int
	ProjectID string
	Slug      string
}

func (input *GetWikiPageInput) Validate() error {
	if len(input.ProjectID) == 0 {
		return errors.New("ProjectID can not be empty")
	}

	projectIDInt, err := strconv.Atoi(input.ProjectID)
	if err != nil {
		return errors.Wrapf(err, "can not convert ProjectID %s to int", input.ProjectID)
	}

	input.projectID = projectIDInt

	if input.Slug == "" {
		return errors.New("Slug can not be empty")
	}

	return nil
}

type GetWikiPageItem struct {
	Title   string
	Content string
	EditURL string
}

func GetWikiPage(ctx context.Context, input *GetWikiPageInput) (*GetWikiPageItem, error) {
	ctx, span := telemetry.Start(ctx, "api.GetWikiPage")
	defer span.End()

	gitlabClient := client.GetGitlabClient()

	if gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	if err := input.Validate(); err != nil {
		return nil, errors.Wrap(err, "input.Validate")
	}

	project, _, err := gitlabClient.Projects.GetProject(
		input.ProjectID,
		&gitlab.GetProjectOptions{},
		gitlab.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "can not get project")
	}

	wikiPage, _, err := gitlabClient.Wikis.GetWikiPage(
		input.ProjectID,
		input.Slug,
		&gitlab.GetWikiPageOptions{
			RenderHTML: gitlab.Bool(true),
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can not get wiki page")
	}

	result := &GetWikiPageItem{
		Title:   wikiPage.Title,
		Content: wikiPage.Content,
		EditURL: project.WebURL + "/wikis/" + input.Slug + "/edit",
	}

	return result, nil
}
