# Desenvolvimento Local - K8s Monitoring App

Este guia explica como executar e testar a aplicação localmente no seu Mac usando o kubeconfig local.

## Pré-requisitos

1. **Go 1.24+** instalado
2. **SQLite** (embutido; nenhum serviço externo necessário)
3. **Acesso a um cluster Kubernetes** (pode ser Minikube, Kind, Docker Desktop, ou um cluster remoto)
4. **kubectl** configurado e funcionando
5. **metrics-server** instalado no cluster

## Configuração do Ambiente

### 1. Verificar Kubeconfig

Primeiro, verifique se seu kubeconfig está funcionando:

```bash
# Verificar contexto atual
kubectl config current-context

# Listar todos os contextos disponíveis
kubectl config get-contexts

# Testar acesso ao cluster
kubectl get nodes
kubectl get pods --all-namespaces
```

### 2. Configurar Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```bash
# Database Configuration (SQLite)
DB_PATH=./data/k8s_monitoring.db

# Kubernetes Configuration (OPCIONAL - se não definir, usa ~/.kube/config automaticamente)
# KUBECONFIG=/Users/seu-usuario/.kube/config

# Application Configuration
LOG_LEVEL=debug

# APM (Opcional - pode deixar desabilitado para desenvolvimento)
ELASTIC_APM_ACTIVE=false
```

### 3. Preparar SQLite

```bash
# Criar diretório de dados (se ainda não existir)
mkdir -p ./data

# Opcional: criar arquivo do banco (a aplicação cria automaticamente se não existir)
touch ./data/k8s_monitoring.db
```

## Como Funciona a Detecção de Kubeconfig

A aplicação detecta automaticamente o ambiente e usa a configuração apropriada:

### Ordem de Prioridade:

1. **KUBECONFIG env var** - Se definida, usa o arquivo especificado
2. **~/.kube/config** - Padrão do kubectl
3. **In-cluster config** - Para quando roda dentro do cluster

```go
// Exemplo de uso no código
// internal/k8s/client.go

// Priority:
// 1. KUBECONFIG environment variable
// 2. ~/.kube/config
// 3. In-cluster config
```

## Executando a Aplicação Localmente

### Opção 1: Usando go run

```bash
# Carregar variáveis de ambiente e executar
export $(cat .env | xargs) && go run cmd/main.go
```

### Opção 2: Build e Executar

```bash
# Build
go build -o k8s-monitoring-app cmd/main.go

# Executar
./k8s-monitoring-app
```

### Opção 3: Usando source para variáveis de ambiente

```bash
# Exportar variáveis
set -a
source .env
set +a

# Executar
go run cmd/main.go
```

### Opção 4: Script de desenvolvimento

Crie um arquivo `run-local.sh`:

```bash
#!/bin/bash

# Load environment variables
export DB_PATH=./data/k8s_monitoring.db
export LOG_LEVEL=debug

# Optional: specify kubeconfig explicitly
# export KUBECONFIG=/Users/seu-usuario/.kube/config

# Build and run
go build -o k8s-monitoring-app cmd/main.go && ./k8s-monitoring-app
```

Torne executável e rode:

```bash
chmod +x run-local.sh
./run-local.sh
```

## Verificando a Conexão com Kubernetes

Ao iniciar a aplicação, você verá uma das seguintes mensagens:

```
# Se usando kubeconfig local:
Using kubeconfig from: /Users/seu-usuario/.kube/config

# Se usando in-cluster config:
Using in-cluster kubernetes config
```

## Testando a Aplicação

### 1. Verificar Health Check

```bash
curl http://localhost:8080/health
```

### 2. Criar um Projeto

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Development",
    "description": "Local development project"
  }'
```

### 3. Listar Metric Types

```bash
curl http://localhost:8080/api/v1/metric-types | jq '.'
```

### 4. Testar com Pods Reais do Cluster

```bash
# Listar namespaces disponíveis
kubectl get namespaces

# Listar pods em um namespace
kubectl get pods -n default -o wide

# Exemplo: Registrar aplicação baseada em pods reais
PROJECT_ID="<uuid-do-projeto>"

curl -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"name\": \"kube-system-metrics\",
    \"description\": \"Monitoring kube-system pods\",
    \"namespace\": \"kube-system\"
  }"
```

## Testando Diferentes Clusters

### Trocar de Contexto Kubernetes

```bash
# Listar contextos
kubectl config get-contexts

# Trocar contexto
kubectl config use-context <nome-do-contexto>

# Reiniciar a aplicação - ela usará o novo contexto automaticamente
```

### Usar Kubeconfig Específico

```bash
# Definir KUBECONFIG para um arquivo específico
export KUBECONFIG=/path/to/specific/kubeconfig.yaml

# Executar a aplicação
go run cmd/main.go
```

### Exemplo com Múltiplos Clusters

```bash
# Produção
export KUBECONFIG=~/.kube/config-production
go run cmd/main.go

# Staging
export KUBECONFIG=~/.kube/config-staging
go run cmd/main.go

# Local (Minikube/Kind)
export KUBECONFIG=~/.kube/config
go run cmd/main.go
```

## Desenvolvimento com Minikube

