# Exemplos de Métricas de Conexão para Postman

Este documento contém exemplos de requisições para configurar e testar as métricas de conexão de banco de dados.

## Pré-requisitos

Antes de usar estas métricas, você precisa:

1. Ter um projeto criado
2. Ter uma aplicação criada
3. Obter os IDs dos tipos de métricas

## 1. Listar Tipos de Métricas de Conexão

```http
GET {{base_url}}/api/v1/metric-types
```

Procure pelos seguintes tipos:
- `RedisConnection`
- `PostgreSQLConnection`
- `MongoDBConnection`
- `MySQLConnection`
- `KongConnection`

Anote os IDs (UUIDs) de cada tipo que você deseja usar.

## 2. Exemplos de Configuração

### Redis Connection

```http
POST {{base_url}}/api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "{{application_id}}",
  "type_id": "{{redis_connection_type_id}}",
  "configuration": {
    "connection_host": "redis-service.default.svc.cluster.local",
    "connection_port": 6379,
    "connection_password": "your-redis-password",
    "connection_db": 0,
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

**Notas:**
- Se Redis não tiver senha, omita `connection_password`
- `connection_db` é o número do banco Redis (0-15)

### PostgreSQL Connection

```http
POST {{base_url}}/api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "{{application_id}}",
  "type_id": "{{postgresql_connection_type_id}}",
  "configuration": {
    "connection_host": "postgres-service.default.svc.cluster.local",
    "connection_port": 5432,
    "connection_username": "postgres",
    "connection_password": "postgres-password",
    "connection_database": "myapp_db",
    "connection_ssl": false,
    "connection_timeout": 10
  }
}
```

**Notas:**
- Para ambientes de produção, use `connection_ssl: true`
- Para RDS da AWS, use o endpoint fornecido como `connection_host`

### MongoDB Connection

```http
POST {{base_url}}/api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "{{application_id}}",
  "type_id": "{{mongodb_connection_type_id}}",
  "configuration": {
    "connection_host": "mongodb-service.default.svc.cluster.local",
    "connection_port": 27017,
    "connection_username": "admin",
    "connection_password": "mongo-password",
    "connection_database": "myapp_db",
    "connection_auth_source": "admin",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

**Notas:**
- `connection_auth_source` geralmente é "admin" para o usuário root
- Para MongoDB Atlas, use `connection_ssl: true` e o hostname fornecido

### MySQL Connection

```http
POST {{base_url}}/api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "{{application_id}}",
  "type_id": "{{mysql_connection_type_id}}",
  "configuration": {
    "connection_host": "mysql-service.default.svc.cluster.local",
    "connection_port": 3306,
    "connection_username": "root",
    "connection_password": "mysql-password",
    "connection_database": "myapp_db",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

**Notas:**
- Para RDS MySQL da AWS, use SSL habilitado
- Verifique se o usuário tem permissão para acessar `information_schema`

### Kong Connection

```http
POST {{base_url}}/api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "{{application_id}}",
  "type_id": "{{kong_connection_type_id}}",
  "configuration": {
    "connection_host": "kong-admin.default.svc.cluster.local",
    "connection_port": 8001,
    "kong_admin_url": "http://kong-admin.default.svc.cluster.local:8001",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

**Notas:**
- Se Kong Admin API requer autenticação, adicione `connection_username` e `connection_password`
- `kong_admin_url` pode ser omitido se usar host:port padrão

## 3. Consultar Valores Coletados

### Obter Últimos Valores por Aplicação

```http
GET {{base_url}}/api/v1/application-metric-values/application/{{application_id}}/latest
```

### Obter Histórico de uma Métrica Específica

```http
GET {{base_url}}/api/v1/application-metric-values/application-metric/{{application_metric_id}}
```

## 4. Cenários de Teste Completos

### Cenário 1: Monitorar Stack Completo (PostgreSQL + Redis + Kong)

```bash
# 1. Criar aplicação
POST {{base_url}}/api/v1/applications
{
  "project_id": "{{project_id}}",
  "name": "Complete Stack",
  "description": "Full application stack monitoring",
  "namespace": "production"
}

# 2. Adicionar PostgreSQL
POST {{base_url}}/api/v1/application-metrics
{
  "application_id": "{{application_id}}",
  "type_id": "{{postgresql_connection_type_id}}",
  "configuration": {
    "connection_host": "postgres.production.svc.cluster.local",
    "connection_port": 5432,
    "connection_username": "app_user",
    "connection_password": "secure_password",
    "connection_database": "app_db",
    "connection_ssl": true,
    "connection_timeout": 10
  }
}

# 3. Adicionar Redis
POST {{base_url}}/api/v1/application-metrics
{
  "application_id": "{{application_id}}",
  "type_id": "{{redis_connection_type_id}}",
  "configuration": {
    "connection_host": "redis.production.svc.cluster.local",
    "connection_port": 6379,
    "connection_password": "redis_password",
    "connection_db": 0,
    "connection_timeout": 5
  }
}

# 4. Adicionar Kong
POST {{base_url}}/api/v1/application-metrics
{
  "application_id": "{{application_id}}",
  "type_id": "{{kong_connection_type_id}}",
  "configuration": {
    "connection_host": "kong-admin.production.svc.cluster.local",
    "connection_port": 8001,
    "connection_timeout": 5
  }
}
```

### Cenário 2: Monitorar Múltiplos Ambientes do Mesmo Banco

```bash
# PostgreSQL - Produção
POST {{base_url}}/api/v1/application-metrics
{
  "application_id": "{{app_prod_id}}",
  "type_id": "{{postgresql_connection_type_id}}",
  "configuration": {
    "connection_host": "postgres.production.svc.cluster.local",
    "connection_port": 5432,
    "connection_username": "prod_user",
    "connection_password": "prod_password",
    "connection_database": "prod_db",
    "connection_ssl": true,
    "connection_timeout": 10
  }
}

# PostgreSQL - Staging
POST {{base_url}}/api/v1/application-metrics
{
  "application_id": "{{app_staging_id}}",
  "type_id": "{{postgresql_connection_type_id}}",
  "configuration": {
    "connection_host": "postgres.staging.svc.cluster.local",
    "connection_port": 5432,
    "connection_username": "staging_user",
    "connection_password": "staging_password",
    "connection_database": "staging_db",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}

# PostgreSQL - Development
POST {{base_url}}/api/v1/application-metrics
{
  "application_id": "{{app_dev_id}}",
  "type_id": "{{postgresql_connection_type_id}}",
  "configuration": {
    "connection_host": "postgres.development.svc.cluster.local",
    "connection_port": 5432,
    "connection_username": "dev_user",
    "connection_password": "dev_password",
    "connection_database": "dev_db",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

## 5. Interpretação dos Resultados

### Conexão Saudável
```json
{
  "value": {
    "connection_status": "connected",
    "connection_time_ms": 45,
    "connection_ping_time_ms": 12,
    "connection_version": "PostgreSQL 14.5",
    "connection_info": "Database size: 52428800 bytes"
  }
}
```

**Indicadores:**
- `connection_status`: "connected"
- `connection_time_ms`: < 100ms (bom), < 500ms (aceitável)
- `connection_ping_time_ms`: < 50ms (bom)

### Conexão com Problemas
```json
{
  "value": {
    "connection_status": "timeout",
    "connection_time_ms": 5000,
    "connection_error": "connection timeout"
  }
}
```

**Ações:**
- Verificar conectividade de rede
- Verificar se o serviço está rodando
- Aumentar timeout se necessário

### Falha de Autenticação
```json
{
  "value": {
    "connection_status": "failed",
    "connection_time_ms": 120,
    "connection_error": "pq: password authentication failed"
  }
}
```

**Ações:**
- Verificar credenciais
- Confirmar permissões do usuário

## 6. Variáveis do Postman

Configure estas variáveis no seu ambiente Postman:

```json
{
  "base_url": "http://localhost:8080",
  "project_id": "<uuid-do-projeto>",
  "application_id": "<uuid-da-aplicacao>",
  "redis_connection_type_id": "<uuid-do-tipo>",
  "postgresql_connection_type_id": "<uuid-do-tipo>",
  "mongodb_connection_type_id": "<uuid-do-tipo>",
  "mysql_connection_type_id": "<uuid-do-tipo>",
  "kong_connection_type_id": "<uuid-do-tipo>"
}
```

## 7. Atualizar Configuração de Métrica

Se você precisa atualizar as credenciais ou configuração:

```http
PUT {{base_url}}/api/v1/application-metrics/{{application_metric_id}}
Content-Type: application/json

{
  "configuration": {
    "connection_host": "new-host.example.com",
    "connection_port": 5432,
    "connection_username": "new_user",
    "connection_password": "new_password",
    "connection_database": "mydb",
    "connection_ssl": true,
    "connection_timeout": 10
  }
}
```

## 8. Deletar Métrica

```http
DELETE {{base_url}}/api/v1/application-metrics/{{application_metric_id}}
```

## 9. Tips e Best Practices

1. **Comece com timeouts baixos**: Use 5 segundos inicialmente e ajuste conforme necessário
2. **Teste em ambiente de desenvolvimento primeiro**: Valide as credenciais antes de usar em produção
3. **Use SSL em produção**: Sempre configure `connection_ssl: true` para dados sensíveis
4. **Monitore os logs**: Verifique os logs da aplicação para ver detalhes das tentativas de conexão
5. **Rotação de credenciais**: Quando rotacionar senhas, lembre-se de atualizar as métricas
6. **Usuários dedicados**: Crie usuários específicos para monitoramento com permissões limitadas

## 10. Troubleshooting

### Problema: "no required module provides package"

**Solução**: Execute `go mod tidy` no servidor

### Problema: Métricas não sendo coletadas

**Solução**: Verifique se o serviço de monitoramento está rodando:
```bash
# Nos logs da aplicação, procure por:
# "Monitoring service started"
```

### Problema: "connection refused"

**Solução**: 
- Verifique se o host e porta estão corretos
- Teste conectividade com `telnet host port`
- Verifique network policies do Kubernetes

### Problema: Valores muito altos de latência

**Solução**:
- Verifique a saúde do banco de dados
- Analise queries longas em execução
- Considere escalar os recursos do banco

