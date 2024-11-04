FROM golang:1.22 as builder

ENV GOOS=linux 
ENV GOARCH=amd64
 
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download 
COPY . . 
RUN mkdir -p /output
VOLUME ["/output"]
CMD ["go", "build", "-v", "-o", "/output/kensho", "main.go"]
