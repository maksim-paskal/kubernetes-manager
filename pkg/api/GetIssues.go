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
	"slices"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
)

const (
	GetIssuesPeriod = "10h"
	GetIssuesLimit  = "25"
)

func (e *Environment) GetIssues(ctx context.Context) (*types.GetIssuesResponse, error) {
	ctx, span := telemetry.Start(ctx, "api.GetIssues")
	defer span.End()

	if client.GetSentryClient() == nil {
		return nil, errors.New("Sentry client not initialized")
	}

	query := "release:" + e.Namespace + "@*"

	issues, err := client.GetSentryClient().GetIssues(ctx, query, GetIssuesPeriod, GetIssuesLimit)
	if err != nil {
		return nil, errors.Wrap(err, "error getting issues")
	}

	result := &types.GetIssuesResponse{
		IssuesExternalLink: client.GetSentryClient().GetIssuesExternalLink(query, GetIssuesPeriod),
		Items:              make([]types.Issue, len(issues)),
		IssuesProjects:     make([]string, 0),
	}

	for i, issue := range issues {
		result.Items[i] = types.Issue{
			ID:            issue.ID,
			ShortID:       issue.ShortID,
			Project:       issue.Project.Name,
			ProjectURL:    client.GetSentryClient().GetProjectURL(issue.Project.Slug),
			Culprit:       issue.Culprit,
			Level:         issue.Level,
			Title:         issue.Metadata.Type,
			Description:   issue.Metadata.Value,
			Link:          issue.PermaLink,
			LastSeen:      utils.TimeToString(issue.LastSeenTime()),
			LastSeenShort: utils.HumanizeDuration(utils.HumanizeDurationShort, time.Since(issue.LastSeenTime())),
			IsNew:         issue.SubStatus == "new",
		}

		if !slices.Contains(result.IssuesProjects, issue.Project.Name) {
			result.IssuesProjects = append(result.IssuesProjects, issue.Project.Name)
		}
	}

	return result, nil
}
