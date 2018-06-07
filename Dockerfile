FROM golang:1.10.1 as build

WORKDIR /go/workspace
COPY . .

ENV GOPATH=/go/workspace
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go get github.com/Mellanox/rdmamap github.com/Mellanox/sriovnet github.com/docker/docker/api/types github.com/spf13/cobra  github.com/vishvananda/netlink github.com/docker/docker/client
RUN go install -ldflags="-s -w -X main.GitCommitId=$GIT_COMMIT -extldflags "-static"" -v docker_rdma_sriov

FROM debian:stretch-slim
COPY --from=build /go/workspace/bin/docker_rdma_sriov /bin/docker_rdma_sriov

CMD ["cp", "-f", "/bin/docker_rdma_sriov", "/tmp/docker_rdma_sriov"]
