# Repository Layer

This directory contains data access layer implementations.

## Purpose

The repository layer handles database operations and data persistence.

## Phase 2 Implementation

This directory is a placeholder for Phase 2 database integration with Bun. Example structure:

```
repository/
├── user.go          # User repository
└── order.go         # Order repository
```

## Guidelines

- Repositories should only handle data access
- Use Bun for database operations
- Implement transaction support where needed
- Return domain models, not DTOs
