# DTO Layer

This directory contains Data Transfer Objects for API requests and responses.

## Purpose

DTOs define the structure of API input and output data, separating them from domain models.

## Structure

```
dto/
└── user.go              # User DTOs (request and response in one file)
```

**Naming Convention**:
- One file per entity: `{entity}.go`
- Each file contains both request and response DTOs for that entity
- Example: `user.go` contains `UserRequest` and `UserResponse` structs

## Phase 2 Implementation

This directory is a placeholder for Phase 2 implementation. Use DTOs to:

- Validate incoming request data
- Format outgoing response data
- Hide sensitive model fields
- Provide version-stable API contracts
