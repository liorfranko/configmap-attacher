# # build image
# FROM ubuntu:latest
# RUN apt-get update
# RUN apt-get install -y wget git gcc

# RUN wget -P /tmp https://go.dev/dl/go1.17.6.linux-amd64.tar.gz
# RUN rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go1.17.6.linux-amd64.tar.gz

# # RUN export PATH=$PATH:/usr/local/go/bin
# ENV PATH $PATH:/usr/local/go/bin

# WORKDIR /app
# COPY . .

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/configmap-attacher

# # executable image
# COPY kubectl /bin/kubectl
# RUN chmod u+x /bin/kubectl

# CMD ["/bin/bash"]

# build image
FROM golang:1.17.6-alpine as builder
RUN apk update && apk add git ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/configmap-attacher

# executable image
FROM scratch
COPY --from=builder /go/bin/configmap-attacher /go/bin/configmap-attacher

ENV VERSION 1.1.5
ENTRYPOINT ["/go/bin/configmap-attacher"]
