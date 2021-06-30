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
$docker-compose up -d rabbitmq
$docker-compose up -d zipkin
$docker-compose up -d prometheus
$docker-compose up -d grafana
$docker-compose up -d account-service
$docker-compose ps
$docker-compose logs --follow
```