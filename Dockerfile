FROM node:lts as front

ARG APPVERSION=dev

WORKDIR /app
COPY front /app
RUN yarn install && yarn generate

FROM alpine:latest

COPY --from=front /app/dist /app/dist
COPY ./kubernetes-manager /app/kubernetes-manager

# app env
ENV KUBERNETES_ENDPOINT=https://api:6443
ENV GITLAB_URL=https://git/api/v4
ENV GITLAB_TOKEN=some-token
ENV SYSTEM_GIT_TAGS=^master$|^release-.*
ENV SYSTEM_NAMESPACES=^kube-system$
ENV FRONT_PHPMYADMIN_URL=https://aaa
ENV FRONT_DEBUG_SERVER_NAME=bbb
ENV FRONT_SENTRY_DSN="https://id@sentry/1"

RUN apk upgrade \
&& apk add --no-cache ca-certificates tzdata

ENTRYPOINT [ "/app/kubernetes-manager" ]

CMD [ "--front.dist=/app/dist" ]
