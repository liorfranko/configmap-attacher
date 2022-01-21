# build image
FROM golang:1.17.6-alpine as builder
RUN apk update && apk add git ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/configmap-attacher

# executable image
FROM scratch
COPY --from=builder /go/bin/configmap-attacher /go/bin/configmap-attacher

ENTRYPOINT ["/go/bin/configmap-attacher"]
