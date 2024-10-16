FROM golang:1.22.5-alpine AS builder

WORKDIR /build

COPY . .

RUN go mod download
RUN go build -o ./docker-commenteer

# Stage 2: Runtime stage (using alpine as the base image)
FROM alpine:latest AS runtime

WORKDIR /app

# Copy the executable from the build stage to the runtime stage
COPY --from=builder /build/docker-commenteer ./docker-commenteer
COPY --from=builder /build/static ./static/

EXPOSE 8090

CMD ["/app/docker-commenteer"]