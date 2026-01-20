FROM golang:1.25-alpine AS build-stage
 
WORKDIR /app/src/core
 
COPY . .
 
RUN chmod +x /app/src/core
  
RUN go mod download -x
 
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /core ./cmd/api
 
FROM alpine:latest AS build-release-stage
 
ENV TZ=America/Sao_Paulo
 
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
 
WORKDIR /app/src/core
 
RUN chmod +x /app/src/core
 
COPY --from=build-stage /core /app/src/core/core
 
EXPOSE 8080
 
ENTRYPOINT ["/app/src/core/core"]