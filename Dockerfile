ARG BUILDER=golang:1.16
FROM ${BUILDER} AS build

ARG GITHUB_USER
ARG GITHUB_PASSWORD
RUN echo "machine github.com login ${GITHUB_USER} password ${GITHUB_PASSWORD}" > ~/.netrc
ENV GOPRIVATE=github.com/IpsoVeritas

WORKDIR /code/realm
ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .

ARG VERSION=0.0.0-snapshot
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/IpsoVeritas/realm/pkg/version.Version=$VERSION" -o /realm ./cmd/realm
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/IpsoVeritas/realm/pkg/version.Version=$VERSION" -o /createKey ./cmd/createKey

FROM alpine:3.11

ADD https://dl.k8s.io/release/v1.20.0/bin/linux/amd64/kubectl /usr/bin/kubectl
RUN chmod +x /usr/bin/kubectl

COPY --from=build /realm /realm
COPY --from=build /createKey /createKey
ADD scripts/createKey.sh /createKey.sh

CMD [ "/realm" ]