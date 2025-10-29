# Guia de Cadastros com Exemplos

Este guia mostra como cadastrar Projetos, Aplicações e cada tipo de Métrica, com exemplos práticos via API (cURL) e via importação YAML pela UI.

## Visão Geral

- UI de importação YAML: `GET /cadastros/importacao` (menu Cadastros → Importação YAML)
- Listar tipos de métricas: `GET /api/v1/metric-types`
- Cadastrar projeto: `POST /api/v1/projects`
- Cadastrar aplicação: `POST /api/v1/applications`
- Cadastrar métrica de aplicação: `POST /api/v1/application-metrics`

Substitua `http://localhost:8080` conforme seu ambiente.

---

## 1) Cadastrar Projeto

### Via API (cURL)

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "k8s-monitoring-app",
    "description": "Projeto principal do monitoramento"
  }'
```

### Via YAML (UI Importação)

```yaml
kind: Project
metadata:
  name: k8s-monitoring-app
  description: "Projeto principal do monitoramento"
```

---

## 2) Cadastrar Aplicação

Campos obrigatórios: `name`, `project` (nome do projeto), `namespace`.

### Via API (cURL)

```bash
# Obter ID do projeto pelo nome (se necessário)
# ou use já o ID que você tem

PROJECT_ID=$(curl -s http://localhost:8080/api/v1/projects | jq -r '.[] | select(.name=="k8s-monitoring-app") | .id')

curl -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"name\": \"app1\",
    \"description\": \"Aplicação 1\",
    \"namespace\": \"cluster-monitoring\"
  }"
```

### Via YAML (UI Importação)

```yaml
kind: Application
metadata:
  name: app1
  description: "Aplicação 1"
  namespace: cluster-monitoring
  project: k8s-monitoring-app
```

---

## 3) Tipos de Métrica e Exemplos

Antes de cadastrar, descubra o `type_id` do tipo de métrica (para uso via API) ou use o `metricType` pelo nome (para YAML).

### Listar Tipos

```bash
curl http://localhost:8080/api/v1/metric-types | jq '.'
```

Procure pelos nomes a seguir e anote o `id`:
- HealthCheck
- PodStatus
- PodMemoryUsage
- PodCpuUsage
- PvcUsage
- PodActiveNodes
- IngressCertificate
- RedisConnection
- PostgreSQLConnection
- MongoDBConnection
- MySQLConnection
- KongConnection
- KafkaConsumerLag

Nos exemplos de cURL, substitua `APPLICATION_ID` e `TYPE_ID` pelos valores reais.

### 3.1) HealthCheck

Campos principais: `health_check_url`, `method`, `expected_status`, `timeout_seconds`.

- Via API (cURL):
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "health_check_url": "http://app1.cluster-monitoring.svc.cluster.local/health",
      "method": "GET",
      "expected_status": 200,
      "timeout_seconds": 5
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: HealthCheck
  configuration:
    health_check_url: http://app1.cluster-monitoring.svc.cluster.local/health
    method: GET
    expected_status: 200
    timeout_seconds: 5
```

### 3.2) PodStatus

Campos: `pod_label_selector`, `container_name` (opcional).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "pod_label_selector": "app=app1",
      "container_name": "main"
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: PodStatus
  configuration:
    pod_label_selector: "app=app1"
    container_name: "main"
```

### 3.3) PodMemoryUsage

Campos: `pod_label_selector`, `container_name` (opcional).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "pod_label_selector": "app=app1",
      "container_name": "main"
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: PodMemoryUsage
  configuration:
    pod_label_selector: "app=app1"
    container_name: "main"
```

### 3.4) PodCpuUsage

Campos: `pod_label_selector`, `container_name` (opcional).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "pod_label_selector": "app=app1",
      "container_name": "main"
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: PodCpuUsage
  configuration:
    pod_label_selector: "app=app1"
    container_name: "main"
```

### 3.5) PvcUsage

Campos: `pvc_name` (obrigatório), `pod_label_selector` (obrigatório), `container_name` (opcional), `pvc_mount_path` (opcional).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "pvc_name": "data-pvc",
      "pod_label_selector": "app=app1",
      "container_name": "main",
      "pvc_mount_path": "/data"
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: PvcUsage
  configuration:
    pvc_name: data-pvc
    pod_label_selector: "app=app1"
    container_name: "main"
    pvc_mount_path: "/data"
```

### 3.6) PodActiveNodes

Campos: `pod_label_selector`.

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "pod_label_selector": "app=app1"
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: PodActiveNodes
  configuration:
    pod_label_selector: "app=app1"
```

### 3.7) IngressCertificate

Campos: `ingress_name` (obrigatório), `ingress_namespace` (opcional; padrão: namespace da aplicação), `tls_secret_name` (opcional; auto-descoberto), `warning_days` (opcional; padrão: 30).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "ingress_name": "my-app-ingress",
      "warning_days": 30
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: IngressCertificate
  configuration:
    ingress_name: my-app-ingress
    warning_days: 30
```

### 3.8) RedisConnection

Campos comuns de conexão: `connection_host`, `connection_port`, `connection_username` (se necessário), `connection_password` (se necessário), `connection_ssl` (bool, opcional), `connection_timeout` (int, opcional). Específico de Redis: `connection_db` (int, opcional; default 0).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "connection_host": "redis.default.svc.cluster.local",
      "connection_port": 6379,
      "connection_password": "senha-redis",
      "connection_db": 0,
      "connection_timeout": 5
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: RedisConnection
  configuration:
    connection_host: redis.default.svc.cluster.local
    connection_port: 6379
    connection_password: senha-redis
    connection_db: 0
    connection_timeout: 5
