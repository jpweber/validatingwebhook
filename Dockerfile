FROM golang:1.10.1
WORKDIR /go/src/github.com/jpweber/validatingwebhook

RUN go get github.com/golang/dep/cmd/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only

COPY cmd cmd
RUN CGO_ENABLED=0 GOOS=linux go install -ldflags="-w -s" -v github.com/jpweber/validatingwebhook/cmd/webhook

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/webhook /bin/webhook