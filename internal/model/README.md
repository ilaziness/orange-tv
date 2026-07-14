# Model Layer

This directory contains data model definitions.

## Purpose

Models represent database tables and domain entities.

## Phase 2 Implementation

This directory is a placeholder for Phase 2 database integration. Example structure:

```
model/
├── user.go          # User model with Bun tags
└── order.go         # Order model with Bun tags
```

## Guidelines

- Use Bun struct tags for database column mapping
- Include JSON tags for API serialization
- Add validation tags where appropriate
- Keep models focused on data structure only
