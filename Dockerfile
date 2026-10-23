FROM golang:1.23-alpine AS build
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /out/sirius ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/sirius .
RUN mkdir -p database
EXPOSE 8080
CMD ["./sirius"]
