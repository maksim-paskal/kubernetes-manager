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
package web

import "errors"

var (
	errNoNamespace          = errors.New("no namespace")
	errNoPod                = errors.New("no pod")
	errNoPodSelected        = errors.New("no pod selected")
	errNoLabelSelector      = errors.New("LabelSelector not set")
	errNoTag                = errors.New("no tag")
	errNoProjectID          = errors.New("projectID not set")
	errNoText               = errors.New("no text")
	errNoOrigin             = errors.New("no origin")
	errNoBranch             = errors.New("no branch")
	errNoCommand            = errors.New("no comand")
	errNoReplicas           = errors.New("no replicas")
	errNoComandFound        = errors.New("no command found")
	errNoPodInStatusRunning = errors.New("pod in status Running not found")
	errUserDeleteALL        = errors.New("user requested deleteALL")
	errUnSupportedVersion   = errors.New("unsupported version")
)
