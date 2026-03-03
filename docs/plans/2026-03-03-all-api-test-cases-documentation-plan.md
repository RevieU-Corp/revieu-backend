# All API Test Cases Documentation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Produce a complete, shareable markdown test handbook that covers every OpenAPI v1 endpoint in this repository with concrete request/response examples.

**Architecture:** Use `apps/core/docs/swagger.json` as the authoritative source for endpoint coverage, auto-generate a full baseline test-case document grouped by tag, then manually refine top-level structure and usage guidance for team execution.

**Tech Stack:** Go, OpenAPI/Swagger JSON, Markdown.

---

### Task 1: Inventory endpoint coverage

**Files:**
- Read: `apps/core/docs/swagger.json`
- Read: `apps/core/docs/swagger.yaml`

**Step 1: Count all operations**
Run: `awk 'BEGIN{c=0} /^  \/[^:]+:/{p=$1} /^    (get|post|patch|put|delete):/{c++} END{print c}' apps/core/docs/swagger.yaml`
Expected: integer count of operations.

**Step 2: List all paths**
Run: `rg -n "^  /" apps/core/docs/swagger.yaml`
Expected: full path inventory.

### Task 2: Generate full markdown test cases from swagger

**Files:**
- Read: `apps/core/docs/swagger.json`
- Create: `docs/testing/all-api-test-cases-v1.md`

**Step 1: Generate a markdown draft grouped by tag**
Run a Go generator that:
- iterates all paths + methods
- outputs one test case section per operation
- includes path/query/header/body example input
- includes response examples per declared status code
- marks auth requirements

**Step 2: Ensure no endpoint omission**
Run a coverage check comparing operation count in swagger vs generated test case count.
Expected: counts match exactly.

### Task 3: Add team-facing execution guide

**Files:**
- Create/Modify: `docs/testing/README.md`

**Step 1: Provide usage index**
Include:
- target environment variables
- auth/token acquisition
- execution order suggestion (smoke -> domain -> regression)
- known data dependencies

### Task 4: Verify and finalize

**Files:**
- Verify: `docs/testing/all-api-test-cases-v1.md`
- Verify: `docs/testing/README.md`

**Step 1: Validate markdown presence and size**
Run: `wc -l docs/testing/all-api-test-cases-v1.md docs/testing/README.md`
Expected: non-zero line counts.

**Step 2: Confirm clean git diff scope**
Run: `git status --short`
Expected: only documentation files changed.
