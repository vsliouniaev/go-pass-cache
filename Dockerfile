FROM golang:1.12-alpine AS build
ARG VERSION=0.0.0
ARG PACKAGE="github.com/vsliouniaev/go-pass-cache"
WORKDIR /go/src/${PACKAGE}
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o main -ldflags \
    "-X ${PACKAGE}/core.Version=${VERSION} -X ${PACKAGE}/core.BuildTime=$(date -u +%FT%TZ)"
RUN mv main /main

FROM gcr.io/distroless/base
COPY --from=build /main /passcache
COPY ./www /www
ENTRYPOINT ["/passcache"]