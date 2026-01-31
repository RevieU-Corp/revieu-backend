# Secrets Management Design

## Overview

This document describes the secrets management strategy for RevieU Backend deployed on a self-hosted K3s cluster.

## Decision

**Chosen Solution: Sealed Secrets + GitOps**

### Why Sealed Secrets

| Consideration | Analysis |
|---------------|----------|
| Resource constraints | 3 VPS nodes (8c8g + 1c1g + 2c2.5g), cannot afford Vault HA |
| Team workflow | GitOps-friendly, secrets can be safely committed to git |
| Operational overhead | Low, only one controller pod (~50MB memory) |
| Learning value | Industry-standard GitOps practice |

### Alternatives Considered

1. **HashiCorp Vault** - Too resource-intensive for our cluster size
2. **External Secrets Operator (ESO)** - Requires external backend; planned for future when migrating to cloud
3. **K8s native Secrets** - Not secure for git storage

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Developer Local                       │
│  kubeseal CLI ──encrypt──▶ SealedSecret YAML            │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼ git push
┌─────────────────────────────────────────────────────────┐
│                    Git Repository                        │
│  k8s/sealed-secrets/*.yaml (encrypted, safe to commit)  │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼ ArgoCD/Flux sync
┌─────────────────────────────────────────────────────────┐
│                    K3s Cluster                           │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Sealed Secrets Controller (decrypt)            │    │
│  │  SealedSecret ──decrypt──▶ K8s Secret          │    │
│  └─────────────────────────────────────────────────┘    │
│                         │                                │
│                         ▼                                │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Application Pods (read secrets via env/volume) │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

## Secrets Inventory

| Category | Secrets | Sensitivity | Rotation Frequency |
|----------|---------|-------------|-------------------|
| Database | DB_PASSWORD | High | Low (manual) |
| Auth | JWT_SECRET | High | Medium (quarterly) |
| OAuth | GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET | Medium | Low (as needed) |
| Email | SMTP_USERNAME, SMTP_PASSWORD | Medium | Low (as needed) |

## File Structure

```
apps/core/
├── configs/
│   ├── config.yaml              # Non-sensitive config (uses ${ENV_VAR})
│   ├── secrets.yaml             # Local template (gitignored)
│   └── secrets.yaml.example     # Example for new team members
└── k8s/
    ├── README.md                # Usage instructions
    ├── pub-cert.pem             # Sealed Secrets public key
    └── sealed-secrets/          # Encrypted SealedSecrets
        └── *.yaml
```

## Key Management

### Initial Setup

```bash
# Install Sealed Secrets Controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.5/controller.yaml

# Export public key
kubeseal --fetch-cert > apps/core/k8s/pub-cert.pem
```

### Daily Workflow

```bash
# Create secret (do not commit this!)
kubectl create secret generic revieu-secrets \
  --from-literal=DB_PASSWORD=xxx \
  --dry-run=client -o yaml > secret.yaml

# Encrypt with public key
kubeseal --cert apps/core/k8s/pub-cert.pem < secret.yaml > apps/core/k8s/sealed-secrets/revieu-secrets.yaml

# Safe to commit
git add apps/core/k8s/sealed-secrets/revieu-secrets.yaml
```

### Key Backup (Critical)

```bash
# Export private key (store offline securely)
kubectl get secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key -o yaml > master-key-backup.yaml

# Recovery
kubectl apply -f master-key-backup.yaml
kubectl delete pod -n kube-system -l name=sealed-secrets-controller
```

## Future: External Secrets Operator (ESO)

When migrating to cloud or when resources allow, consider ESO for:
- Dynamic secrets rotation
- Integration with cloud-native secret managers (AWS SM, GCP SM)
- Unified CRD interface across environments

See `apps/core/k8s/README.md` for ESO migration plan.
