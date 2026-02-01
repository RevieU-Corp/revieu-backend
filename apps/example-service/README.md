# Example Service

This is a placeholder directory for future microservices.

## Purpose

RevieU uses a multi-language microservice architecture. This directory serves as:
- A template/example for adding new services
- A placeholder to maintain the apps/ directory structure

## Adding a New Service

When creating a new microservice:

1. Copy this directory structure or create a new one
2. Choose your technology stack (Go, Python, Node.js, etc.)
3. Follow the naming convention: `apps/<service-name>`
4. Add appropriate CI/CD workflows in `.github/workflows/`

## Current Services

- **core** (Go) - Main API service handling user profiles, authentication, and core business logic

## Notes

- Each service should be independently deployable
- Services communicate via REST APIs or message queues
- Each service has its own database schema (if needed)
