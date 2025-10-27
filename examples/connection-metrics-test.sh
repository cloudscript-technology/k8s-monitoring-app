#!/bin/bash

# Script para testar métricas de conexão de banco de dados
# Usage: ./connection-metrics-test.sh [base_url]

set -e

BASE_URL="${1:-http://localhost:8080}"
API_URL="${BASE_URL}/api/v1"

echo "=========================================="
echo "Teste de Métricas de Conexão"
echo "=========================================="
echo "API URL: $API_URL"
echo ""

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para verificar se jq está instalado
check_dependencies() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        echo "Install with: brew install jq (macOS) or apt-get install jq (Linux)"
        exit 1
    fi
}

# Função para fazer requisição e tratar erros
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" "${API_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -X "$method" "${API_URL}${endpoint}")
    fi
    
    echo "$response"
}

# 1. Criar projeto
echo "1. Criando projeto..."
project_data='{
  "name": "Database Monitoring Test",
  "description": "Test project for database connection metrics"
}'

project_response=$(api_call "POST" "/projects" "$project_data")
PROJECT_ID=$(echo "$project_response" | jq -r '.id')

if [ "$PROJECT_ID" == "null" ] || [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Failed to create project${NC}"
    echo "$project_response"
    exit 1
fi

echo -e "${GREEN}✓ Project created: $PROJECT_ID${NC}"
echo ""

# 2. Criar aplicação
echo "2. Criando aplicação..."
app_data=$(cat <<EOF
{
  "project_id": "$PROJECT_ID",
  "name": "Database Test App",
  "description": "Application for testing database connections",
  "namespace": "default"
}
EOF
)

app_response=$(api_call "POST" "/applications" "$app_data")
APP_ID=$(echo "$app_response" | jq -r '.id')

if [ "$APP_ID" == "null" ] || [ -z "$APP_ID" ]; then
    echo -e "${RED}Failed to create application${NC}"
    echo "$app_response"
    exit 1
fi

echo -e "${GREEN}✓ Application created: $APP_ID${NC}"
echo ""

# 3. Obter tipos de métricas
echo "3. Obtendo tipos de métricas de conexão..."
metric_types_response=$(api_call "GET" "/metric-types")

REDIS_TYPE_ID=$(echo "$metric_types_response" | jq -r '.[] | select(.name=="RedisConnection") | .id')
POSTGRES_TYPE_ID=$(echo "$metric_types_response" | jq -r '.[] | select(.name=="PostgreSQLConnection") | .id')
MONGODB_TYPE_ID=$(echo "$metric_types_response" | jq -r '.[] | select(.name=="MongoDBConnection") | .id')
MYSQL_TYPE_ID=$(echo "$metric_types_response" | jq -r '.[] | select(.name=="MySQLConnection") | .id')
KONG_TYPE_ID=$(echo "$metric_types_response" | jq -r '.[] | select(.name=="KongConnection") | .id')

echo "Available metric types:"
echo "  Redis:      $REDIS_TYPE_ID"
echo "  PostgreSQL: $POSTGRES_TYPE_ID"
echo "  MongoDB:    $MONGODB_TYPE_ID"
echo "  MySQL:      $MYSQL_TYPE_ID"
echo "  Kong:       $KONG_TYPE_ID"
echo ""

# 4. Menu interativo para escolher qual métrica testar
echo "=========================================="
echo "Escolha qual métrica de conexão testar:"
echo "=========================================="
echo "1) Redis"
echo "2) PostgreSQL"
echo "3) MongoDB"
echo "4) MySQL"
echo "5) Kong"
echo "6) Todas (exemplo com valores de teste)"
echo "0) Sair"
echo ""
read -p "Opção: " option

case $option in
    1)
        echo -e "\n${YELLOW}Configurando Redis Connection${NC}"
        read -p "Host (default: localhost): " redis_host
        redis_host=${redis_host:-localhost}
        read -p "Port (default: 6379): " redis_port
        redis_port=${redis_port:-6379}
        read -p "Password (deixe vazio se não houver): " redis_password
        read -p "Database number (default: 0): " redis_db
        redis_db=${redis_db:-0}
        
        metric_data=$(cat <<EOF
{
  "application_id": "$APP_ID",
  "type_id": "$REDIS_TYPE_ID",
  "configuration": {
    "connection_host": "$redis_host",
    "connection_port": $redis_port,
    "connection_password": "$redis_password",
    "connection_db": $redis_db,
    "connection_timeout": 5
  }
}
EOF
        )
        ;;
    
    2)
        echo -e "\n${YELLOW}Configurando PostgreSQL Connection${NC}"
        read -p "Host (default: localhost): " pg_host
        pg_host=${pg_host:-localhost}
        read -p "Port (default: 5432): " pg_port
        pg_port=${pg_port:-5432}
        read -p "Username: " pg_user
        read -s -p "Password: " pg_password
        echo ""
        read -p "Database: " pg_database
        read -p "Use SSL? (y/n, default: n): " pg_ssl
        pg_ssl_bool="false"
        [[ "$pg_ssl" == "y" ]] && pg_ssl_bool="true"
        
        metric_data=$(cat <<EOF
{
  "application_id": "$APP_ID",
  "type_id": "$POSTGRES_TYPE_ID",
  "configuration": {
    "connection_host": "$pg_host",
    "connection_port": $pg_port,
    "connection_username": "$pg_user",
    "connection_password": "$pg_password",
    "connection_database": "$pg_database",
    "connection_ssl": $pg_ssl_bool,
    "connection_timeout": 10
  }
}
EOF
        )
        ;;
    
    3)
        echo -e "\n${YELLOW}Configurando MongoDB Connection${NC}"
        read -p "Host (default: localhost): " mongo_host
        mongo_host=${mongo_host:-localhost}
        read -p "Port (default: 27017): " mongo_port
        mongo_port=${mongo_port:-27017}
        read -p "Username: " mongo_user
        read -s -p "Password: " mongo_password
        echo ""
        read -p "Database: " mongo_database
        read -p "Auth Source (default: admin): " mongo_auth_source
        mongo_auth_source=${mongo_auth_source:-admin}
        
        metric_data=$(cat <<EOF
{
  "application_id": "$APP_ID",
  "type_id": "$MONGODB_TYPE_ID",
  "configuration": {
    "connection_host": "$mongo_host",
    "connection_port": $mongo_port,
    "connection_username": "$mongo_user",
    "connection_password": "$mongo_password",
    "connection_database": "$mongo_database",
    "connection_auth_source": "$mongo_auth_source",
    "connection_timeout": 5
  }
}
EOF
        )
        ;;
    
    4)
        echo -e "\n${YELLOW}Configurando MySQL Connection${NC}"
        read -p "Host (default: localhost): " mysql_host
        mysql_host=${mysql_host:-localhost}
        read -p "Port (default: 3306): " mysql_port
        mysql_port=${mysql_port:-3306}
        read -p "Username: " mysql_user
        read -s -p "Password: " mysql_password
        echo ""
        read -p "Database: " mysql_database
        
        metric_data=$(cat <<EOF
{
  "application_id": "$APP_ID",
  "type_id": "$MYSQL_TYPE_ID",
  "configuration": {
    "connection_host": "$mysql_host",
    "connection_port": $mysql_port,
    "connection_username": "$mysql_user",
    "connection_password": "$mysql_password",
    "connection_database": "$mysql_database",
    "connection_timeout": 5
  }
}
EOF
        )
        ;;
    
    5)
        echo -e "\n${YELLOW}Configurando Kong Connection${NC}"
        read -p "Host (default: localhost): " kong_host
        kong_host=${kong_host:-localhost}
        read -p "Port (default: 8001): " kong_port
        kong_port=${kong_port:-8001}
        
        metric_data=$(cat <<EOF
{
  "application_id": "$APP_ID",
  "type_id": "$KONG_TYPE_ID",
  "configuration": {
    "connection_host": "$kong_host",
    "connection_port": $kong_port,
    "connection_timeout": 5
  }
}
EOF
        )
        ;;
    
    6)
        echo -e "\n${YELLOW}Criando métricas de exemplo (valores padrão)${NC}"
        echo "Nota: Estas conexões provavelmente falharão se os serviços não estiverem rodando"
        
        # Criar todas as métricas de exemplo
        metrics_created=0
        
        # Redis
        if [ "$REDIS_TYPE_ID" != "null" ]; then
            echo "Criando Redis metric..."
            redis_metric=$(cat <<EOF
{
  "application_id": "$APP_ID",
  "type_id": "$REDIS_TYPE_ID",
  "configuration": {
    "connection_host": "localhost",
    "connection_port": 6379,
    "connection_db": 0,
    "connection_timeout": 5
  }
}
EOF
            )
            api_call "POST" "/application-metrics" "$redis_metric" > /dev/null
            ((metrics_created++))
        fi
        
        echo -e "${GREEN}✓ Created $metrics_created example metrics${NC}"
        echo ""
        echo "Aguarde 60 segundos para a coleta automática das métricas..."
        echo "Ou execute manualmente o serviço de monitoramento"
        echo ""
        echo "Para ver os resultados:"
        echo "  curl $API_URL/application-metric-values/application/$APP_ID/latest | jq"
        exit 0
        ;;
    
    0)
        echo "Saindo..."
        exit 0
        ;;
    
    *)
        echo -e "${RED}Opção inválida${NC}"
        exit 1
        ;;
