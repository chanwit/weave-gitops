ARG FLUX_VERSION=0.24.1
ARG FLUX_CLI=ghcr.io/fluxcd/flux-cli:v$FLUX_VERSION

# Alias for flux
FROM $FLUX_CLI as flux

# Go build
FROM golang:1.17 AS go-build
# Add known_hosts entries for GitHub and GitLab
RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN ssh-keyscan gitlab.com >> ~/.ssh/known_hosts

COPY Makefile /app/
COPY tools /app/tools
WORKDIR /app
COPY --from=flux /usr/local/bin/flux /app/pkg/flux/bin/flux
COPY go.* /app/
RUN go mod download
COPY . /app
RUN make gitops

# Distroless
FROM gcr.io/distroless/base as runtime
COPY --from=go-build /app/bin/gitops /gitops
COPY --from=go-build /root/.ssh/known_hosts /root/.ssh/known_hosts

ENTRYPOINT ["/gitops"]