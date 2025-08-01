FROM golang:1.23.3

WORKDIR /app

RUN apt-get update && apt-get install -y \
    libzmq3-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*


COPY go.mod go.sum ./
RUN go mod download



COPY . .


CMD ["go", "run", "./cmd/main.go"]
