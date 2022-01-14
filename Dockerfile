FROM golang:1.17

WORKDIR /app

copy go.mod ./
copy go.sum ./

RUN go mod download

copy * ./

RUN go build -o /docker-qualtrics-vis

CMD ["/docker-gs-ping"]