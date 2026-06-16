FROM golang:1.26-alpine AS base   
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM base AS dev
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", ".air.toml"]

FROM base AS prod
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /manager ./cmd/loanTracker/main.go

CMD ["/manager"]

