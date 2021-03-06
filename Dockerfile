FROM node:14 as front

WORKDIR /app
COPY front /app
RUN yarn install && yarn generate

FROM golang:1.16 as build

WORKDIR /usr/src/kubernetes-manager

COPY ./cmd ./cmd
COPY go.* ./
COPY ./.git ./

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GOFLAGS="-trimpath"

RUN go mod download \
  && go mod verify \
  && go build -v -o kubernetes-manager -ldflags "-X main.gitVersion=$(git describe --tags `git rev-list --tags --max-count=1`)-$(date +%Y%m%d%H%M%S)-$(git log -n1 --pretty='%h')" ./cmd/main \
  && ./kubernetes-manager --version

FROM alpine:3.13

COPY --from=front /app/dist /app/dist
COPY --from=build /usr/src/kubernetes-manager/kubernetes-manager /app/kubernetes-manager

# app env
ENV KUBERNETES_ENDPOINT=https://api:6443
ENV GITLAB_URL=https://git/api/v4
ENV GITLAB_TOKEN=some-token
ENV SYSTEM_GIT_TAGS=^master$,^release-.*
ENV SYSTEM_NAMESPACES=kube-system
ENV FRONT_PHPMYADMIN_URL=https://aaa
ENV FRONT_DEBUG_SERVER_NAME=bbb
ENV FRONT_SENTRY_DSN="https://id@sentry/1"

# rclone params for cleanOldTags
ENV RCLONE_CONFIG_S3_TYPE=s3
ENV RCLONE_CONFIG_S3_PROVIDER=AWS
ENV RCLONE_CONFIG_S3_ACCESS_KEY_ID=change-it
ENV RCLONE_CONFIG_S3_SECRET_ACCESS_KEY=change-it
ENV RCLONE_CONFIG_S3_REGION=eu-central-1

COPY --from=registry:2.7.1 /bin/registry /usr/local/bin
COPY --from=registry:2.7.1 /etc/docker/registry/config.yml /etc/docker/registry/config.yml

RUN apk add --no-cache ca-certificates curl \
&& cd /tmp \
&& curl -o rclone.zip https://downloads.rclone.org/v1.51.0/rclone-v1.51.0-linux-amd64.zip \
&& unzip rclone.zip \
&& mv rclone-v1.51.0-linux-amd64/rclone /usr/local/bin \
&& rm -rf /tmp/*

CMD /app/kubernetes-manager --front.dist=/app/dist
