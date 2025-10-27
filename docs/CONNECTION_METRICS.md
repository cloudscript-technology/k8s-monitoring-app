# Métricas de Conexão de Banco de Dados e Serviços

Este documento descreve como configurar e usar as métricas de conexão para monitorar a disponibilidade e performance de bancos de dados e serviços.

## Tipos de Métricas de Conexão Disponíveis

A aplicação suporta os seguintes tipos de métricas de conexão:

1. **RedisConnection** - Teste de conexão Redis com autenticação
2. **PostgreSQLConnection** - Teste de conexão PostgreSQL com autenticação
3. **MongoDBConnection** - Teste de conexão MongoDB com autenticação
4. **MySQLConnection** - Teste de conexão MySQL com autenticação
5. **KongConnection** - Teste de conexão Kong API Gateway

## Campos de Configuração

Todos os tipos de métricas de conexão compartilham os seguintes campos na configuração:

### Campos Comuns

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `connection_host` | string | Sim | Host ou endereço IP do serviço |
| `connection_port` | int | Sim | Porta do serviço |
| `connection_username` | string | Condicional | Usuário para autenticação (se necessário) |
| `connection_password` | string | Condicional | Senha para autenticação (se necessário) |
| `connection_ssl` | bool | Não | Usar conexão SSL/TLS (padrão: false) |
| `connection_timeout` | int | Não | Timeout em segundos (padrão: 5) |

### Campos Específicos por Tipo

#### PostgreSQL e MySQL
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `connection_database` | string | Sim | Nome do banco de dados |

#### MongoDB
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `connection_database` | string | Sim | Nome do banco de dados |
| `connection_auth_source` | string | Não | Banco de autenticação (padrão: "admin") |

#### Redis
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `connection_db` | int | Não | Número do banco Redis (padrão: 0) |

#### Kong
| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `kong_admin_url` | string | Não | URL da API Admin do Kong |

## Valores de Métricas Retornadas

Todas as métricas de conexão retornam os seguintes campos:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `connection_status` | string | Status da conexão: "connected", "failed", ou "timeout" |
| `connection_time_ms` | int64 | Tempo para estabelecer a conexão em milissegundos |
| `connection_error` | string | Mensagem de erro se a conexão falhou |
| `connection_version` | string | Versão do banco de dados ou serviço |
| `connection_info` | string | Informações adicionais sobre a conexão |
| `connection_ping_time_ms` | int64 | Tempo de ping/query em milissegundos |

## Exemplos de Uso

### 1. Monitorar Conexão Redis

```bash
POST /api/v1/application-metrics
{
  "application_id": "uuid-da-aplicacao",
  "type_id": "uuid-do-tipo-RedisConnection",
  "configuration": {
    "connection_host": "redis.default.svc.cluster.local",
    "connection_port": 6379,
    "connection_password": "senha-do-redis",
    "connection_db": 0,
    "connection_timeout": 5
  }
}
```

### 2. Monitorar Conexão PostgreSQL

```bash
POST /api/v1/application-metrics
{
  "application_id": "uuid-da-aplicacao",
  "type_id": "uuid-do-tipo-PostgreSQLConnection",
  "configuration": {
    "connection_host": "postgres.default.svc.cluster.local",
    "connection_port": 5432,
    "connection_username": "myuser",
    "connection_password": "mypassword",
    "connection_database": "mydb",
    "connection_ssl": true,
    "connection_timeout": 10
  }
}
```

### 3. Monitorar Conexão MongoDB

