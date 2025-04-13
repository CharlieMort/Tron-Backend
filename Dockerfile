FROM golang:alpine AS build
RUN apk --no-cache add gcc g++ make git
WORKDIR /app
COPY . .
RUN go mod tidy
RUN GOOS=linux go build -ldflags="-s -w" -o ./tron-backend .

FROM alpine:3.17
WORKDIR /
COPY --from=build /app /app
EXPOSE 3001
ENTRYPOINT /app/tron-backend