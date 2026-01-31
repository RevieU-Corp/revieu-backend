# Kubernetes Secrets Management

This directory contains Kubernetes secrets management configuration for RevieU Core service.

## Current Solution: Sealed Secrets

We use [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) for GitOps-friendly secrets management.

### Prerequisites

1. Sealed Secrets Controller installed in cluster
2. `kubeseal` CLI installed locally
3. Public key (`pub-cert.pem`) in this directory

### Quick Start

#### 1. Install Sealed Secrets Controller (one-time)

```bash
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.5/controller.yaml
```

#### 2. Fetch Public Key (one-time per cluster)

```bash
kubeseal --fetch-cert > pub-cert.pem
```

#### 3. Create and Encrypt Secrets

```bash
# Option A: From literal values
kubectl create secret generic revieu-secrets \
  --namespace=revieu \
  --from-literal=DB_PASSWORD='your-password' \
  --from-literal=JWT_SECRET='your-jwt-secret-min-32-chars' \
  --from-literal=SMTP_USERNAME='your-email@gmail.com' \
  --from-literal=SMTP_PASSWORD='your-app-password' \
  --from-literal=GOOGLE_CLIENT_ID='your-client-id' \
  --from-literal=GOOGLE_CLIENT_SECRET='your-client-secret' \
  --dry-run=client -o yaml | kubeseal --cert pub-cert.pem -o yaml > sealed-secrets/revieu-secrets.yaml

# Option B: From secrets.yaml template
# 1. Copy ../configs/secrets.yaml.example to ../configs/secrets.yaml
# 2. Fill in real values
# 3. Run:
kubeseal --cert pub-cert.pem < ../configs/secrets.yaml > sealed-secrets/revieu-secrets.yaml
```

#### 4. Apply to Cluster

```bash
kubectl apply -f sealed-secrets/revieu-secrets.yaml
```

### Directory Structure

```
k8s/
├── README.md           # This file
├── pub-cert.pem        # Sealed Secrets public key (safe to commit)
└── sealed-secrets/     # Encrypted SealedSecrets (safe to commit)
    └── *.yaml
```

### Key Backup & Recovery

**IMPORTANT:** If the cluster is destroyed, all SealedSecrets become unrecoverable without the private key.

```bash
# Backup (store securely offline)
kubectl get secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key -o yaml > master-key-backup.yaml

# Recovery
kubectl apply -f master-key-backup.yaml
kubectl delete pod -n kube-system -l name=sealed-secrets-controller
```

### Updating Secrets

To update a secret:
1. Create new secret YAML with updated values
2. Re-encrypt with `kubeseal`
3. Commit and push
4. GitOps tool (ArgoCD/Flux) will sync automatically

---

## Future: External Secrets Operator (ESO)

When resources allow or when migrating to cloud, consider [External Secrets Operator](https://external-secrets.io/).

### Why ESO?

| Feature | Sealed Secrets | ESO |
|---------|---------------|-----|
| Backend | Self-contained | Pluggable (Vault, AWS SM, GCP SM, etc.) |
| Dynamic rotation | No | Yes |
| Audit logging | No | Depends on backend |
| Cloud integration | No | Native |

### ESO Architecture

```
┌─────────────────────────────────────────────────────────┐
│  External Backend (Vault / AWS SM / GCP SM / Doppler)   │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼ sync
┌─────────────────────────────────────────────────────────┐
│                    K8s Cluster                           │
│  ┌─────────────────────────────────────────────────┐    │
│  │  External Secrets Operator                       │    │
│  │  ExternalSecret CR ──sync──▶ K8s Secret         │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

### ESO with Free Backend Options

For resource-constrained environments:

1. **Doppler** (Recommended)
   - Free tier: Unlimited secrets
   - Easy setup, good UI
   - `doppler` CLI for local dev

2. **Infisical**
   - Open source, can self-host
   - Free tier: 5 team members

### ESO Migration Steps (Future)

```bash
# 1. Install ESO
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets

# 2. Create SecretStore (example with Doppler)
cat <<EOF | kubectl apply -f -
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: doppler-store
  namespace: revieu
spec:
  provider:
    doppler:
      auth:
        secretRef:
          dopplerToken:
            name: doppler-token
            key: token
EOF

# 3. Create ExternalSecret
cat <<EOF | kubectl apply -f -
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: revieu-secrets
  namespace: revieu
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: doppler-store
    kind: SecretStore
  target:
    name: revieu-secrets
  dataFrom:
    - find:
        name:
          regexp: ".*"
EOF
```

### Learning Resources

- [ESO Documentation](https://external-secrets.io/latest/)
- [Doppler + ESO Guide](https://docs.doppler.com/docs/kubernetes-external-secrets-operator)
- [Vault + ESO Guide](https://external-secrets.io/latest/provider/hashicorp-vault/)
