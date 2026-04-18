# DORA Compliance — Advanced Example

Template for generating restore validation evidence that satisfies DORA (Digital Operational Resilience Act) Article 11 requirements.

This is not a standalone application — it is a RestoreTest and HealthCheckPolicy template you apply to your existing production workloads.

## DORA Article 11.6 and Kubernetes backups

DORA Article 11.6 requires financial entities to:

> "periodically test the backup and restoration procedures and verify that the backup data is fully and timely recoverable"

In practice, this means:

1. **Periodic testing** — automated, scheduled restore tests (not manual quarterly drills)
2. **Full recoverability** — proving the backup produces a working application, not just "restore completed"
3. **Timely recovery** — measuring actual RTO against your declared recovery targets
4. **Evidence** — auditable records showing test date, result, and measured RTO

Kymaros generates all four automatically via RestoreReport CRDs.

## How it works

The RestoreTest runs daily at 3:00 AM (outside business hours) and validates your production backup against a comprehensive HealthCheckPolicy. Each run produces a RestoreReport that serves as an audit artifact.

The Grafana panel provides a quarterly compliance dashboard showing:
- Total tests run
- Success rate (must be 100% for audit)
- Average RTO vs your DORA target
- Trend over time

## Deploy

### 1. Adapt the templates to your workload

Edit `kymaros-restoretest.yaml`:
- Replace `your-production-namespace` with your actual namespace
- Adjust `sla.maxRTO` to your declared RTO target
- Update the schedule if needed (default: daily 3:00 AM UTC)

Edit `kymaros-healthcheckpolicy.yaml`:
- Replace the example checks with checks relevant to your application
- Ensure you cover pod startup, connectivity, and critical resources

### 2. Apply the resources

```bash
kubectl apply -f kymaros-healthcheckpolicy.yaml
kubectl apply -f kymaros-restoretest.yaml
```

### 3. Import the Grafana dashboard

Import `grafana-compliance-panel.json` into Grafana:
- Go to Dashboards > Import
- Upload the JSON file or paste its contents
- Select your Prometheus data source

### 4. Generate a quarterly report

Query Prometheus to extract the compliance summary for a given quarter:

```promql
# Total tests run this quarter
increase(kymaros_tests_total{test="dora-compliance-test"}[90d])

# Success rate (should be 100%)
avg_over_time(
  (kymaros_score{test="dora-compliance-test"} >= 90)[90d:]
)

# Average RTO vs target
avg_over_time(kymaros_rto_seconds{test="dora-compliance-test"}[90d])
```

Or list all RestoreReports for the audit period:

```bash
kubectl get restorereport -l kymaros.io/compliance=dora \
  --sort-by=.status.timing.startedAt \
  -o custom-columns=\
NAME:.metadata.name,\
SCORE:.status.score,\
RESULT:.status.result,\
RTO:.status.rto,\
DATE:.status.timing.startedAt
```

## Audit checklist

Use this checklist when preparing for a DORA compliance audit:

- [ ] RestoreTest runs daily (verify `kubectl get rt dora-compliance-test`)
- [ ] All RestoreReports for the audit period show `result: pass`
- [ ] Measured RTO is consistently below declared maxRTO
- [ ] Grafana dashboard shows 100% success rate for the quarter
- [ ] Export RestoreReports as evidence artifacts

## Cleanup

```bash
kubectl delete restoretest dora-compliance-test
kubectl delete healthcheckpolicy dora-compliance-checks
```