```bash
POST /api/v1/application-metrics
{
  "application_id": "uuid-da-aplicacao",
  "type_id": "uuid-do-tipo-MongoDBConnection",
  "configuration": {
    "connection_host": "mongodb.default.svc.cluster.local",
    "connection_port": 27017,
    "connection_username": "admin",
    "connection_password": "admin-password",
    "connection_database": "mydb",
    "connection_auth_source": "admin",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

### 4. Monitorar Conexão MySQL

```bash
POST /api/v1/application-metrics
{
  "application_id": "uuid-da-aplicacao",
  "type_id": "uuid-do-tipo-MySQLConnection",
  "configuration": {
    "connection_host": "mysql.default.svc.cluster.local",
    "connection_port": 3306,
    "connection_username": "root",
    "connection_password": "root-password",
    "connection_database": "mydb",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

### 5. Monitorar Conexão Kong

```bash
POST /api/v1/application-metrics
{
  "application_id": "uuid-da-aplicacao",
  "type_id": "uuid-do-tipo-KongConnection",
  "configuration": {
    "connection_host": "kong-admin.default.svc.cluster.local",
    "connection_port": 8001,
    "kong_admin_url": "http://kong-admin.default.svc.cluster.local:8001",
    "connection_username": "admin",
    "connection_password": "admin-password",
    "connection_ssl": false,
    "connection_timeout": 5
  }
}
```

## Exemplo de Resposta de Métrica

Quando a aplicação coleta uma métrica de conexão, ela retorna um objeto com os seguintes dados:

### Conexão Bem-Sucedida

```json
{
  "id": "uuid-do-valor",
  "application_metric_id": "uuid-da-metrica",
  "value": {
    "connection_status": "connected",
    "connection_time_ms": 45,
    "connection_ping_time_ms": 12,
    "connection_version": "PostgreSQL 14.5 on x86_64-pc-linux-gnu",
    "connection_info": "Database size: 52428800 bytes"
  },
  "created_at": "2024-10-27T10:30:00Z"
}
```

### Conexão Falhou

```json
{
  "id": "uuid-do-valor",
  "application_metric_id": "uuid-da-metrica",
  "value": {
    "connection_status": "failed",
    "connection_time_ms": 5002,
    "connection_error": "dial tcp 10.0.0.1:5432: i/o timeout"
  },
  "created_at": "2024-10-27T10:30:00Z"
}
```

### Timeout de Conexão

```json
{
  "id": "uuid-do-valor",
  "application_metric_id": "uuid-da-metrica",
  "value": {
    "connection_status": "timeout",
    "connection_time_ms": 5000,
    "connection_error": "connection timeout"
  },
  "created_at": "2024-10-27T10:30:00Z"
}
```

## Segurança

### Armazenamento de Credenciais

As credenciais de conexão (usuário e senha) são armazenadas no campo `configuration` da tabela `application_metrics` no formato JSONB. 

**⚠️ IMPORTANTE:** As credenciais são armazenadas em texto claro no banco de dados. Para ambientes de produção, considere:

1. Usar secrets do Kubernetes montados como variáveis de ambiente
2. Integrar com um sistema de gerenciamento de secrets (HashiCorp Vault, AWS Secrets Manager, etc.)
3. Implementar criptografia para o campo `configuration`
4. Restringir o acesso ao banco de dados da aplicação

### Boas Práticas

1. **Use usuários com permissões limitadas**: Crie usuários de banco de dados específicos para monitoramento com permissões somente de leitura
2. **Rotação de credenciais**: Implemente rotação regular de senhas
3. **Conexões SSL/TLS**: Sempre use `connection_ssl: true` em produção quando o banco suportar
4. **Timeouts adequados**: Configure timeouts apropriados para evitar bloqueios longos
5. **Monitoramento de falhas**: Configure alertas para quando as conexões falharem consistentemente

## Consultar Métricas Coletadas

Para consultar os valores coletados das métricas de conexão:

```bash
# Obter últimos valores para uma aplicação
GET /api/v1/application-metric-values/application/{application_id}/latest

# Obter histórico de uma métrica específica
GET /api/v1/application-metric-values/application-metric/{application_metric_id}
```

## Frequência de Coleta

A frequência de coleta das métricas é controlada pela variável de ambiente `METRICS_COLLECTION_INTERVAL` (em segundos). O padrão é 60 segundos.

Para alterar a frequência:

```bash
export METRICS_COLLECTION_INTERVAL=30  # Coletar a cada 30 segundos
```

## Troubleshooting

### Erro: "connection timeout"

**Causa**: O serviço não está acessível dentro do timeout configurado.

**Solução**:
- Verifique se o host e porta estão corretos
- Aumente o `connection_timeout`
- Verifique as regras de firewall/network policies
- Confirme que o serviço está rodando

### Erro: "authentication failed"

**Causa**: Credenciais incorretas.

**Solução**:
- Verifique username e password
- Para MongoDB, confirme o `connection_auth_source`
- Para Redis, verifique se a senha está correta e se AUTH está habilitado

### Erro: "database does not exist"

**Causa**: O banco de dados especificado não existe.

**Solução**:
- Verifique o nome do banco de dados em `connection_database`
- Crie o banco de dados se necessário

### Métricas não sendo coletadas

**Causa**: Serviço de monitoramento não está rodando ou encontrou erro.

**Solução**:
- Verifique os logs da aplicação
- Confirme que o serviço de monitoramento foi iniciado
- Verifique se a métrica está corretamente configurada

## Exemplos Completos com cURL

### Criar Projeto e Aplicação

```bash
# 1. Criar projeto
PROJECT_ID=$(curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Project",
    "description": "Project for database monitoring"
  }' | jq -r '.id')

# 2. Criar aplicação
APP_ID=$(curl -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"name\": \"My Database App\",
    \"description\": \"Application with database connections\",
    \"namespace\": \"default\"
  }" | jq -r '.id')

# 3. Listar tipos de métricas e pegar o ID de PostgreSQLConnection
METRIC_TYPE_ID=$(curl -X GET http://localhost:8080/api/v1/metric-types \
  | jq -r '.[] | select(.name=="PostgreSQLConnection") | .id')

# 4. Criar métrica de conexão PostgreSQL
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$METRIC_TYPE_ID\",
    \"configuration\": {
      \"connection_host\": \"postgres.default.svc.cluster.local\",
      \"connection_port\": 5432,
      \"connection_username\": \"myuser\",
      \"connection_password\": \"mypassword\",
      \"connection_database\": \"mydb\",
      \"connection_ssl\": false,
      \"connection_timeout\": 5
    }
  }"

# 5. Aguardar coleta automática ou consultar valores após intervalo de coleta
sleep 65

# 6. Obter últimos valores coletados
curl -X GET "http://localhost:8080/api/v1/application-metric-values/application/$APP_ID/latest"
```

## Próximos Passos

- Implementar criptografia de credenciais
- Adicionar suporte para mais tipos de banco de dados (Oracle, Cassandra, etc.)
- Implementar cache de conexões para melhor performance
- Adicionar métricas de latência e throughput
- Criar dashboards específicos para métricas de conexão