esac

# Criar a métrica escolhida
echo ""
echo "4. Criando métrica de conexão..."
metric_response=$(api_call "POST" "/application-metrics" "$metric_data")
METRIC_ID=$(echo "$metric_response" | jq -r '.id')

if [ "$METRIC_ID" == "null" ] || [ -z "$METRIC_ID" ]; then
    echo -e "${RED}Failed to create metric${NC}"
    echo "$metric_response"
    exit 1
fi

echo -e "${GREEN}✓ Metric created: $METRIC_ID${NC}"
echo ""

# Aguardar coleta
echo "=========================================="
echo "Aguardando coleta automática..."
echo "=========================================="
echo "A coleta de métricas ocorre a cada 60 segundos (padrão)"
echo ""
echo "Aguardando 65 segundos..."

for i in {65..1}; do
    echo -ne "${YELLOW}$i segundos restantes...${NC}\r"
    sleep 1
done

echo -e "\n"

# Consultar valores coletados
echo "5. Consultando valores coletados..."
values_response=$(api_call "GET" "/application-metric-values/application/$APP_ID/latest")

echo ""
echo "=========================================="
echo "Resultados:"
echo "=========================================="
echo "$values_response" | jq '.'

# Analisar resultado
if echo "$values_response" | jq -e '.[] | select(.value.connection_status == "connected")' > /dev/null 2>&1; then
    echo ""
    echo -e "${GREEN}✓ Conexão estabelecida com sucesso!${NC}"
    
    # Mostrar métricas de performance
    conn_time=$(echo "$values_response" | jq -r '.[0].value.connection_time_ms')
    ping_time=$(echo "$values_response" | jq -r '.[0].value.connection_ping_time_ms')
    version=$(echo "$values_response" | jq -r '.[0].value.connection_version')
    
    echo ""
    echo "Métricas de Performance:"
    echo "  Tempo de conexão: ${conn_time}ms"
    echo "  Tempo de ping:    ${ping_time}ms"
    echo "  Versão:           $version"
