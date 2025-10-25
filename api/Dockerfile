FROM golang:1.22-alpine 
WORKDIR /app 
COPY . . 
RUN go mod tidy 
CMD ["go","run","cmd/server/main.go"] 
