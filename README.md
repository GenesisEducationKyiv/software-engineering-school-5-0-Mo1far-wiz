# Some kind of weather notification service

## Setup
_if you want to use linter you should install it before_
```cmd
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
```
[more info see here](https://golangci-lint.run/welcome/install/)

## To run
_don't forget running docker before_
```cmd
make up
```

## About
I used Gin, SQL and migrate.

Also I have added postman collection for tests.

## Postman
You can access postman collection for tests [on this url](https://www.postman.com/avionics-operator-63001856/workspace/genesis-weather).

## Monitoring & Alerts

We rely on both structured logging and Prometheus metrics to drive our alerting. Below are the key alerts we would configure, along with the rationale for each:

### 1. High Error Rate  
- **Metric:** `app_errors_total`  
- **Alert Rule:** Trigger if the error counter increases by more than 5 errors/minute over a 5-minute window.  
- **Why:** A sudden spike in application errors usually indicates a bug or downstream degradation (e.g. RabbitMQ, database, third-party API) that needs immediate investigation.

### 2. Elevated Latency  
- **Metric:** `app_request_latency_seconds` (95th percentile)  
- **Alert Rule:** Trigger if the 95th percentile latency exceeds 1 second for more than 5 minutes.  
- **Why:** Sustained slow responses impact user experience and can signal resource exhaustion or network issues.

### 3. Failed Scheduled Batch Jobs  
- **Metric:** `app_requests_total{service="email_batch_daily",status="error"}` and similarly for `"email_batch_hourly"`  
- **Alert Rule:** Trigger if any daily or hourly batch job fails in two consecutive runs.  
- **Why:** Ensures that weather‐email dispatch jobs are running reliably—missed batches mean subscribers don’t get their updates.

### 4. Subscription Service Errors  
- **Metric:** `app_subscription_requests_total{method="confirm",status="error"}` (and `subscribe`, `unsubscribe`)  
- **Alert Rule:** Trigger if confirm errors exceed 1% of total confirms over 15 minutes.  
- **Why:** Failures in the subscription workflow directly block new users or unsubscribes, leading to incorrect subscriber counts and bad UX.

### 5. High Cache Miss Rate  
- **Metric:** `app_cache_ops_total{operation="get",status="miss"}` vs. `{operation="get",status="success"}`  
- **Alert Rule:** Trigger if miss rate > 20% over 10 minutes.  
- **Why:** A rising cache miss rate increases load on Redis or the weather API and indicates stale or improperly warmed cache data.

### 6. Prometheus Scrape Failure  
- **Target:** The `/metrics` endpoint on port 9090  
- **Alert Rule:** Trigger if Prometheus has not scraped the service in > 2 minutes.  
- **Why:** Guarantees that monitoring itself is healthy; if scrapes stop, none of the above alerts will fire.

### 7. Disk / Log Path Usage  
- **Metric:** Host disk usage on the `weather.log` directory  
- **Alert Rule:** Trigger if disk usage exceeds 80%.  
- **Why:** Prevents log writes from failing due to full disks, which could hide critical errors.

---

## Log Retention Policy

We use Zap with Lumberjack rotation and define retention periods according to log volume and severity:

| Log Level | Retention   | Rotation & Archival                         | Rationale                                          |
|-----------|-------------|---------------------------------------------|----------------------------------------------------|
| **DEBUG** | 7 days      | Rotate daily; purge after 7 days            | High‐volume developmental logs, only useful short-term |
| **INFO**  | 28 days     | Rotate when file > 100 MB; keep 7 backups; purge after 28 days | Operational visibility into normal flows           |
| **WARN**  | 28 days     | Same as INFO                                | Capture warnings for troubleshooting recent issues |
| **ERROR** | 90 days     | Rotate monthly into compressed archives; retain on S3 | Critical failures need long-term audit trail       |

- **Rotation Settings** (Lumberjack):  
  - `MaxSize: 100 MB`, `MaxBackups: 7`, `MaxAge: 28 days`, `Compress: true`.  
- **Archive Process:**  
  - Once an “ERROR” log file is older than 28 days, compress and move to cold storage (e.g. AWS S3 Glacier).  
  - INFO/WARN files older than 28 days are deleted automatically.

> **Why these choices?**  
> - **Short DEBUG retention** keeps disk usage low.  
> - **Month-long INFO/WARN retention** covers most operational troubleshooting windows.  
> - **Extended ERROR retention** meets compliance and forensic analysis needs for critical incidents.  
> - **Automated rotation & compression** prevents manual intervention and ensures we never run out of disk space.  
