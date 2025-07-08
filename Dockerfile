# Stage 1: Python Environment
FROM python:3.9-slim as python-env

# The slim image doesn't include some common tools. Install them.
RUN apt-get update && apt-get install -y python3-venv gcc

WORKDIR /app

COPY emotion-core/requirements.txt .

# Install python packages
RUN pip install --no-cache-dir -r requirements.txt

COPY emotion-core .

# Stage 2: Go Build
FROM golang:1.23-alpine as go-builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./

RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/server main.go

# Stage 3: Final Image
FROM debian:buster-slim

WORKDIR /app

COPY --from=python-env /usr/local/lib/python3.9 /usr/local/lib/python3.9
COPY --from=python-env /app /app/emotion-core
COPY --from=go-builder /go/bin/server /app/

RUN mkdir -p /app/uploads

EXPOSE 8080

ENV BACKEND_URL=http://backend:8080

CMD ["/app/server"]
