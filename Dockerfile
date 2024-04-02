FROM golang:alpine as build
COPY cmd /go/src/app/cmd
COPY pkg /go/src/app/pkg
COPY go.mod /go/src/app/
COPY go.sum /go/src/app/

WORKDIR /go/src/app
RUN go mod vendor && go mod tidy && go build -o main cmd/camelo/*.go

FROM scratch
COPY --from=build --chmod=755 /go/src/app/main /main
COPY weaver.toml /weaver.toml
ENTRYPOINT ["/main"]
