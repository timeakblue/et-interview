# Build the Go backend
FROM public.ecr.aws/docker/library/golang:1.26-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Populated by BuildKit; defaults to the build platform. Override for
# cross-builds, e.g. `docker build --platform linux/amd64 .`
ARG TARGETOS=linux
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /app/et-interview .

# Download node_modules
FROM public.ecr.aws/docker/library/node:24-alpine AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

# Build the frontend bundle
FROM public.ecr.aws/docker/library/node:24-alpine AS frontend-builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY frontend/ ./
RUN npm run build

# Deploy backend + bundle
FROM public.ecr.aws/docker/library/alpine:3.24.1 AS deploy
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=backend-builder /app/et-interview /app/et-interview
COPY --from=frontend-builder /app/dist /app/web/

EXPOSE 8080
CMD [ "/app/et-interview" ]
