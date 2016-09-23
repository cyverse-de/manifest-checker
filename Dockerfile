FROM golang:1.7-alpine

COPY . /go/src/github.com/cyverse-de/manifest-checker
RUN go install github.com/cyverse-de/manifest-checker

ENTRYPOINT ["manifest-checker"]
CMD ["--help"]
