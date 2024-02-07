# Pull golang alpine to build binary
FROM golang:1.21-alpine3.19 as builder

RUN apk add --no-cache make

WORKDIR /app

# Build binary
COPY . .
RUN make build-app

# Use alpine to run app
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/faultdetector .
COPY --from=builder /app/config.yaml .

EXPOSE 8080

# Run app
CMD [ "./faultdetector" ]
