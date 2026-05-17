# Juriscan

> Intelligent workflow automation for legal operations.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=111827)
![Vite](https://img.shields.io/badge/Vite-6-646CFF?style=for-the-badge&logo=vite&logoColor=white)
![CI](https://img.shields.io/badge/CI-GitHub%20Actions-2088FF?style=for-the-badge&logo=githubactions&logoColor=white)
![Status](https://img.shields.io/badge/Status-Active%20Development-F59E0B?style=for-the-badge)

## Overview

Juriscan is a legal workflow automation platform designed to support operational productivity, structured intake, lead handling, publication analysis, deadline routines and assisted review flows.

The repository is organized as a full-stack product workspace with a Go backend, a React frontend, technical documentation and infrastructure automation. Public documentation intentionally stays at a high level to avoid exposing client data, proprietary logic or sensitive operational details.

## High-Level Architecture

```txt
Users / Operators
       |
       v
React Frontend
       |
       v
Go API Backend
       |
       +--> Identity, OTP, RBAC and audit trails
       +--> Legal workflow modules
       +--> WhatsApp integration layer
       +--> Assisted analysis and validation flows
       |
       v
Persistence / Infrastructure
```

The internal workflow rules and automation strategies are intentionally not documented in detail.

## Tech Stack

- **Backend:** Go
- **Frontend:** React + Vite
- **Persistence:** SQLite for local development, MySQL-compatible deployment path
- **Testing:** Go tests, Vitest and Playwright
- **CI:** GitHub Actions
- **Infrastructure:** Terraform-based deployment assets
- **Runtime:** Designed for cloud-hosted environments

## Core Capabilities

- OTP-based authentication flow
- Role-based access control foundations
- Audit-oriented backend behavior
- Legal publication intake and assisted analysis
- Deadline and validation workflows
- Lead and commercial pipeline routines
- WhatsApp integration layer with mock and provider-based modes
- AI-assisted review surfaces with human validation
- Frontend MVP for operational workflows

## Repository Structure

```txt
.
├── .github/
│   └── workflows/          # CI pipeline
├── deploy/                 # Local deployment helpers
├── docs/
│   ├── brand/              # Brand assets
│   ├── product/            # Product planning and roadmap
│   └── technical/          # Architecture and API documentation
├── infra/                  # Terraform and infrastructure assets
├── juriscan-backend/       # Go API backend
└── juriscan-frontend/      # React frontend
```

## Running Locally

### Backend

```bash
cd juriscan-backend
go mod download
go run ./cmd/juriscan
```

### Frontend

```bash
cd juriscan-frontend
npm install
npm run dev
```

## Tests And Build

### Backend tests

```bash
cd juriscan-backend
go test ./...
```

### Frontend tests

```bash
cd juriscan-frontend
npm run test
```

### Frontend build

```bash
cd juriscan-frontend
npm run build
```

## Scalability And Automation

Juriscan is evolving around a few engineering principles:

- Clear separation between frontend, backend, infrastructure and documentation
- Workflow-oriented backend modules
- Human-in-the-loop validation for assisted analysis
- Provider abstraction for external integrations
- Infrastructure-as-code for reproducible environments
- CI validation for backend and frontend changes

## Roadmap

- Continue hardening authentication, authorization and audit flows
- Expand workflow orchestration across legal routines
- Improve assisted analysis reliability and traceability
- Strengthen observability and operational metrics
- Expand automated test coverage
- Refine deployment automation and environment configuration

## Project Status

Juriscan is in active development.  
The repository reflects an evolving MVP and technical foundation for legal workflow automation.

## Security And Confidentiality

This project must not expose:

- Client data
- Legal documents
- Production credentials
- Infrastructure identifiers
- Internal business rules
- Proprietary automation logic
- Sensitive AI prompts or decision strategies

Some materials may be subject to confidentiality restrictions. Public documentation should remain portfolio-safe and intentionally high level.

## Contributing

Contributions are currently limited to authorized collaborators.  
Before submitting changes, align on scope, security impact and architecture boundaries.

## License

License information should be defined by the repository owner.

---

Built with Go, React and an automation-first approach for modern legal operations.
