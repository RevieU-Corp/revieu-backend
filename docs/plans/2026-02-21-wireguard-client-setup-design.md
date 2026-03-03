# WireGuard Client Setup Design

**Date:** 2026-02-21
**Status:** Approved

## Goal
Create a starter WireGuard client configuration on this host using locally generated keys, with server values left as placeholders for later fill-in.

## Selected Approach
Use a static config file at `/etc/wireguard/wg0.conf` and generate a client keypair on this machine.

## Design
- Generate a private key and public key using `wg genkey` and `wg pubkey`.
- Store key material in `/etc/wireguard` with strict permissions.
- Write `/etc/wireguard/wg0.conf` with:
  - `[Interface]`: generated `PrivateKey`, sample `Address`, DNS placeholder.
  - `[Peer]`: placeholders for server `PublicKey`, `Endpoint`, and `AllowedIPs`.
  - `PersistentKeepalive = 25` for NAT-friendly client behavior.
- Do not bring the interface up yet because placeholder peer values are incomplete.

## Validation
- Confirm key files exist and permissions are restricted.
- Confirm `/etc/wireguard/wg0.conf` exists and includes expected sections/placeholders.