elif echo "$values_response" | jq -e '.[] | select(.value.connection_status == "failed")' > /dev/null 2>&1; then
    echo ""
    echo -e "${RED}✗ Conexão falhou${NC}"
    
    error=$(echo "$values_response" | jq -r '.[0].value.connection_error')
    echo "Erro: $error"
elif echo "$values_response" | jq -e '.[] | select(.value.connection_status == "timeout")' > /dev/null 2>&1; then
    echo ""
    echo -e "${YELLOW}⚠ Timeout na conexão${NC}"
    echo "Verifique a conectividade de rede e se o serviço está acessível"
else
    echo ""
    echo -e "${YELLOW}⚠ Nenhum valor coletado ainda${NC}"
    echo "Verifique se o serviço de monitoramento está rodando"
fi

echo ""
echo "=========================================="
echo "IDs para referência:"
echo "=========================================="
echo "Project ID:     $PROJECT_ID"
echo "Application ID: $APP_ID"
echo "Metric ID:      $METRIC_ID"
echo ""
echo "Para consultar valores novamente:"
echo "  curl $API_URL/application-metric-values/application/$APP_ID/latest | jq"
echo ""
echo "Para consultar histórico da métrica:"
echo "  curl $API_URL/application-metric-values/application-metric/$METRIC_ID | jq"
echo ""

# Verificar dependências
check_dependencies

