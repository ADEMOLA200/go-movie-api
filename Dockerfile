FROM golang:1.19-alpine
WORKDIR /
COPY ./ .
RUN go mod download
RUN go build -o main .
EXPOSE 3000
CMD ["./main"]
