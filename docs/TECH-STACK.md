# HIS Technology Stack & Learning Resources

Share this comprehensive guide with developers before starting development. All links are official documentation and proven resources.

## Backend Technologies

### Core Language & Runtime
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Go** | 1.22+ | Primary backend language | [Official Go Tour](https://go.dev/tour/welcome/1) | Must complete for all backend devs |
| **gRPC** | Latest | Service-to-service communication | [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/) | Master this — it's our API standard |
| **Protocol Buffers** | proto3 | API contract definition | [Proto3 Guide](https://developers.google.com/protocol-buffers/docs/proto3) | Define all contracts first |

### Database & Migration
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **PostgreSQL** | 13+ | Primary data store | [PostgreSQL Getting Started](https://www.postgresql.org/docs/current/tutorial.html) | Learn ACID, indexes, transactions |
| **pgx/v5** | 5.x | Go PostgreSQL driver | [pgx Documentation](https://github.com/jackc/pgx/wiki) | Type-safe database interactions |
| **sqlx** | Latest | Query builder + scanning | [sqlx GitHub](https://github.com/jmoiron/sqlx) | Simplifies row scanning |
| **golang-migrate** | 4.x | Database migrations | [golang-migrate Docs](https://github.com/golang-migrate/migrate) | All migrations as code (up/down) |

### Message Queue & Events
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Apache Kafka** | 3.x+ | Event streaming & async communication | [Kafka Quickstart](https://kafka.apache.org/quickstart) | Learn partitions, consumer groups |
| **kafka-go** | Latest | Go Kafka client | [kafka-go GitHub](https://github.com/segmentio/kafka-go) | No JVM dependency, pure Go |

### Configuration & Secrets
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Viper** | Latest | Config management (YAML/env) | [Viper Tutorial](https://github.com/spf13/viper) | Environment-driven 12-factor config |
| **env** | Latest | Environment variable parsing | [godotenv Docs](https://github.com/joho/godotenv) | Load from .env files in dev |

### Logging & Observability
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **zerolog** | Latest | Structured JSON logging | [zerolog GitHub](https://github.com/rs/zerolog) | Zero-allocation logger; required format |
| **OpenTelemetry (Go SDK)** | Latest | Distributed tracing & metrics | [OpenTelemetry Go Guide](https://opentelemetry.io/docs/instrumentation/go/) | Trace every RPC handler |
| **Prometheus** | Latest | Metrics collection | [Prometheus Docs](https://prometheus.io/docs/introduction/first_steps/) | Scrape /metrics endpoint every service |
| **Grafana** | Latest | Visualization | [Grafana Getting Started](https://grafana.com/docs/grafana/latest/getting-started/) | Create dashboards for latency/errors |

### Testing & Quality
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Go Testing** | Built-in | Unit testing framework | [Go Testing Guide](https://golang.org/doc/effective_go#testing) | Table-driven tests required |
| **testify** | Latest | Assertion library | [testify GitHub](https://github.com/stretchr/testify) | Use assertions + mock for stubbing |
| **mockery** | Latest | Mock generation | [mockery GitHub](https://github.com/vektra/mockery) | Auto-generate mocks from interfaces |
| **testcontainers-go** | Latest | Containerized test doubles | [testcontainers Docs](https://testcontainers.com/modules/go/) | Start Postgres/Kafka in tests |
| **golangci-lint** | Latest | Linter aggregator | [golangci-lint Docs](https://golangci-lint.run/) | Must pass repo's config before commit |

### Dependency Injection
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Wire** | Latest | Code generation DI | [Wire Guide](https://github.com/google/wire/blob/main/docs/guide.md) | Compile-time DI; no reflection |

### Validation
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **validator** | v10+ | Struct validation | [validator GitHub](https://github.com/go-playground/validator) | Use on all domain entities |

### Cloud & Object Storage
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **AWS SDK v2** | Latest | AWS S3 + cloud services | [AWS SDK Go v2](https://aws.amazon.com/sdk-for-go/) | Works with both AWS S3 and MinIO |
| **MinIO SDK** | Latest | On-prem S3-compatible | [MinIO Go SDK](https://min.io/docs/minio/linux/developers/go/minio-go.html) | Same API as S3; local development |

### Additional Utilities
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **uuid** | Latest | UUID generation | [google/uuid GitHub](https://github.com/google/uuid) | Required for all ID fields |
| **jwt-go** | v5+ | JWT token handling | [jwt-go GitHub](https://github.com/golang-jwt/jwt) | Parse/validate JWT claims |
| **crypto** | Built-in | Encryption (for PHI fields) | [Go Crypto Docs](https://golang.org/pkg/crypto/) | AES-256-GCM for at-rest encryption |

---

## Frontend Technologies

### Core Framework
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **React** | 18+ | UI library | [React Docs](https://react.dev) | Learn hooks, functional components |
| **Next.js** | 14+ | React framework (App Router) | [Next.js App Router Guide](https://nextjs.org/docs) | **Must use App Router** — Pages Router forbidden |
| **TypeScript** | 5+ | Type-safe JavaScript | [TypeScript Handbook](https://www.typescriptlang.org/docs/) | `strict: true` mandatory |

### State Management
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Redux Toolkit** | Latest | State management | [Redux Toolkit Docs](https://redux-toolkit.js.org/introduction/getting-started) | Use slices; no hand-written reducers |
| **Redux-Saga** | Latest | Side effects middleware | [Redux-Saga Tutorial](https://redux-saga.js.org/) | Handle async data fetching here |

### UI Components & Styling
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Radix UI** | Latest | Accessible primitives | [Radix UI Docs](https://www.radix-ui.com/docs/primitives/overview/introduction) | WCAG 2.1 AA compliant; use for all interactive elements |
| **Tailwind CSS** | 3+ | Utility-first CSS | [Tailwind Docs](https://tailwindcss.com/docs) | Style via classes; no inline styles |
| **CSS Custom Properties** | Native | Design tokens | [CSS Variables Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/--*) | Via @his/theme package |

### Internationalization
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **next-intl** | Latest | i18n for Next.js App Router | [next-intl Docs](https://next-intl-docs.vercel.app/) | 8 locales: en, hi, mr, ta, te, kn, gu, bn |

### Form Handling
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **React Hook Form** | Latest | Form state management | [React Hook Form Docs](https://react-hook-form.com/) | Minimal re-renders |
| **Zod** | Latest | Schema validation | [Zod Documentation](https://zod.dev) | Define schemas; validate at API boundaries |

### API Communication
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **gRPC-Web** | Latest | gRPC in browser | [gRPC-Web JS Guide](https://grpc.io/docs/platforms/web/quickstart/) | Generated from .proto files |
| **generated API client** | Auto | Type-safe service clients | [Generated from proto](../proto/README.md) | Never hand-write API calls |

### Testing
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Vitest** | Latest | Fast unit test runner | [Vitest Docs](https://vitest.dev/) | Jest-compatible; better performance |
| **React Testing Library** | Latest | Component testing | [RTL Docs](https://testing-library.com/docs/react-testing-library/intro/) | Test behavior, not implementation |
| **Playwright** | Latest | End-to-end testing | [Playwright Docs](https://playwright.dev/) | Critical workflows E2E tested |

### Code Quality
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **ESLint** | Latest | JavaScript linter | [ESLint Rules](https://eslint.org/docs/rules/) | Must pass before commit |
| **Prettier** | Latest | Code formatter | [Prettier Docs](https://prettier.io/docs/) | Auto-format on save |

### Build & Monorepo
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Turborepo** | Latest | Monorepo orchestration | [Turborepo Handbook](https://turbo.build/repo/docs) | Parallel builds; shared caching |
| **Vite** | Latest | Build tool (per MFE) | [Vite Guide](https://vitejs.dev/guide/) | Fast dev server |
| **Module Federation** | Latest | Micro-frontend composition | [Module Federation Docs](https://webpack.js.org/concepts/module-federation/) | Via Next.js MF plugin |

---

## Infrastructure & DevOps Technologies

### Container & Orchestration
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Docker** | 20.10+ | Container runtime | [Docker Getting Started](https://docs.docker.com/get-started/) | Learn Dockerfile, images, registries |
| **Kubernetes (K8s)** | 1.24+ | Container orchestration | [K8s Interactive Tutorial](https://kubernetes.io/docs/tutorials/kubernetes-basics/) | Deployments, StatefulSets, Services |
| **Helm** | 3+ | Kubernetes package manager | [Helm Charts Guide](https://helm.sh/docs/getting_started/) | Template Kubernetes manifests |

### Infrastructure as Code
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Terraform** | 1.x+ | Infrastructure provisioning | [Terraform AWS Guide](https://www.terraform.io/language) | Define AWS S3, RDS, VPC |
| **kubectl** | Latest | K8s CLI | [kubectl Cheatsheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/) | Deploy and debug in cluster |

### Data Store Infrastructure
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **PostgreSQL** | 13+ | (See Backend section) | [PostgreSQL Docs](https://www.postgresql.org/docs/) | Primary relational store |
| **Redis** | 6+ | Caching layer | [Redis Tutorial](https://redis.io/docs/getting-started/) | Session cache, rate limiting |
| **MinIO** | Latest | Self-hosted S3 | [MinIO Quickstart](https://min.io/docs/minio/container/index.html) | Local file storage dev/testing |

### CI/CD & Git
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **GitHub Actions** | Latest | CI/CD workflow automation | [GitHub Actions Docs](https://docs.github.com/en/actions) | Lint, test, build, push images |
| **Git** | 2.30+ | Version control | [Git Documentation](https://git-scm.com/doc) | Feature branches, squash commits |

---

## Documentation & Communication

### Technical Documentation
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **Markdown** | GFM | Documentation format | [GitHub Markdown Guide](https://github.github.com/gfm/) | READMEs, ADRs, runbooks |
| **Mermaid** | Latest | Diagram as code | [Mermaid Docs](https://mermaid.js.org/) | Architecture diagrams, flowcharts |

### Specification
| Technology | Version | Purpose | Tutorial | Notes |
|-----------|---------|---------|----------|-------|
| **FHIR R4** | R4+ | Healthcare interoperability | [FHIR R4 Overview](https://www.hl7.org/fhir/R4/overview.html) | External integration standard |
| **HL7 v2** | 2.7+ | Legacy healthcare standards | [HL7 v2 Messaging](https://www.hl7.org/implement/standards/product_brief.cfm?product_id=185) | Legacy system integration |

---

## Quick Onboarding Path for New Developers

### Backend Developers (Week 1-2 priorities)
1. Go Basics: [Go Tour](https://go.dev/tour/welcome/1) (4 hrs)
2. gRPC: [gRPC Go Quickstart](https://grpc.io/docs/languages/go/quickstart/) (6 hrs)
3. PostgreSQL: [PostgreSQL Tutorial](https://www.postgresql.org/docs/current/tutorial.html) (4 hrs)
4. This repo's `cmd/server/main.go` and `services/<domain>/internal/domain/` (read-only, 4 hrs)
5. Write first unit test in `services/patient/internal/domain/entity_test.go` (8 hrs)
6. Deploy sample service to cluster via GitHub Actions (4 hrs)

### Frontend Developers (Week 1-2 priorities)
1. React Hooks: [React Docs](https://react.dev/reference/react) (4 hrs)
2. Next.js App Router: [Next.js Guide](https://nextjs.org/docs/getting-started) (6 hrs)
3. Redux Toolkit: [Redux Toolkit Tutorial](https://redux-toolkit.js.org/tutorials/quick-start) (4 hrs)
4. Redux-Saga: [Redux-Saga Getting Started](https://redux-saga.js.org/docs/introduction/GettingSaga) (6 hrs)
5. This repo's `packages/ui/` and `apps/shell/` (read-only, 4 hrs)
6. Build first component using Radix UI + Tailwind (8 hrs)
7. Write first unit test in `apps/<domain>/__tests__/` (4 hrs)

### DevOps/SRE (Week 1-2 priorities)
1. Docker: [Docker Getting Started](https://docs.docker.com/get-started/) (4 hrs)
2. Kubernetes: [K8s Tutorials](https://kubernetes.io/docs/tutorials/) (8 hrs)
3. Terraform: [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (6 hrs)
4. Helm: [Helm Getting Started](https://helm.sh/docs/intro/quickstart/) (4 hrs)
5. Review `infra/k8s/`, `infra/terraform/`, and `infra/helm/` in this repo (4 hrs)
6. Deploy sample service to dev cluster (6 hrs)

---

## Recommended IDE Setup

All developers should use **Visual Studio Code** with these extensions:
- **Go** (golang.go) — Go language support
- **Protocol Buffers** (zxh404.vscode-proto3) — .proto syntax highlighting
- **Kubernetes** (ms-kubernetes-tools.vscode-kubernetes-tools) — K8s manifests
- **Docker** (ms-azuretools.vscode-docker) — Docker file support
- **Prettier** (esbenp.prettier-vscode) — Code formatting
- **ESLint** (dbaeumer.vscode-eslint) — JavaScript linting
- **GitLens** (eamodio.gitlens) — Git history & blame
- **Thunder Client** or **REST Client** — API testing

---

## Critical Pre-Development Checklist

Before starting development, all team members must:
- [ ] Complete "Quick Onboarding Path" for their role (see above)
- [ ] Have read access to AWS / on-premise cloud account
- [ ] Docker running locally (`docker --version` returns 20.10+)
- [ ] Go 1.22+ installed (`go version` returns go1.22+)
- [ ] Node 18+ installed for frontend (`node --version`)
- [ ] kubectl configured for dev cluster (`kubectl cluster-info`)
- [ ] GitHub SSH keys configured (`ssh -T git@github.com`)
- [ ] Cloned all 4 repos (`his-be`, `his-fe`, `his-doc`, `his-prototype-reference`)
- [ ] Run `make setup` or equivalent in each repo
- [ ] All linters passing (`golangci-lint run`, `npm run lint`)
- [ ] All unit tests passing (`go test ./...`, `npm test`)
- [ ] Shared this document with your team ✅
