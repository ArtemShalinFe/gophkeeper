FROM golang:1.21 AS build-stage
WORKDIR /app
COPY . .
RUN go mod download && go mod verify
RUN CGO_ENABLED=0 go build -C ./cmd/gclient -o /gclient

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=build-stage /gclient /gclient

ENV GKS_ADDRESS=gserver:${GKSERVER_PORT}

USER nonroot:nonroot
CMD ["/gclient"]