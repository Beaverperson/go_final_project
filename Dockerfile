FROM golang:1.22.5

WORKDIR /appdir

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /scheduler.bin
#'CGO_ENABLED=0', go-sqlite3 requires cgo to work

CMD ["/scheduler.bin"]