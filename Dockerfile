# Multi-stage build
FROM golang:1.24-alpine3.22 AS build-base

# Use ARG to declare a build-time variable
ARG GIT_HASH

LABEL org.opencontainers.image.authors="github.com/travboz"
# The version is now embedded in the binary, so this label is optional
LABEL org.opencontainers.image.version=$GIT_HASH
# The version is now embedded in the binary, so this ENV is optional
ENV APP_VERSION=$GIT_HASH

# Install Git in the build stage - for versioning down the line
RUN apk add --no-cache git

# Set workdir to /app
WORKDIR /app

# Copy files needed to install dependencies
COPY go.mod go.sum ./

# Use cache mount to speed up install of existing dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

FROM build-base AS build-production

# Add non root user (Alpine uses adduser, not useradd)
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy the remaining files over
COPY . .

# Compile application during build rather than at runtime
# Add flags to statically link binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X 'main.vcs.revision=$GIT_HASH'" \
    -buildvcs=true \
    -o greenlight-api-binary ./cmd/api

FROM scratch

WORKDIR /

# Copy the passwd and group files for user resolution
COPY --from=build-production /etc/passwd /etc/passwd
COPY --from=build-production /etc/group /etc/group

# Copy the app binary from the build stage
COPY --from=build-production /app/greenlight-api-binary greenlight-api

# Use nonroot user (use the user we created)
USER appuser

# Document expected port
EXPOSE 4000

# Run API binary
CMD ["/greenlight-api"]