```

### 3.9) PostgreSQLConnection

Campos comuns de conexão + `connection_database` (string).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "connection_host": "postgres.default.svc.cluster.local",
      "connection_port": 5432,
      "connection_username": "monitor",
      "connection_password": "senha",
      "connection_database": "appdb",
      "connection_ssl": false,
      "connection_timeout": 5
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: PostgreSQLConnection
  configuration:
    connection_host: postgres.default.svc.cluster.local
    connection_port: 5432
    connection_username: monitor
    connection_password: senha
    connection_database: appdb
    connection_ssl: false
    connection_timeout: 5
```

### 3.10) MongoDBConnection

Campos comuns de conexão + `connection_database` e, opcionalmente, `connection_auth_source`.

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "connection_host": "mongodb.default.svc.cluster.local",
      "connection_port": 27017,
      "connection_username": "monitor",
      "connection_password": "senha",
      "connection_database": "appdb",
      "connection_auth_source": "admin",
      "connection_timeout": 5
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: MongoDBConnection
  configuration:
    connection_host: mongodb.default.svc.cluster.local
    connection_port: 27017
    connection_username: monitor
    connection_password: senha
    connection_database: appdb
    connection_auth_source: admin
    connection_timeout: 5
```

### 3.11) MySQLConnection

Campos comuns de conexão + `connection_database`.

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "connection_host": "mysql.default.svc.cluster.local",
      "connection_port": 3306,
      "connection_username": "monitor",
      "connection_password": "senha",
      "connection_database": "appdb",
      "connection_timeout": 5
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: MySQLConnection
  configuration:
    connection_host: mysql.default.svc.cluster.local
    connection_port: 3306
    connection_username: monitor
    connection_password: senha
    connection_database: appdb
    connection_timeout: 5
```

### 3.12) KongConnection

Campos comuns de conexão. Opcional: `kong_admin_url` (se for necessário checar a API Admin).

- Via API:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "connection_host": "kong.default.svc.cluster.local",
      "connection_port": 8001,
      "connection_username": "admin",
      "connection_password": "senha",
      "connection_timeout": 5,
      "kong_admin_url": "http://kong.default.svc.cluster.local:8001"
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: KongConnection
  configuration:
    connection_host: kong.default.svc.cluster.local
    connection_port: 8001
    connection_username: admin
    connection_password: senha
    connection_timeout: 5
    kong_admin_url: http://kong.default.svc.cluster.local:8001
```

### 3.13) KafkaConsumerLag

Campos: `kafka_bootstrap_servers` (obrigatório), `kafka_consumer_group` (obrigatório), `kafka_topic` (opcional), `kafka_security_protocol` (opcional: `PLAINTEXT`, `SASL_PLAINTEXT`, `SASL_SSL`, `SSL`), `kafka_sasl_mechanism` (opcional: `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`), `kafka_sasl_username`/`kafka_sasl_password` (opcionais), `kafka_lag_threshold` (int, opcional; default: 1000).

- Via API (básico):
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "kafka_bootstrap_servers": "kafka:9092",
      "kafka_consumer_group": "my-consumer-group",
      "kafka_lag_threshold": 1000
    }
  }'
```

- Via API (com SASL):
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID",
    "type_id": "TYPE_ID",
    "configuration": {
      "kafka_bootstrap_servers": "kafka.production.svc.cluster.local:9092",
      "kafka_consumer_group": "order-processor",
      "kafka_topic": "orders",
      "kafka_security_protocol": "SASL_PLAINTEXT",
      "kafka_sasl_mechanism": "SCRAM-SHA-256",
      "kafka_sasl_username": "consumer-user",
      "kafka_sasl_password": "consumer-password",
      "kafka_lag_threshold": 500
    }
  }'
```

- Via YAML:
```yaml
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: KafkaConsumerLag
  configuration:
    kafka_bootstrap_servers: kafka:9092
    kafka_consumer_group: my-consumer-group
    kafka_lag_threshold: 1000
```

---

## 4) Importação YAML com Múltiplos Documentos

Você pode colar vários documentos YAML separados por `---` na página de Importação YAML.

Exemplo completo:

```yaml
kind: Project
metadata:
  name: k8s-monitoring-app
  description: "K8s Monitoring App"
---
kind: Application
metadata:
  name: app1
  description: "App1"
  namespace: cluster-monitoring
  project: k8s-monitoring-app
---
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: HealthCheck
  configuration:
    health_check_url: http://app1.cluster-monitoring.svc.cluster.local/health
    expected_status: 200
    timeout_seconds: 5
    method: GET
---
kind: ApplicationMetric
metadata:
  application: app1
  project: k8s-monitoring-app
  metricType: PodStatus
  configuration:
    pod_label_selector: "app=app1"
```

---

## 5) Dicas e Observações

- O sistema impede duplicar o mesmo tipo de métrica na mesma aplicação.
- Campos sensíveis (senhas) são redigidos quando exibidos na UI.
- Em YAML, chaves podem ser camelCase ou snake_case; o importador normaliza para a estrutura interna.
- Para conexão com bancos/serviços, verifique `docs/CONNECTION_METRICS.md` para detalhes completos dos campos e comportamento.
- Para certificados, verifique `docs/INGRESS_CERTIFICATE_EXAMPLE.md` para casos avançados (cross-namespace, secret explícito).

