ARG GOLANG_VERSION=1
# Pull golang alpine to build binary
FROM golang:${GOLANG_VERSION}-alpine as builder

RUN apk add --no-cache make

WORKDIR /app

# Build binary
COPY . .
RUN make build

# Use alpine to run app
FROM alpine
RUN adduser -D onchain && \
    mkdir /home/onchain/faultdetector && \
    chown -R onchain:onchain /home/onchain/
USER onchain
WORKDIR /home/onchain/faultdetector
COPY --from=builder /app/bin/faultdetector ./bin/
COPY --from=builder /app/config.yaml .

EXPOSE 8080

# Run app
CMD [ "./bin/faultdetector" ]
