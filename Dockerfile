##
## Build phase
##
FROM docker.io/golang:alpine as builder
WORKDIR /build/src 
ADD . /build/src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o memuser .

##
## Deploy
##
FROM alpine
COPY --from=builder /build/src/memuser /app/
WORKDIR /app
EXPOSE 8080
CMD ["/app/memuser", "-fast", "-maxmemory", "1000"]