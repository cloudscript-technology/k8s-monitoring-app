# Novos Endpoints de Visualização de Métricas

## Endpoints Implementados

### 1. Obter Valor de Métrica por ID
```
GET /api/v1/metric-values/:id
```

### 2. Listar Valores de uma Métrica Específica
```
GET /api/v1/application-metrics/:application_metric_id/values?limit=100
```

**Query Parameters:**
- `limit` (opcional) - Número de registros a retornar (padrão: 100, máximo: 1000)

### 3. Obter Últimas Métricas de uma Aplicação
```
GET /api/v1/applications/:application_id/latest-metrics
```

Este endpoint retorna o valor mais recente coletado para cada métrica configurada na aplicação.

## Exemplos de Uso com cURL

### Exemplo 1: Ver últimas métricas de uma aplicação
```bash
# Substitua APPLICATION_ID pelo ID da sua aplicação
curl http://localhost:8080/api/v1/applications/APPLICATION_ID/latest-metrics | jq '.'
```

**Resposta Esperada:**
```json
[
  {
    "metric_id": "550e8400-e29b-41d4-a716-446655440000",
    "metric_type_id": "750e8400-e29b-41d4-a716-446655440002",
    "configuration": {
      "health_check_url": "http://web-app.production.svc.cluster.local:8080/health",
      "method": "GET",
      "expected_status": 200,
      "timeout_seconds": 10
    },
    "latest_value": {
      "id": "850e8400-e29b-41d4-a716-446655440003",
      "application_metric_id": "550e8400-e29b-41d4-a716-446655440000",
      "value": {
        "status": "up",
        "response_time_ms": 150,
        "status_code": 200,
        "error_message": ""
      },
      "created_at": "2024-01-15T10:30:00Z"
    }
  }
]
```

### Exemplo 2: Ver histórico de uma métrica específica
```bash
# Substitua METRIC_ID pelo ID da métrica configurada
curl "http://localhost:8080/api/v1/application-metrics/METRIC_ID/values?limit=10" | jq '.'
```

**Resposta Esperada:**
```json
[
  {
    "id": "uuid-1",
    "application_metric_id": "METRIC_ID",
    "value": {
      "status": "up",
      "response_time_ms": 150,
      "status_code": 200
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "uuid-2",
    "application_metric_id": "METRIC_ID",
    "value": {
      "status": "up",
      "response_time_ms": 145,
      "status_code": 200
    },
    "created_at": "2024-01-15T10:29:00Z"
  }
]
```

### Exemplo 3: Ver um valor específico
```bash
# Substitua VALUE_ID pelo ID do valor
curl http://localhost:8080/api/v1/metric-values/VALUE_ID | jq '.'
```

## Adicionar no Postman

Para adicionar esses endpoints na sua collection do Postman, crie uma nova pasta chamada "Metric Values" com os seguintes requests:

### Request 1: Get Metric Value by ID
```
GET {{base_url}}/api/v1/metric-values/{{metric_value_id}}
```

### Request 2: List Metric Values
```
GET {{base_url}}/api/v1/application-metrics/{{application_metric_id}}/values?limit=100
```

### Request 3: Get Latest Metrics for Application
```
GET {{base_url}}/api/v1/applications/{{application_id}}/latest-metrics
```

## Workflow Completo de Teste

```bash
# 1. Criar projeto
PROJECT_ID=$(curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","description":"Test project"}' | jq -r '.id')

# 2. Criar aplicação
APP_ID=$(curl -s -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":\"$PROJECT_ID\",\"name\":\"test-app\",\"description\":\"Test\",\"namespace\":\"default\"}" | jq -r '.id')

# 3. Obter ID do tipo HealthCheck
HEALTH_ID=$(curl -s http://localhost:8080/api/v1/metric-types | jq -r '.[] | select(.name=="HealthCheck") | .id')

# 4. Configurar health check
METRIC_ID=$(curl -s -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{\"application_id\":\"$APP_ID\",\"type_id\":\"$HEALTH_ID\",\"configuration\":{\"health_check_url\":\"http://httpbin.org/status/200\",\"method\":\"GET\",\"expected_status\":200,\"timeout_seconds\":10}}" | jq -r '.id')

# 5. Aguardar 1-2 minutos para o cron coletar dados

# 6. Ver últimas métricas
curl -s "http://localhost:8080/api/v1/applications/$APP_ID/latest-metrics" | jq '.'

# 7. Ver histórico de uma métrica
curl -s "http://localhost:8080/api/v1/application-metrics/$METRIC_ID/values?limit=5" | jq '.'
```

## Casos de Uso

### Dashboard em Tempo Real
Use o endpoint `latest-metrics` para criar um dashboard mostrando o status atual de todas as aplicações:

```bash
# Listar todas as aplicações
curl -s http://localhost:8080/api/v1/applications | jq -r '.[] | .id' | while read app_id; do
  echo "Application: $app_id"
  curl -s "http://localhost:8080/api/v1/applications/$app_id/latest-metrics" | jq '.'
done
```

### Análise de Tendências
Use o endpoint `values` com limit alto para analisar tendências ao longo do tempo:

```bash
# Últimas 100 medições de uma métrica
curl -s "http://localhost:8080/api/v1/application-metrics/$METRIC_ID/values?limit=100" \
  | jq '.[] | {time: .created_at, status: .value.status, response_time: .value.response_time_ms}'
```

### Monitoramento de SLA
Verifique se o serviço está cumprindo SLAs:

```bash
# Verificar uptime nas últimas medições
curl -s "http://localhost:8080/api/v1/application-metrics/$METRIC_ID/values?limit=100" \
  | jq '[.[] | .value.status] | group_by(.) | map({status: .[0], count: length})'
```

## Integração com Ferramentas de Visualização

### Grafana
Os dados podem ser consultados via queries SQL direto no PostgreSQL:

```sql
-- Últimas 24h de métricas
SELECT 
  a.name as application,
  mt.name as metric_type,
  amv.value,
  amv.created_at
FROM application_metric_values amv
JOIN application_metrics am ON amv.application_metric_id = am.id
JOIN applications a ON am.application_id = a.id
JOIN metric_types mt ON am.type_id = mt.id
WHERE amv.created_at > NOW() - INTERVAL '24 hours'
ORDER BY amv.created_at DESC;
```

### Prometheus (Futuro)
Estes endpoints podem ser usados como base para exportar métricas no formato Prometheus.

## Limitações e Retenção

- Por padrão, não há limite de retenção de dados
- Considere implementar uma política de retenção (ex: manter apenas últimos 30 dias)
- Para volumes grandes, use paginação via limit

## Performance

- O endpoint `latest-metrics` faz uma query por métrica configurada
- Para aplicações com muitas métricas, considere adicionar cache
- Use índices no banco de dados para melhorar performance:

```sql
CREATE INDEX idx_metric_values_metric_id_created ON application_metric_values(application_metric_id, created_at DESC);
```

