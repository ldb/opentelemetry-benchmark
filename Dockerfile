FROM golang:1.17-alpine AS build
WORKDIR /go/src/app
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/benchd cmd/benchd/*.go

FROM scratch
COPY --from=build /out/benchd /
CMD ["/benchd"]