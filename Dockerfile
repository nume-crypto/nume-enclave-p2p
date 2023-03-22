FROM golang:1.17.6

RUN mkdir /app
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN chmod a+x bn256_aggregatesign
RUN chmod a+x bn256_verify
ENV GOPATH=/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin
ENV GO111MODULE=on

RUN go build -o main .
CMD ["/app/main"]
