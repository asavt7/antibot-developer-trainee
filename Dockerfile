FROM golang:1.16 as builder
ADD . /
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v -o /bin/main /cmd/main.go
FROM scratch as app
COPY --from=builder /bin/main /
COPY ./static /static
CMD ["/main"]