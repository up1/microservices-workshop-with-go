# Workshop :: Develop microservice with Go
* REST APIs
* Heahth check
* Circuit breaker
* Tracing
* Metric
* Logging
* Pipeline with (Jenkins)

## Step to run with Docker compose
```
$docker-compose build
$docker-compose up -d db
$docker-compose up -d rabbitmq
$docker-compose up -d zipkin
$docker-compose up -d prometheus
$docker-compose up -d grafana
$docker-compose up -d account-service
$docker-compose up -d data-service
$docker-compose up -d report-service
$docker-compose ps
$docker-compose logs --follow
```

URL for testing
* Call from account service :: http://localhost:8787/accounts/1
* Call from data service :: http://localhost:8787/accounts/1
* Metrics
  * http://localhost:6767/metrics
  * http://localhost:8787/metrics
  * http://localhost:9797/metrics
  