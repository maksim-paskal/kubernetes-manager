FROM node:lts as front

WORKDIR /app
COPY front /app
RUN yarn install && yarn generate


FROM golang:1.12 as build

COPY main.go /usr/src/kubernetes-manager/main.go
COPY go.mod /usr/src/kubernetes-manager/go.mod
COPY go.sum /usr/src/kubernetes-manager/go.sum

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN cd /usr/src/kubernetes-manager \
  && go mod download \
  && go mod verify \
  && go build -v -o kubernetes-manager -ldflags "-X main.buildTime=$(date +"%Y%m%d%H%M%S")"

FROM alpine:latest

COPY --from=front /app/dist /app/dist
COPY --from=build /usr/src/kubernetes-manager/kubernetes-manager /app/kubernetes-manager

RUN apk add --no-cache ca-certificates
CMD /app/kubernetes-manager --front.dist=/app/dist