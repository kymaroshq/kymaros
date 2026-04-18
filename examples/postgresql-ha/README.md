# PostgreSQL HA — Intermediate Example

A PostgreSQL primary-replica setup using a StatefulSet with streaming replication, validating that both nodes come back healthy after a restore.

## What gets validated

| Check | Type | What it proves |
|-------|------|----------------|
| Primary pod ready | podStatus | Primary postgres-0 starts and passes readiness probe |
| Replica pod ready | podStatus | Replica postgres-1 starts and connects to primary |
| PostgreSQL TCP | tcpSocket | Port 5432 accepts connections |
| pg_isready | exec | `pg_isready` returns 0 — database is accepting queries |
| Credentials and PVCs | resourceExists | Secret and both PVCs survived the restore |

- **Expected score:** 95–100
- **Typical run time:** ~4 minutes
- **RTO target:** 8 minutes

## Deploy

### 1. Create the namespace and secret

```bash
kubectl create namespace kymaros-demo-postgres

kubectl create secret generic postgres-credentials \
  --namespace kymaros-demo-postgres \
  --from-literal=postgres-password="$(openssl rand -base64 16)" \
  --from-literal=replication-password="$(openssl rand -base64 16)"
```

### 2. Deploy the application

```bash
kubectl apply -f app.yaml
```

Wait for both pods to be ready:

```bash
kubectl -n kymaros-demo-postgres wait --for=condition=Ready pod --all --timeout=180s
```

### 3. Create a Velero backup schedule

```bash
kubectl apply -f velero-backup-schedule.yaml
```

Wait for the first backup to complete:

```bash
velero backup get -l velero.io/schedule-name=postgres-ha-demo
```

### 4. Deploy Kymaros resources

```bash
kubectl apply -f kymaros-healthcheckpolicy.yaml
kubectl apply -f kymaros-restoretest.yaml
```

### 5. Watch the results

```bash
kubectl get restoretest postgres-ha-test -n kymaros-demo-postgres
kubectl get restorereport -n kymaros-demo-postgres
```

## Architecture

```
┌────────────────┐     streaming      ┌────────────────┐
│  postgres-0    │───replication────▶│  postgres-1    │
│  (primary)     │                    │  (replica)     │
└────────────────┘                    └────────────────┘
        │
  postgres (headless Service)
```

The primary accepts writes. The replica streams WAL from the primary for read replicas. Both use separate PVCs for data persistence.

## Cleanup

```bash
kubectl delete namespace kymaros-demo-postgres
```
