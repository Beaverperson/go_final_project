FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /final_project

#RUN go test -run ^TestApp$ ./tests

#RUN go test -run ^TestDB$ ./tests

#RUN go test -run ^TestNextDate$ ./tests