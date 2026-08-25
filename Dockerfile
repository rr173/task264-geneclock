# 单阶段构建：合成生物基因调控时序复核台
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
