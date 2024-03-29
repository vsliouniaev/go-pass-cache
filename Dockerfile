FROM golang:latest AS build
ARG version=0.0.0
ARG created=0
ARG PACKAGE="github.com/vsliouniaev/go-pass-cache"
WORKDIR /go/src/${PACKAGE}
COPY . .
RUN test -z $(gofmt -l .)

RUN go vet   ./...
RUN CGO_ENABLED=1 go test ./... --race
RUN CGO_ENABLED=0 go build -o main -ldflags \
    "-X ${PACKAGE}/core.Version=${version} -X ${PACKAGE}/core.BuildTime=${created}"
RUN mv main /main

FROM scratch
COPY --from=build /main /passcache
COPY ./www /www
ENTRYPOINT ["/passcache"]