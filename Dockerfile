FROM node:lts AS front

ARG APPVERSION=dev

# bug in node 16+
ENV NODE_OPTIONS=--openssl-legacy-provider

WORKDIR /app
COPY front /app
RUN yarn install && yarn generate

FROM alpine:latest
ARG TARGETARCH

# app env
ENV KUBERNETES_ENDPOINT=https://api:6443
ENV GITLAB_URL=https://git/api/v4
ENV GITLAB_TOKEN=some-token
ENV FRONT_SENTRY_DSN="https://id@sentry/1"

RUN apk upgrade \
&& apk add --no-cache ca-certificates tzdata \
&& addgroup -g 30001 -S app \
&& adduser -u 30001 -D -S -G app app

COPY --from=front /app/dist /app/dist
COPY ./kubernetes-manager-${TARGETARCH} /app/kubernetes-manager

USER app

ENTRYPOINT [ "/app/kubernetes-manager" ]

CMD [ "--front.dist=/app/dist" ]
