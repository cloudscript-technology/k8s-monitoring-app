# K8s Monitoring App - Resumo de Endpoints

## 📋 Todos os Endpoints Disponíveis

### Health Check
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/health` | Verifica status da aplicação |

---

### Projects (Projetos)
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/v1/projects` | Listar todos os projetos |
| `GET` | `/api/v1/projects/:id` | Obter projeto por ID |
| `POST` | `/api/v1/projects` | Criar novo projeto |
| `PUT` | `/api/v1/projects/:id` | Atualizar projeto |
| `DELETE` | `/api/v1/projects/:id` | Deletar projeto |

---

### Applications (Aplicações)
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/v1/applications` | Listar todas as aplicações |
| `GET` | `/api/v1/applications/:id` | Obter aplicação por ID |
| `GET` | `/api/v1/projects/:project_id/applications` | Listar aplicações de um projeto |
| `POST` | `/api/v1/applications` | Criar nova aplicação |
| `PUT` | `/api/v1/applications/:id` | Atualizar aplicação |
| `DELETE` | `/api/v1/applications/:id` | Deletar aplicação |

---

### Metric Types (Tipos de Métricas)
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/v1/metric-types` | Listar tipos de métricas disponíveis |
| `GET` | `/api/v1/metric-types/:id` | Obter tipo de métrica por ID |

**Tipos Disponíveis:**
- `HealthCheck` - Health check HTTP
- `PodStatus` - Status do pod
- `PodMemoryUsage` - Uso de memória
- `PodCpuUsage` - Uso de CPU
- `PvcUsage` - Uso de PVC
- `PodActiveNodes` - Nodes ativos

---

### Application Metrics (Configuração de Métricas)
| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/api/v1/application-metrics` | Listar todas as configurações de métricas |
| `GET` | `/api/v1/application-metrics/:id` | Obter configuração por ID |
| `GET` | `/api/v1/applications/:application_id/metrics` | Listar métricas de uma aplicação |
| `POST` | `/api/v1/application-metrics` | Configurar nova métrica |
| `PUT` | `/api/v1/application-metrics/:id` | Atualizar configuração |
| `DELETE` | `/api/v1/application-metrics/:id` | Deletar configuração |

---

### 🆕 Metric Values (Dados Coletados)
| Método | Endpoint | Descrição | Query Params |
|--------|----------|-----------|--------------|
| `GET` | `/api/v1/metric-values/:id` | Obter valor específico | - |
| `GET` | `/api/v1/application-metrics/:application_metric_id/values` | Listar histórico de valores | `limit` (padrão: 100, max: 1000) |
| `GET` | `/api/v1/applications/:application_id/latest-metrics` | Obter últimos valores de todas as métricas | - |

---

## 🎯 Endpoints Principais por Caso de Uso

### Configuração Inicial
```
1. POST /api/v1/projects              → Criar projeto
2. POST /api/v1/applications          → Registrar aplicação
3. GET  /api/v1/metric-types          → Ver tipos disponíveis
4. POST /api/v1/application-metrics   → Configurar métricas
```

### Visualização em Tempo Real
```
GET /api/v1/applications/:id/latest-metrics
→ Retorna valor mais recente de cada métrica
```

### Análise Histórica
```
GET /api/v1/application-metrics/:metric_id/values?limit=100
→ Retorna últimas 100 medições de uma métrica
```

### Dashboard de Monitoramento
```
GET /api/v1/applications               → Listar aplicações
GET /api/v1/applications/:id/latest-metrics  → Status de cada uma
```

---

## 📝 Exemplos Práticos

### 1. Ver Status Atual de uma Aplicação
```bash
curl http://localhost:8080/api/v1/applications/{application_id}/latest-metrics
```

**Resposta (com informações completas):**
```json
{
  "application_id": "uuid",
  "application_name": "web-app",
  "application_description": "Main web application",
  "application_namespace": "production",
  "project_id": "uuid",
  "project_name": "Production",
  "project_description": "Production environment",
  "metrics": [
    {
      "metric_id": "uuid",
      "metric_type_id": "uuid",
      "metric_type_name": "HealthCheck",
      "metric_type_description": "Health check of the pod",
      "configuration": {...},
      "latest_value": {
        "value": {
          "status": "up",
          "response_time_ms": 150
        },
        "created_at": "2024-01-15T10:30:00Z"
      }
    }
  ]
}
```

### 2. Ver Últimas 10 Medições de uma Métrica
```bash
curl "http://localhost:8080/api/v1/application-metrics/{metric_id}/values?limit=10"
```

### 3. Workflow Completo
```bash
# 1. Criar projeto
PROJECT=$(curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"Prod","description":"Production"}' | jq -r '.id')

