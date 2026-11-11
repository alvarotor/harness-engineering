# Builder and runtime bases are pinned by digest (engineering.md, D-0003). Update with `docker buildx imagetools inspect`.
FROM golang:1.26-alpine@sha256:30b1e546a9c7fa3b747c2bdba3a0c7bdf3a02ebb4c5b8405d2f948d7e8aac4ec AS build
ENV GOTOOLCHAIN=local CGO_ENABLED=0
WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/service

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
ARG VERSION=dev
ENV VERSION=$VERSION
COPY --from=build /out/service /service
USER 65532:65532
EXPOSE 8080 8081
HEALTHCHECK --interval=10s --timeout=2s --retries=3 CMD ["/service","healthcheck"]
ENTRYPOINT ["/service"]
