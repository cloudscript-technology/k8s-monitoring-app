# Quick Start - Métricas de Conexão

Guia rápido para começar a usar as métricas de conexão de banco de dados em 5 minutos.

## 🚀 Início Rápido

### Passo 1: Executar Migrações

```bash
# A aplicação executa migrações automaticamente ao iniciar
# Ou execute manualmente:
docker exec k8s-monitoring-app-postgres psql -U monitoring k8s_monitoring -f /migrations/1730000000_add_connection_metrics.up.sql
```

### Passo 2: Listar Tipos de Métricas Disponíveis

```bash
curl http://localhost:8080/api/v1/metric-types | jq '.[] | select(.name | contains("Connection"))'
```

Você verá algo como:
```json
{
  "id": "uuid-1",
  "name": "RedisConnection",
  "description": "Test Redis connection with authentication"
},
{
  "id": "uuid-2",
  "name": "PostgreSQLConnection",
  "description": "Test PostgreSQL database connection with authentication"
},
...
```

### Passo 3: Criar uma Métrica de Conexão

#### Exemplo: PostgreSQL

```bash
# Salvar o ID do tipo PostgreSQLConnection
POSTGRES_TYPE_ID="uuid-aqui"

# Criar métrica
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "sua-aplicacao-id",
    "type_id": "'$POSTGRES_TYPE_ID'",
    "configuration": {
      "connection_host": "postgres.default.svc.cluster.local",
      "connection_port": 5432,
      "connection_username": "myuser",
      "connection_password": "mypassword",
      "connection_database": "mydb",
      "connection_timeout": 5
    }
  }'
```

### Passo 4: Aguardar Coleta

As métricas são coletadas automaticamente a cada 60 segundos. Aguarde ~65 segundos.

### Passo 5: Consultar Resultados

```bash
curl http://localhost:8080/api/v1/application-metric-values/application/sua-aplicacao-id/latest | jq
```

**Resultado esperado:**
```json
[
  {
    "id": "value-id",
    "application_metric_id": "metric-id",
    "value": {
      "connection_status": "connected",
      "connection_time_ms": 45,
      "connection_ping_time_ms": 12,
      "connection_version": "PostgreSQL 14.5",
      "connection_info": "Database size: 52428800 bytes"
    },
    "created_at": "2024-10-27T10:30:00Z"
  }
]
```

## 📋 Exemplos Rápidos por Tipo

### Redis
```json
{
  "configuration": {
    "connection_host": "redis.default.svc.cluster.local",
    "connection_port": 6379,
    "connection_password": "redis-pass",
    "connection_db": 0
  }
}
```

### MongoDB
```json
{
  "configuration": {
    "connection_host": "mongodb.default.svc.cluster.local",
    "connection_port": 27017,
    "connection_username": "admin",
    "connection_password": "admin-pass",
    "connection_database": "mydb",
    "connection_auth_source": "admin"
  }
}
```

### MySQL
```json
{
  "configuration": {
    "connection_host": "mysql.default.svc.cluster.local",
    "connection_port": 3306,
    "connection_username": "root",
    "connection_password": "root-pass",
    "connection_database": "mydb"
  }
}
```

### Kong
```json
{
  "configuration": {
    "connection_host": "kong-admin.default.svc.cluster.local",
    "connection_port": 8001
  }
}
```

## 🎯 Script Interativo

Para um setup guiado:

```bash
./examples/connection-metrics-test.sh
```

O script irá:
1. ✅ Criar projeto e aplicação
2. ✅ Listar tipos disponíveis
3. ✅ Pedir suas credenciais
4. ✅ Criar a métrica
5. ✅ Aguardar coleta
6. ✅ Mostrar resultados

## 🔍 Interpretando Resultados

### ✅ Conexão OK
```json
{
  "connection_status": "connected",
  "connection_time_ms": 45,      // < 100ms é bom
  "connection_ping_time_ms": 12  // < 50ms é bom
}
```

### ❌ Conexão Falhou
```json
{
  "connection_status": "failed",
  "connection_error": "authentication failed for user",
  "connection_time_ms": 120
}
```

### ⏱️ Timeout
```json
{
  "connection_status": "timeout",
  "connection_error": "connection timeout",
  "connection_time_ms": 5000
}
```

## 🔧 Troubleshooting Rápido

### "connection refused"
```bash
# Verificar se o serviço está acessível
kubectl get svc -n default | grep postgres
kubectl get pods -n default | grep postgres
```

### "authentication failed"
```bash
# Confirmar credenciais
kubectl exec -it postgres-pod -- psql -U myuser -d mydb -c "SELECT version();"
```

### "metric type not found"
```bash
# Verificar se migrations rodaram
curl http://localhost:8080/api/v1/metric-types | jq '.[].name'
```

### Métricas não sendo coletadas
```bash
# Verificar logs do monitoramento
kubectl logs -f deployment/k8s-monitoring-app | grep "metric collection"
```

## 📊 Monitoramento Contínuo

### Loop simples de monitoramento
```bash
#!/bin/bash
while true; do
  clear
  echo "=== Connection Status ==="
  curl -s http://localhost:8080/api/v1/application-metric-values/application/YOUR_APP_ID/latest | \
    jq -r '.[] | "\(.value.connection_status) - \(.value.connection_time_ms)ms"'
  sleep 60
done
```

## 📚 Próximos Passos

1. **Configurar mais conexões**
   - Adicione todas as dependências da sua aplicação
   - Redis, banco primário, banco de cache, etc.

2. **Ajustar timeouts**
   - Comece com 5s
   - Ajuste baseado nos seus SLAs

3. **Configurar alertas**
   - Monitore status "failed"
   - Alerte em timeouts consecutivos

4. **Documentar credenciais**
   - Mantenha registro de quais usuários são usados
   - Configure rotação periódica

5. **Segurança**
   - Use usuários read-only
   - Configure SSL em produção
   - Planeje criptografia de credenciais

## 🆘 Precisa de Ajuda?

- **Documentação completa**: [CONNECTION_METRICS.md](CONNECTION_METRICS.md)
- **Exemplos Postman**: [../postman/CONNECTION_METRICS_EXAMPLES.md](../postman/CONNECTION_METRICS_EXAMPLES.md)
- **Script de teste**: [../examples/connection-metrics-test.sh](../examples/connection-metrics-test.sh)
- **Resumo técnico**: [CONNECTION_METRICS_SUMMARY.md](CONNECTION_METRICS_SUMMARY.md)

## 💡 Dicas

1. **Teste localmente primeiro**
   ```bash
   docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=test postgres:14
   ```

2. **Use o Postman**
   - Importe a collection atualizada
   - Todos os exemplos já estão prontos

3. **Monitore os logs**
   - Erros de conexão aparecem no log da aplicação
   - Útil para debugging

4. **Comece simples**
   - Configure uma conexão por vez
   - Valide que funciona antes de adicionar mais

5. **Documente**
   - Mantenha registro de host:port:database
   - Facilita troubleshooting futuro

