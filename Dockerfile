FROM golang:1.20

RUN mkdir /app
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV GOPATH=/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin
ENV GO111MODULE=on
RUN chmod a+x bn256_aggregatesign_darwin
RUN chmod a+x bn256_aggregatesign_linux
RUN go build -o main .
CMD ["/app/main"]
