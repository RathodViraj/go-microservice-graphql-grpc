FROM golang:1.25-alpine AS build
RUN apk add --no-cache gcc g++ make ca-certificates
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o account-app ./account/cmd/account

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /build/account-app .
EXPOSE 8080
CMD ["./account-app"]