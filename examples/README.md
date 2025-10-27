# Exemplos de Uso

Este diretório contém scripts e exemplos para testar e utilizar a aplicação de monitoramento K8s.

## Scripts Disponíveis

### connection-metrics-test.sh

Script interativo para testar métricas de conexão de banco de dados e serviços.

**Uso:**

```bash
# Usando URL padrão (http://localhost:8080)
./examples/connection-metrics-test.sh

# Usando URL customizada
./examples/connection-metrics-test.sh http://seu-servidor:8080
```

**Pré-requisitos:**
- `curl` instalado
- `jq` instalado (para parsing de JSON)
  - macOS: `brew install jq`
  - Linux: `apt-get install jq` ou `yum install jq`

**O que o script faz:**

1. Cria um projeto de teste
2. Cria uma aplicação de teste
3. Lista os tipos de métricas disponíveis
4. Permite escolher qual tipo de conexão testar:
   - Redis
   - PostgreSQL
   - MongoDB
   - MySQL
   - Kong
5. Solicita as credenciais e configurações de conexão
6. Cria a métrica na aplicação
7. Aguarda a coleta automática (60 segundos)
8. Exibe os resultados da conexão

**Exemplo de uso para PostgreSQL:**

```bash
$ ./examples/connection-metrics-test.sh

==========================================
Teste de Métricas de Conexão
==========================================
API URL: http://localhost:8080/api/v1

1. Criando projeto...
✓ Project created: 123e4567-e89b-12d3-a456-426614174000

2. Criando aplicação...
✓ Application created: 223e4567-e89b-12d3-a456-426614174001

3. Obtendo tipos de métricas de conexão...
Available metric types:
  Redis:      323e4567-e89b-12d3-a456-426614174002
  PostgreSQL: 423e4567-e89b-12d3-a456-426614174003
  MongoDB:    523e4567-e89b-12d3-a456-426614174004
  MySQL:      623e4567-e89b-12d3-a456-426614174005
  Kong:       723e4567-e89b-12d3-a456-426614174006

==========================================
Escolha qual métrica de conexão testar:
==========================================
1) Redis
2) PostgreSQL
3) MongoDB
4) MySQL
5) Kong
6) Todas (exemplo com valores de teste)
0) Sair

Opção: 2

Configurando PostgreSQL Connection
Host (default: localhost): postgres.default.svc.cluster.local
Port (default: 5432): 5432
Username: myuser
Password: ********
Database: mydb
Use SSL? (y/n, default: n): n

4. Criando métrica de conexão...
✓ Metric created: 823e4567-e89b-12d3-a456-426614174007

==========================================
Aguardando coleta automática...
==========================================
A coleta de métricas ocorre a cada 60 segundos (padrão)

Aguardando 65 segundos...

5. Consultando valores coletados...

==========================================
Resultados:
==========================================
[
  {
    "id": "923e4567-e89b-12d3-a456-426614174008",
    "application_metric_id": "823e4567-e89b-12d3-a456-426614174007",
    "value": {
      "connection_status": "connected",
      "connection_time_ms": 45,
      "connection_ping_time_ms": 12,
      "connection_version": "PostgreSQL 14.5 on x86_64-pc-linux-gnu",
      "connection_info": "Database size: 52428800 bytes"
    },
    "created_at": "2024-10-27T10:30:00Z"
  }
]

✓ Conexão estabelecida com sucesso!

Métricas de Performance:
  Tempo de conexão: 45ms
  Tempo de ping:    12ms
  Versão:           PostgreSQL 14.5 on x86_64-pc-linux-gnu

==========================================
IDs para referência:
==========================================
Project ID:     123e4567-e89b-12d3-a456-426614174000
Application ID: 223e4567-e89b-12d3-a456-426614174001
Metric ID:      823e4567-e89b-12d3-a456-426614174007

Para consultar valores novamente:
  curl http://localhost:8080/api/v1/application-metric-values/application/223e4567-e89b-12d3-a456-426614174001/latest | jq

Para consultar histórico da métrica:
  curl http://localhost:8080/api/v1/application-metric-values/application-metric/823e4567-e89b-12d3-a456-426614174007 | jq
```

## Exemplos Adicionais

### Criar métrica Redis manualmente

```bash
# 1. Obter tipo de métrica Redis
REDIS_TYPE_ID=$(curl -s http://localhost:8080/api/v1/metric-types | jq -r '.[] | select(.name=="RedisConnection") | .id')

# 2. Criar métrica
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"YOUR_APP_ID\",
    \"type_id\": \"$REDIS_TYPE_ID\",
    \"configuration\": {
      \"connection_host\": \"redis.default.svc.cluster.local\",
      \"connection_port\": 6379,
      \"connection_password\": \"your-password\",
      \"connection_db\": 0,
      \"connection_timeout\": 5
    }
  }"
```

