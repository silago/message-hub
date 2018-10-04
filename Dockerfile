FROM golang:1.10
RUN mkdir app
ADD . ./app/
WORKDIR ./app 
RUN go get "github.com/rs/xid"
RUN go get "github.com/Jeffail/gabs"
RUN go build -o main .
ENTRYPOINT ["/go/app/main"]