```bash
# Iniciar Minikube
minikube start

# Habilitar metrics-server
minikube addons enable metrics-server

# Verificar
kubectl top nodes
kubectl top pods -A

# Executar aplicação (usará automaticamente o contexto do Minikube)
go run cmd/main.go
```

## Desenvolvimento com Kind (Kubernetes in Docker)

```bash
# Criar cluster Kind
kind create cluster --name monitoring-dev

# Instalar metrics-server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Configurar metrics-server para desenvolvimento (sem TLS)
kubectl patch deployment metrics-server -n kube-system --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

# Executar aplicação
go run cmd/main.go
```

## Debug e Troubleshooting

### Verificar Logs Detalhados

```bash
# Com log level debug
export LOG_LEVEL=debug
go run cmd/main.go
```

### Verificar Conexão com Kubernetes

```bash
# Criar um script de teste test-k8s.go
cat > test-k8s.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "k8s-monitoring-app/internal/k8s"
)

func main() {
    client, err := k8s.NewClient()
    if err != nil {
        fmt.Printf("Error creating client: %v\n", err)
        return
    }

    pods, err := client.GetPodsByLabelSelector(context.Background(), "kube-system", "")
    if err != nil {
        fmt.Printf("Error listing pods: %v\n", err)
        return
    }

    fmt.Printf("Found %d pods in kube-system\n", len(pods.Items))
    for _, pod := range pods.Items {
        fmt.Printf("- %s (Phase: %s)\n", pod.Name, pod.Status.Phase)
    }
}
EOF

# Executar teste
go run test-k8s.go

# Limpar
rm test-k8s.go
```

### Verificar Acesso ao Metrics Server

```bash
# Testar diretamente com kubectl
kubectl top nodes
kubectl top pods -A

# Se não funcionar, verificar se metrics-server está rodando
kubectl get deployment metrics-server -n kube-system
kubectl logs -n kube-system deployment/metrics-server
```

### Problemas Comuns

#### 1. "failed to create in-cluster config"

**Solução**: Normal em desenvolvimento local. A aplicação automaticamente tentará usar o kubeconfig local.

#### 2. "failed to get pod metrics"

**Solução**: Instalar/configurar metrics-server:

```bash
# Instalar
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Para ambientes locais (Minikube/Kind), pode precisar desabilitar TLS
kubectl patch deployment metrics-server -n kube-system --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'
```

#### 3. "connection refused" ao banco de dados

**Solução**: Verificar se PostgreSQL está rodando:

```bash
# Docker
docker ps | grep postgres

# Homebrew
brew services list | grep postgresql

# Testar conexão
psql -h localhost -U monitoring -d k8s_monitoring -c "SELECT 1;"
```

#### 4. Cron job não está coletando métricas

**Solução**: 
- Verificar se há aplicações e métricas cadastradas
- Ver logs da aplicação
- Por padrão, coleta a cada 1 minuto

```bash
# Ver logs
tail -f /tmp/k8s-monitoring.log

# Ou se estiver rodando no terminal, acompanhe a saída
```

## Hot Reload durante Desenvolvimento

Para recarregar automaticamente ao fazer mudanças no código:

```bash
# Instalar air (hot reload tool)
go install github.com/cosmtrek/air@latest

# Criar .air.toml na raiz do projeto
cat > .air.toml << 'EOF'
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/main.go"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor"]
delay = 1000

[log]
time = true
EOF

# Executar com hot reload
air
```

## Simulando Ambiente de Produção Local

Para testar o comportamento in-cluster localmente:

```bash
# Build da imagem Docker
docker build -t k8s-monitoring-app:dev -f Dockerfile.goreleaser .

# Deploy local com Kind
kind load docker-image k8s-monitoring-app:dev --name monitoring-dev

# Deploy com Helm
helm install k8s-monitoring-app ./chart \
  --set image.tag=dev \
  --set image.pullPolicy=Never
```

## Limpeza

```bash
# Parar aplicação (Ctrl+C)

# Limpar banco de dados
docker exec -it postgres-monitoring psql -U monitoring -d k8s_monitoring -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Parar e remover PostgreSQL Docker
docker stop postgres-monitoring
docker rm postgres-monitoring

# Ou com docker-compose
docker-compose down -v
```

## Próximos Passos

Depois de configurar o ambiente local, você pode:

1. Adicionar novos tipos de métricas
2. Criar APIs para visualização de dados
3. Implementar alertas
4. Desenvolver dashboard web

## Dicas de Produtividade

1. **Use alias para comandos frequentes**:
```bash
alias dev-db='docker-compose up -d postgres'
alias dev-run='export $(cat .env | xargs) && go run cmd/main.go'
alias dev-api='curl http://localhost:8080'
```

2. **Mantenha um terminal com logs do banco**:
```bash
docker-compose logs -f postgres
```

3. **Use extensões do VS Code**:
- Go
- PostgreSQL
- Kubernetes
- REST Client (para testar APIs)

4. **Crie snippets para requests comuns**:
```bash
# Salvar em requests.http (REST Client extension)
### Health Check
GET http://localhost:8080/health

### List Projects
GET http://localhost:8080/api/v1/projects

### Create Project
POST http://localhost:8080/api/v1/projects
Content-Type: application/json

{
  "name": "Test Project",
  "description": "Testing"
}
```
