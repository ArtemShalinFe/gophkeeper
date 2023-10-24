FROM golang:1.21 AS build-stage
WORKDIR /app
COPY . .
RUN go mod download && go mod verify
RUN CGO_ENABLED=0 go build -C ./cmd/gserver -o /gserver

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=build-stage /gserver /gserver
USER nonroot:nonroot
ENTRYPOINT ["/gserver"]