# Redis Sentinel — Intermediate Example

A Redis master-replica setup with Sentinel for automatic failover, validating that all Redis and Sentinel pods recover correctly after a restore.

## What gets validated

| Check | Type | What it proves |
|-------|------|----------------|
| Redis pods ready | podStatus | All 3 Redis pods (master + 2 replicas) start |
| Sentinel pods ready | podStatus | All 3 Sentinel pods start and discover the master |
| Redis TCP | tcpSocket | Port 6379 accepts connections |
| Sentinel TCP | tcpSocket | Port 26379 accepts connections |
| Redis PING | exec | `redis-cli ping` returns PONG |
| Sentinel masters | exec | Sentinel knows the current master |

- **Expected score:** 95–100
- **Typical run time:** ~3 minutes
- **RTO target:** 6 minutes

## Deploy

### 1. Create the namespace

```bash
kubectl create namespace kymaros-demo-redis
```

### 2. Deploy the application

```bash
kubectl apply -f app.yaml
```

Wait for all pods to be ready:

```bash
kubectl -n kymaros-demo-redis wait --for=condition=Ready pod --all --timeout=180s
```

### 3. Create a Velero backup schedule

```bash
kubectl apply -f velero-backup-schedule.yaml
```

Wait for the first backup to complete:

```bash
velero backup get -l velero.io/schedule-name=redis-sentinel-demo
```

### 4. Deploy Kymaros resources

```bash
kubectl apply -f kymaros-healthcheckpolicy.yaml
kubectl apply -f kymaros-restoretest.yaml
```

### 5. Watch the results

```bash
kubectl get restoretest redis-sentinel-test -n kymaros-demo-redis
kubectl get restorereport -n kymaros-demo-redis
```

## Architecture

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ sentinel-0  │  │ sentinel-1  │  │ sentinel-2  │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────────────┼────────────────┘
                        │ monitors
       ┌────────────────┼────────────────┐
┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐
│  redis-0    │  │  redis-1    │  │  redis-2    │
│  (master)   │  │  (replica)  │  │  (replica)  │
└─────────────┘  └─────────────┘  └─────────────┘
```

Sentinel monitors the Redis pods and triggers automatic failover if the master goes down. After a restore, Sentinel must re-discover the topology — this is exactly what Kymaros validates.

## Cleanup

```bash
kubectl delete namespace kymaros-demo-redis
```
