# Service Layer

This directory contains business logic layer implementations.

## Purpose

The service layer implements business logic and coordinates between handlers and repositories.

## Phase 2 Implementation

This directory is a placeholder for Phase 2 database integration. Example structure:

```
service/
├── user.go          # User service
└── order.go         # Order service
```

## Guidelines

- Services should contain business logic only
- Use dependency injection (Fx) to inject repositories
- Services should be stateless where possible
- Return domain-specific errors using the internal/errors package
