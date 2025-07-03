FROM golang:1.22-bookworm AS builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    git \
    wget

# 1. Build the Go service
WORKDIR /src
COPY main.go .

RUN go mod init eflomal-service && go mod tidy
# Build the Go binary. CGO_ENABLED=0 creates a static binary with no C dependencies.
RUN CGO_ENABLED=0 go build -o /app/eflomal-service .


# 2. Build eflomal
# We can do this in a separate directory within the same builder stage
RUN git clone --depth 1 https://github.com/robertostling/eflomal.git /eflomal
RUN make -C /eflomal/src


# 3. Get the atools binary from fast_align
# Instead of installing the .deb, we can just extract the binary from it.
RUN wget -c https://github.com/Unbabel/fast-align-deb/raw/refs/heads/master/fast-align_2016.05.31-1_amd64.deb && \
    dpkg-deb -x fast-align_2016.05.31-1_amd64.deb /fast_align_pkg && \
    rm fast-align_2016.05.31-1_amd64.deb


# --- Final Stage ---
FROM debian:bookworm-slim

# Install runtime dependencies for eflomal
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# Set a non-root user for better security
RUN useradd -ms /bin/bash appuser
USER appuser

WORKDIR /app

# Copy ONLY the necessary compiled artifacts from the builder stage
COPY --from=builder /app/eflomal-service /app/eflomal-service
COPY --from=builder /eflomal/src/eflomal /app/eflomal
COPY --from=builder /fast_align_pkg/usr/bin/atools /usr/bin/atools

EXPOSE 8000

# The Go binary is the entrypoint.
ENTRYPOINT ["/app/eflomal-service"]
