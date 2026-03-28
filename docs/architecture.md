# Architecture Map

- Backend entrypoint: `backend/cmd/server/main.go`
- Router/API namespace: `backend/internal/api/router.go` (`/api/v1/*`)
- Domain handlers: `backend/internal/handlers/handlers.go`
- Security middleware: `backend/internal/middleware/*`
- Core business services: `backend/internal/services/*`
- Data runtime store: `backend/internal/store/*`
- SQL model baseline: `backend/migrations/001_init.sql`
- OpenAPI draft: `backend/openapi/openapi.yaml`
- Frontend app: `frontend/src/main.jsx`
- Frontend offline queue: `frontend/src/offline/queue.js`
