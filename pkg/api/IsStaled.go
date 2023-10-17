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
	"fmt"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

type StaledReason string

const (
	StaledReasonLastScaleDays StaledReason = "staledByLastScaleDays"
	StaledReasonNone          StaledReason = ""
)

// simulate IsStaled if diff > 0.
func (e *Environment) IsStaled(diffDays int) (StaledReason, string) {
	reason := ""

	// ignore system namespaces
	if e.IsSystemNamespace() {
		return StaledReasonNone, "system namespace"
	}

	// ignore namespace if created less than 3 days ago
	if e.NamespaceCreatedDays <= config.StaledNewNamespaceDurationDays {
		return StaledReasonNone, "new namespace"
	}

	// if namespace was not scale up more than 10 days (by default) - it is staled
	if e.NamespaceLastScaledDays > *config.Get().RemoveBranchLastScaleDate-diffDays {
		reason = fmt.Sprintf("NamespaceLastScaledDays=%d, RemoveBranchLastScaleDate=%d",
			e.NamespaceLastScaledDays,
			*config.Get().RemoveBranchLastScaleDate,
		)

		return StaledReasonLastScaleDays, reason
	}

	return StaledReasonNone, reason
}
