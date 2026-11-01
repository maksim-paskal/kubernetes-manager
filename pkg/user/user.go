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
package user

import (
	"context"
	"net/http"
)

// use oauth proxy header for identifying the owner
// include additional headers to ingress
// nginx.ingress.kubernetes.io/auth-response-headers: "X-Auth-Request-User".
const (
	headerUser  = "X-Auth-Request-User"
	headerEmail = "X-Auth-Request-Email"
)

type contextKeys string

const contextSecurityKey contextKeys = "app.security"

type User struct {
	Name  string
	Email string
}

func Get(ctx context.Context) User {
	user, ok := ctx.Value(contextSecurityKey).(User)
	if ok {
		return user
	}

	return User{}
}

func Set(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, contextSecurityKey, user)
}

func FromHeaders(ctx context.Context, header http.Header) context.Context {
	if user := header[headerUser]; len(user) > 0 {
		return Set(ctx, User{
			Name:  header.Get(headerUser),
			Email: header.Get(headerEmail),
		})
	}

	return ctx
}
