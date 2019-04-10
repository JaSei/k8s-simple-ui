FROM golang:1.11-alpine as build-server
RUN mkdir /dashboard
WORKDIR /dashboard
RUN apk add --update --no-cache ca-certificates git

COPY server/go.mod .
COPY server/go.sum .

RUN go mod download
COPY server/ ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /dashboard/dashboard


FROM node:9.11.1-alpine as build-ui
WORKDIR /ui
COPY ui/package*.json ./
RUN npm install
COPY ui/ ./
RUN npm run build


FROM scratch
LABEL org.label-schema.schema-version = "1.0"
LABEL org.label-schema.vendor = "Jan Seidl"
LABEL org.label-schema.vcs-url = "https://github.com/jasei/k8s-simple-ui"
EXPOSE 8080

COPY --from=build-server /dashboard/dashboard /app/dashboard
COPY --from=build-ui /ui/dist /app/static
WORKDIR /app
ENTRYPOINT ["/app/dashboard"]
