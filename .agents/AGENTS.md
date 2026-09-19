# Avtoplaneta Agent Rules

## Communication
- **Language:** ALWAYS communicate with the user in Russian (Русский язык). All explanations, summaries, and questions must be in Russian.

## Inventory and Orders Logic
- **Quantity States:**
  - `quantity > 0`: Regular part available in stock.
  - `quantity == 0`: Defective parts (e.g. from defect reports) or special parts that MUST remain visible in the inventory and XML generation (`quantity >= 0`).
  - `quantity == -1`: Parts that have been completely sold via the `orders-service`.
- **Order Logic:** When an order is created and the part's stock is depleted (i.e. `quantity - amount <= 0`), the `orders-service` sets the part's quantity to `-1` in the database. This acts as a marker so the part is hidden from the UI and XML (since it fails the `quantity >= 0` check) rather than setting it to `0`.
- **Parts Display:** Always ensure that UI components and XML generators use `quantity >= 0` (not `> 0`) to fetch parts from the database or Elasticsearch, ensuring that defective/special `0` quantity parts remain visible.

## Git & Deployment Workflow
- **Committing:** Whenever you make significant changes or fix bugs, always commit them locally using `git add` and `git commit` with descriptive messages.
- **Pushing:** **DO NOT** execute `git push` on behalf of the user. The terminal requires SSH/password authentication, and the user prefers to manually trigger the `git push` to control the CI/CD pipeline and deployments to the production server.
- **Docker Compose:** Use `docker compose` (with a space), NOT the legacy `docker-compose` command, as this is the standard syntax supported on the server.

- **Microservices:** The backend is split into multiple services (`auth-service`, `orders-service`, `parts-service`, `messaging-service`, `export-service`). External traffic is routed directly via **Traefik Ingress** with session verification via **Traefik ForwardAuth** (`auth-service /auth/verify`). The custom monolithic API Gateway has been removed.
- **Communication:** Services use **gRPC** (Protobuf) for 100% of inter-service communication (with HTTP fallback). Always consider gRPC stubs when creating new inter-service endpoints.
- **Database:** PostgreSQL accessed via `sqlc` (for type-safe query generation) and `pgx/v5`. Do not use heavy ORMs. If you change SQL queries, remember to run `sqlc generate`. Each microservice supports its own isolated database (`*_DATABASE_URL`) with fallback to a common `DATABASE_URL`.
- **Search:** Elasticsearch is used for high-performance full-text search.
- **Infrastructure:** Traefik for reverse proxy/TLS and auth forwarding, Redis Streams for asynchronous event publishing (e.g. user rename events, order events).
