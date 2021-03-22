FROM golang:1.16

WORKDIR /usr/local/go/src/budget2
COPY . .

EXPOSE 3000
CMD ["/usr/local/go/src/budget2/main"]
