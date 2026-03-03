# pull official base image
FROM golang:1.25-alpine

# set work directory
WORKDIR /backend

# install system dependencies
RUN apk update && apk add --no-cache \
    netcat-openbsd \
    postgresql-client

RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum* ./
RUN go mod download


# Copy the rest of the project
COPY . .

# Copy and setup entrypoint.sh
COPY ./entrypoint.sh /entrypoint.sh
RUN sed -i 's/\r$//g' /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8000

# Run entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

# Default command runs Air for live-reloading
CMD ["air"]