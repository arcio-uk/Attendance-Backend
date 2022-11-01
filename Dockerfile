# syntax=docker/dockerfile:1
FROM golang:1.18-alpine

WORKDIR /app

# Copy in source code
COPY . .

# Compile code
RUN go build

# expose port 8010
# If this isn't the bind port in your .env,
# change the line below to match your config.
EXPOSE 8010

# Run web server
CMD [ "/app/attendance-system" ]
