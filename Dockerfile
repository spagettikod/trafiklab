FROM --platform=$BUILDPLATFORM golang:1.18.0 AS build
ARG TARGETOS TARGETARCH
WORKDIR /trafiklab
COPY ./ .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags="-extldflags=-static" trafiklab.go

FROM scratch
COPY www /www
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /trafiklab/trafiklab /
ENTRYPOINT [ "/trafiklab" ]