FROM golang:1.11-alpine as build
RUN apk update && apk add git
ADD . /src
WORKDIR /src
ENV GO111MODULE on
ENV CGO_ENABLED 0
RUN go build -o k8s-event-exporter -ldflags '-extldflags "-static"'

FROM scratch
COPY --from=build /src/k8s-event-exporter /k8s-event-exporter
EXPOSE 9102
ENTRYPOINT  [ "/k8s-event-exporter" ]