# 2. Criar aplicação
APP=$(curl -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":\"$PROJECT\",\"name\":\"web\",\"description\":\"Web App\",\"namespace\":\"production\"}" | jq -r '.id')

# 3. Obter tipo HealthCheck
HEALTH_TYPE=$(curl http://localhost:8080/api/v1/metric-types | jq -r '.[] | select(.name=="HealthCheck") | .id')

# 4. Configurar health check
METRIC=$(curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{\"application_id\":\"$APP\",\"type_id\":\"$HEALTH_TYPE\",\"configuration\":{\"health_check_url\":\"http://web.production.svc.cluster.local/health\",\"method\":\"GET\",\"expected_status\":200,\"timeout_seconds\":10}}" | jq -r '.id')

# 5. Aguardar 1-2 minutos

# 6. Ver métricas coletadas
curl "http://localhost:8080/api/v1/applications/$APP/latest-metrics" | jq '.'
```

---

## 🔄 Fluxo de Dados

```
1. Usuário configura métricas via API
   ↓
2. Cron job coleta dados a cada minuto
   ↓
3. Dados armazenados em application_metric_values
   ↓
4. Usuário consulta via endpoints de visualização
```

---

## 💡 Dicas

### Performance
- Use `limit` nos endpoints de histórico para evitar resultados grandes
- O endpoint `latest-metrics` é otimizado para dashboards
- Para análises pesadas, considere consultar direto o banco de dados

### Retenção de Dados
```sql
-- Exemplo: Limpar dados antigos (30 dias)
DELETE FROM application_metric_values 
WHERE created_at < NOW() - INTERVAL '30 days';
```

### Índices Recomendados
```sql
-- Para melhorar performance
CREATE INDEX idx_metric_values_metric_created 
  ON application_metric_values(application_metric_id, created_at DESC);

CREATE INDEX idx_metric_values_created 
  ON application_metric_values(created_at DESC);
```

---

## 🔐 Autenticação

Atualmente a API não possui autenticação. Para produção, considere adicionar:
- Bearer Token
- API Keys
- OAuth2
- JWT

---

## 📊 Estrutura de Resposta dos Valores

### HealthCheck
```json
{
  "status": "up|down",
  "response_time_ms": 150,
  "status_code": 200,
  "error_message": ""
}
```

### PodStatus
```json
{
  "pod_phase": "Running|Pending|Failed",
  "pod_ready": true,
  "restart_count": 0
}
```

### PodMemoryUsage
```json
{
  "memory_usage_bytes": 536870912,
  "memory_limit_bytes": 1073741824,
  "memory_percent": 50.0
}
```

### PodCpuUsage
```json
{
  "cpu_usage_millicores": 250,
  "cpu_limit_millicores": 1000,
  "cpu_percent": 25.0
}
```

### PvcUsage
```json
{
  "pvc_capacity_bytes": 10737418240,
  "pvc_used_bytes": 5368709120,
  "pvc_percent": 50.0
}
```

### PodActiveNodes
```json
{
  "active_nodes_count": 3,
  "node_names": ["node-1", "node-2", "node-3"]
}
```

---

## 📚 Documentação Adicional

- [API Completa](API.md) - Documentação detalhada de cada endpoint
- [Deployment](DEPLOYMENT.md) - Como fazer deploy
- [Exemplos](EXAMPLES.md) - Mais exemplos de uso
- [Local Development](LOCAL_DEVELOPMENT.md) - Desenvolvimento local
- [Postman](../postman/README.md) - Collection do Postman

