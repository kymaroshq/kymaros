# Microservices App — Advanced Example

A multi-namespace application with an API gateway, PostgreSQL database, and Redis cache spread across three namespaces. This example demonstrates Kymaros cross-namespace dependency validation.

## What gets validated

| Check | Type | What it proves |
|-------|------|----------------|
| API pod ready | podStatus | API gateway starts in its namespace |
| Postgres pod ready | podStatus | Database starts in its namespace |
| Redis pod ready | podStatus | Cache starts in its namespace |
| API HTTP health | httpGet | API responds 200 on `/healthz` |
| Postgres TCP | tcpSocket | Database accepts connections cross-namespace |
| Redis TCP | tcpSocket | Cache accepts connections cross-namespace |
| Cross-namespace DNS | exec | API can resolve `postgres.kymaros-demo-ms-db` and `redis.kymaros-demo-ms-cache` |
| Critical resources | resourceExists | Secrets and ConfigMaps present in all namespaces |

- **Expected score:** 90–100
- **Typical run time:** ~5 minutes
- **RTO target:** 10 minutes

## Architecture

```
 kymaros-demo-ms-api        kymaros-demo-ms-db        kymaros-demo-ms-cache
┌───────────────────┐    ┌───────────────────┐    ┌───────────────────┐
│                   │    │                   │    │                   │
│   nginx (API)     │───▶│   postgres:16     │    │   redis:7         │
│                   │    │                   │    │                   │
│                   │────────────────────────────▶│                   │
│                   │    │                   │    │                   │
└───────────────────┘    └───────────────────┘    └───────────────────┘
```

Three namespaces communicate via cross-namespace DNS (`service.namespace.svc.cluster.local`). The sandbox uses `networkIsolation: "group"` to allow intra-group traffic while blocking everything else.

## Deploy

### 1. Create namespaces and secrets

```bash
kubectl create namespace kymaros-demo-ms-api
kubectl create namespace kymaros-demo-ms-db
kubectl create namespace kymaros-demo-ms-cache

kubectl create secret generic db-credentials \
  --namespace kymaros-demo-ms-db \
  --from-literal=postgres-password="$(openssl rand -base64 16)"

# The API needs the database password too
kubectl create secret generic db-credentials \
  --namespace kymaros-demo-ms-api \
  --from-literal=postgres-password="$(kubectl get secret db-credentials -n kymaros-demo-ms-db -o jsonpath='{.data.postgres-password}' | base64 -d)"
```

### 2. Deploy the application

```bash
kubectl apply -f app.yaml
```

Wait for all pods across namespaces:

```bash
kubectl -n kymaros-demo-ms-api wait --for=condition=Ready pod --all --timeout=120s
kubectl -n kymaros-demo-ms-db wait --for=condition=Ready pod --all --timeout=120s
kubectl -n kymaros-demo-ms-cache wait --for=condition=Ready pod --all --timeout=120s
```

### 3. Create a Velero backup schedule

```bash
kubectl apply -f velero-backup-schedule.yaml
```

Wait for the first backup:

```bash
velero backup get -l velero.io/schedule-name=microservices-demo
```

### 4. Deploy Kymaros resources

The RestoreTest and HealthCheckPolicy are deployed in the API namespace:

```bash
kubectl apply -f kymaros-healthcheckpolicy.yaml
kubectl apply -f kymaros-restoretest.yaml
```

### 5. Watch the results

```bash
kubectl get restoretest microservices-test -n kymaros-demo-ms-api
kubectl get restorereport -n kymaros-demo-ms-api
```

## Key configuration

The RestoreTest uses two important settings for multi-namespace restores:

- **`backupSource.namespaces`** lists all three namespaces — Kymaros restores them together as a group
- **`sandbox.networkIsolation: "group"`** — allows traffic between the sandbox namespaces while still blocking external traffic

Without `"group"` mode, the sandboxed API would be unable to reach the sandboxed database, and the cross-namespace health checks would fail.

## Cleanup

```bash
kubectl delete namespace kymaros-demo-ms-api kymaros-demo-ms-db kymaros-demo-ms-cache
```
