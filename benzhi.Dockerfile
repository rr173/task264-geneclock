# 评测构建用（与 Dockerfile 同源，供 build_benzhi_docker.sh 引用）
FROM golang:1.26.3-bookworm

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -o /bin/geneclock ./cmd/geneclock

EXPOSE 8080
ENTRYPOINT ["/bin/geneclock"]
CMD ["--smoke-test"]