### Criar métrica MongoDB manualmente

```bash
# 1. Obter tipo de métrica MongoDB
MONGODB_TYPE_ID=$(curl -s http://localhost:8080/api/v1/metric-types | jq -r '.[] | select(.name=="MongoDBConnection") | .id')

# 2. Criar métrica
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"YOUR_APP_ID\",
    \"type_id\": \"$MONGODB_TYPE_ID\",
    \"configuration\": {
      \"connection_host\": \"mongodb.default.svc.cluster.local\",
      \"connection_port\": 27017,
      \"connection_username\": \"admin\",
      \"connection_password\": \"admin-password\",
      \"connection_database\": \"mydb\",
      \"connection_auth_source\": \"admin\",
      \"connection_ssl\": false,
      \"connection_timeout\": 5
    }
  }"
```

### Consultar métricas coletadas

```bash
# Últimos valores de todas as métricas de uma aplicação
curl http://localhost:8080/api/v1/application-metric-values/application/YOUR_APP_ID/latest | jq

# Histórico de uma métrica específica
curl http://localhost:8080/api/v1/application-metric-values/application-metric/YOUR_METRIC_ID | jq

# Filtrar apenas conexões bem-sucedidas
curl http://localhost:8080/api/v1/application-metric-values/application/YOUR_APP_ID/latest | \
  jq '.[] | select(.value.connection_status == "connected")'

# Calcular média de tempo de conexão
curl http://localhost:8080/api/v1/application-metric-values/application-metric/YOUR_METRIC_ID | \
  jq '[.[].value.connection_time_ms] | add / length'
```

## Dicas de Uso

### 1. Testando Localmente

Para testar as métricas de conexão localmente sem Kubernetes:

```bash
# Inicie serviços locais com Docker
docker run -d --name postgres -e POSTGRES_PASSWORD=test -p 5432:5432 postgres:14
docker run -d --name redis -p 6379:6379 redis:7
docker run -d --name mongodb -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=admin -p 27017:27017 mongo:6
docker run -d --name mysql -e MYSQL_ROOT_PASSWORD=test -p 3306:3306 mysql:8

# Execute o script de teste
./examples/connection-metrics-test.sh
```

### 2. Monitoramento Contínuo

Para monitorar continuamente as conexões:

```bash
# Criar script de monitoramento
cat > monitor-connections.sh << 'EOF'
#!/bin/bash
APP_ID="YOUR_APP_ID"
while true; do
  clear
  echo "=== Connection Monitoring ==="
  echo "Time: $(date)"
  echo ""
  curl -s http://localhost:8080/api/v1/application-metric-values/application/$APP_ID/latest | \
    jq -r '.[] | "\(.value.connection_status) - \(.value.connection_time_ms)ms - \(.value.connection_version // "N/A")"'
  sleep 60
done
EOF

chmod +x monitor-connections.sh
./monitor-connections.sh
```

### 3. Alertas Simples

Criar um alerta simples para falhas de conexão:

```bash
#!/bin/bash
APP_ID="YOUR_APP_ID"
WEBHOOK_URL="YOUR_WEBHOOK_URL"  # Slack, Discord, etc.

while true; do
  status=$(curl -s http://localhost:8080/api/v1/application-metric-values/application/$APP_ID/latest | \
    jq -r '.[0].value.connection_status')
  
  if [ "$status" != "connected" ]; then
    curl -X POST $WEBHOOK_URL \
      -H "Content-Type: application/json" \
      -d "{\"text\":\"⚠️ Connection failed: $status\"}"
  fi
  
  sleep 60
done
```

## Próximos Passos

Após configurar as métricas de conexão:

1. Configure alertas baseados nos status de conexão
2. Crie dashboards para visualizar as métricas ao longo do tempo
3. Implemente rotação de credenciais automatizada
4. Configure diferentes frequências de coleta para ambientes críticos
5. Integre com sistemas de monitoramento existentes (Prometheus, Grafana, etc.)

## Troubleshooting

### Script não encontra jq

```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

### Conexão falha com "connection refused"

- Verifique se o host e porta estão corretos
- Para serviços Kubernetes, use o formato: `service-name.namespace.svc.cluster.local`
- Teste conectividade: `telnet host port`

### Timeout na conexão

- Aumente o `connection_timeout`
- Verifique network policies do Kubernetes
- Confirme que o serviço está rodando: `kubectl get pods -n namespace`

### Autenticação falha

- Verifique credenciais (username/password)
- Para MongoDB, confirme o `connection_auth_source`
- Para PostgreSQL/MySQL, verifique permissões do usuário

## Contribuindo

Para adicionar novos exemplos ou melhorar os existentes, por favor:

1. Crie um branch
2. Adicione seus exemplos
3. Teste completamente
4. Envie um Pull Request com descrição detalhada

