# OAuth gRPC Server (MySQL + Authentication + Authorization)


A minimal example OAuth-like gRPC service in Go. This repo provides:


- proto definition for the OAuth service: `proto/oauth.proto`
- server implementation that issues JWT access and refresh tokens
- a simple in-memory storage implementation for demo and testing
- instructions to swap in persistent storage (MySQL/Postgres)


## Quick start


1. copy `.env.example` to `.env` and set `JWT_SECRET`.
2. generate protobuf stubs:


```bash
protoc --go_out=paths=source_relative:internal/pb --go-grpc_out=paths=source_relative:internal/pb proto/oauth.proto
```


3. build and run:


```bash
go build ./cmd/server && ./server
```


4. test with a gRPC client (grpcurl / BloomRPC / your app):


Register a user:


```bash
grpcurl -plaintext -d '{"username":"alice","password":"secret"}' localhost:50051 oauth.OAuth/Register
```


Get token:


```bash
grpcurl -plaintext -d '{"username":"alice","password":"secret"}' localhost:50051 oauth.OAuth/Token
```


Verify token:


```bash
grpcurl -plaintext -d '{"access_token":"<token>"}' localhost:50051 oauth.OAuth/Verify
```




## Notes and next steps


- This example focuses on clarity and ease-of-use. For production:
- Use secure secrets management (don't keep secret in .env for prod).
- Persist refresh tokens or use rotating refresh tokens to allow revocation.
- Add scopes, clients (client_id/secret), PKCE support for public clients.
- Add rate-limiting, logging, tracing, and observability.