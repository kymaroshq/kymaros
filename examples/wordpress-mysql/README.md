# WordPress + MySQL — Beginner Example

A classic WordPress deployment backed by MySQL, validating that a full restore brings both the database and the web frontend back to a working state.

## What gets validated

| Check | Type | What it proves |
|-------|------|----------------|
| MySQL pod ready | podStatus | Database container starts and passes readiness probe |
| WordPress pod ready | podStatus | Web server starts and connects to the database |
| WordPress login page | httpGet | HTTP 200 on `/wp-login.php` — app is fully functional |
| Database credentials | resourceExists | Secret `wordpress-mysql-credentials` survived the restore |
| MySQL data volume | resourceExists | PVC `mysql-data` is bound and accessible |

- **Expected score:** 95–100
- **Typical run time:** ~3 minutes
- **RTO target:** 5 minutes

## Deploy

### 1. Create the namespace and secret

```bash
kubectl create namespace kymaros-demo-wordpress

kubectl create secret generic wordpress-mysql-credentials \
  --namespace kymaros-demo-wordpress \
  --from-literal=mysql-root-password="$(openssl rand -base64 16)" \
  --from-literal=mysql-password="$(openssl rand -base64 16)"
```

### 2. Deploy the application

```bash
kubectl apply -f app.yaml
```

Wait for all pods to be ready:

```bash
kubectl -n kymaros-demo-wordpress wait --for=condition=Ready pod --all --timeout=120s
```

### 3. Create a Velero backup schedule

```bash
kubectl apply -f velero-backup-schedule.yaml
```

Wait for the first backup to complete:

```bash
velero backup get -l velero.io/schedule-name=wordpress-demo
```

### 4. Deploy Kymaros resources

```bash
kubectl apply -f kymaros-healthcheckpolicy.yaml
kubectl apply -f kymaros-restoretest.yaml
```

### 5. Watch the results

Wait for the scheduled run or check the RestoreTest status:

```bash
kubectl get restoretest wordpress-mysql-test -n kymaros-demo-wordpress
kubectl get restorereport -n kymaros-demo-wordpress
```

## Failure scenarios

Try rotating the MySQL secret after a backup is taken. On the next restore test, WordPress will fail to connect to MySQL and the score will drop to ~60 (pod startup passes but HTTP check fails). This demonstrates how Kymaros catches real restore issues that a simple "backup completed" status would miss.

## Cleanup

```bash
kubectl delete namespace kymaros-demo-wordpress
```
