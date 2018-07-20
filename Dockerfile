FROM golang:1.10.1 as build

WORKDIR /go/workspace
COPY . .

ARG GIT_COMMIT
ENV GOPATH=/go/workspace
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go get github.com/Mellanox/rdmamap github.com/Mellanox/sriovnet github.com/docker/docker/api/types github.com/spf13/cobra  github.com/vishvananda/netlink github.com/docker/docker/client
RUN go install -ldflags="-s -w -X main.GitCommitId=$GIT_COMMIT -extldflags "-static"" -v docker_rdma_sriov
RUN go install -ldflags="-s -w -X main.GitCommitId=$GIT_COMMIT -extldflags "-static"" -v k8sdo

FROM debian:stretch-slim
RUN mkdir /container_tools

COPY --from=build /go/workspace/bin/docker_rdma_sriov /container_tools/docker_rdma_sriov
COPY --from=build /go/workspace/bin/k8sdo /container_tools/k8sdo

CMD cp -f /container_tools/* /tmp/
