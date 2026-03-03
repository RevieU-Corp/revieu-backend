# WireGuard Client Setup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Provision a local WireGuard client starter config (`wg0`) with generated key material and placeholder server values.

**Architecture:** Use `wireguard-tools` to generate client keys, persist them in `/etc/wireguard`, and assemble a starter `wg0.conf` with secure permissions. Keep interface inactive until real server values are provided.

**Tech Stack:** Ubuntu package-managed `wireguard-tools`, shell commands, `/etc/wireguard` config files.

---

### Task 1: Generate Client Key Material

**Files:**
- Create: `/etc/wireguard/privatekey`
- Create: `/etc/wireguard/publickey`

**Step 1: Write key generation command**
- Run:
  - `umask 077`
  - `wg genkey | tee /etc/wireguard/privatekey | wg pubkey > /etc/wireguard/publickey`

**Step 2: Verify files exist**
- Run: `ls -l /etc/wireguard/privatekey /etc/wireguard/publickey`
- Expected: both files exist, owner-readable only.

### Task 2: Create Starter Client Config

**Files:**
- Create: `/etc/wireguard/wg0.conf`

**Step 1: Build config file**
- Include:
  - `[Interface]` with private key and sample local address.
  - `[Peer]` with placeholders for server values.
  - `PersistentKeepalive = 25`.

**Step 2: Restrict permissions**
- Run: `chmod 600 /etc/wireguard/wg0.conf`
- Expected: file mode prevents non-owner read.

### Task 3: Validate Result

**Files:**
- Verify: `/etc/wireguard/wg0.conf`

**Step 1: Confirm config completeness**
- Run: `wg-quick strip /etc/wireguard/wg0.conf`
- Expected: no parse errors.

**Step 2: Confirm interface not started**
- Run: `wg show wg0`
- Expected: no such device (until user fills peer values and runs `wg-quick up wg0`).
