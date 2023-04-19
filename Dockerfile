FROM golang:1.17.6

RUN mkdir /app
WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY kmstool_enclave .

RUN go mod download

COPY . .
COPY libnsm.so /usr/lib64/
COPY libnsm.so /lib64/
COPY bn256_aggregatesign /lib64/

ENV GOPATH=/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin
ENV GO111MODULE=on
ENV LD_LIBRARY_PATH=/usr/lib64/:/lib64/

RUN go build -o main .
CMD ["/app/main"]
