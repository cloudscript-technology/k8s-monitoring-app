# Quick Start - Web UI

Guia rápido para começar a usar a interface web do K8s Monitoring App.

## 🚀 Passo 1: Iniciar a Aplicação

### Opção A: Localmente (Desenvolvimento)

```bash
# IMPORTANTE: Execute a partir da raiz do projeto!
cd /path/to/k8s-monitoring-app

# Configurar variáveis de ambiente (opcional)
export KUBECONFIG=~/.kube/config
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=monitoring
export DB_PASSWORD=monitoring
export DB_NAME=k8s_monitoring

# Iniciar banco de dados (se usando docker-compose)
docker-compose up -d postgres

# Opção 1: Usar o script helper (RECOMENDADO)
./run.sh

# Opção 2: Executar manualmente (certifique-se de estar na raiz!)
go run cmd/main.go
```

⚠️ **ATENÇÃO**: O aplicativo deve ser executado a partir do **diretório raiz do projeto**, não de dentro de `cmd/`. Caso contrário, os templates não serão encontrados.

### Opção B: No Kubernetes

```bash
# Deploy da aplicação
kubectl apply -f chart/

# Port-forward para acessar localmente
kubectl port-forward service/k8s-monitoring-app 8080:8080
```

## 🌐 Passo 2: Acessar a Interface

Abra seu navegador:

```
http://localhost:8080
```

Você verá o dashboard principal!

## 📋 Passo 3: Cadastrar Dados (Primeira Vez)

Se é a primeira vez, você precisa cadastrar projetos e aplicações via API.

### Via Postman

1. Importe a collection: `postman/K8s-Monitoring-App.postman_collection.json`
2. Importe o environment: `postman/K8s-Monitoring-App.postman_environment.json`
3. Execute o workflow "Complete Workflow"

### Via cURL

```bash
# 1. Criar um projeto
PROJECT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production",
    "description": "Production environment applications"
  }')

PROJECT_ID=$(echo $PROJECT_RESPONSE | jq -r '.id')
echo "Project ID: $PROJECT_ID"

# 2. Criar uma aplicação
APP_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"api-opb\",
    \"description\": \"API OPB\",
    \"namespace\": \"production\",
    \"project_id\": \"$PROJECT_ID\"
  }")

APP_ID=$(echo $APP_RESPONSE | jq -r '.id')
echo "Application ID: $APP_ID"

# 3. Listar tipos de métricas disponíveis
curl -s http://localhost:8080/api/v1/metric-types | jq '.[] | {id, name}'

# 4. Configurar Health Check
HEALTH_TYPE_ID="<ID do tipo HealthCheck>"
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$HEALTH_TYPE_ID\",
    \"configuration\": {
      \"health_check_url\": \"http://api-opb.production.svc.cluster.local:8080/health\",
      \"method\": \"GET\",
      \"expected_status\": 200,
      \"timeout_seconds\": 10
    }
  }"

# 5. Configurar Pod Status
POD_STATUS_TYPE_ID="<ID do tipo PodStatus>"
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$POD_STATUS_TYPE_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=api-opb\",
      \"container_name\": \"api-opb\"
    }
  }"

# 6. Configurar CPU Usage
CPU_TYPE_ID="<ID do tipo PodCpuUsage>"
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$CPU_TYPE_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=api-opb\",
      \"container_name\": \"api-opb\"
    }
  }"

# 7. Configurar Memory Usage
MEMORY_TYPE_ID="<ID do tipo PodMemoryUsage>"
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$MEMORY_TYPE_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=api-opb\",
      \"container_name\": \"api-opb\"
    }
  }"

# 8. Configurar PVC Usage (opcional)
PVC_TYPE_ID="<ID do tipo PvcUsage>"
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$PVC_TYPE_ID\",
    \"configuration\": {
      \"pvc_name\": \"api-opb-data\",
      \"pod_label_selector\": \"app=api-opb\"
    }
  }"

# Nota: O sistema descobre automaticamente o mount path do PVC no pod
# Se quiser especificar manualmente, adicione "pvc_mount_path": "/data"
```

## ⏱️ Passo 4: Aguardar Coleta de Métricas

As métricas são coletadas a cada minuto pelo cron job.

```bash
# Acompanhar logs de coleta
tail -f logs/k8s-monitoring.log | grep "metric collection"

# Você verá:
# "Starting metric collection"
# ...
# "Metric collection completed"
```

Aguarde 1-2 minutos para a primeira coleta.

## 🎉 Passo 5: Visualizar no Dashboard

Recarregue a página ou aguarde o auto-refresh (10 segundos):

```
http://localhost:8080
```

Você verá:

```
┌────────────────────────────────────────────────┐
│ Production                                     │
│                                                │
│ ┌────────────────────────────────────────────┐│
│ │ api-opb         [n61][n62] [ok]   mem 45% ││
│ │                                    cpu 0.1%││
│ │ pods: [1][2][3]                   disk 23% ││
│ │ health: http://...                         ││
│ └────────────────────────────────────────────┘│
└────────────────────────────────────────────────┘
```

## 🔍 Troubleshooting

### Não vejo nenhum projeto

**Problema:** Dashboard vazio ou "Loading projects..."

**Solução:**
```bash
# Verificar se há projetos cadastrados
curl http://localhost:8080/api/v1/projects

# Se retornar [], cadastre um projeto (ver Passo 3)
```

### Aplicação aparece mas sem métricas

**Problema:** Card da aplicação vazio ou com "?"

**Soluções:**

1. **Verificar se métricas estão configuradas:**
```bash
curl http://localhost:8080/api/v1/applications/$APP_ID/metrics
```

2. **Aguardar coleta:**
- Métricas são coletadas a cada minuto
- Aguarde 1-2 minutos após configurar

3. **Verificar logs de coleta:**
```bash
tail -f logs/k8s-monitoring.log | grep "error"
```

### Métricas sempre aparecem como "-"

**Problema:** Side panel mostra `-` para mem/cpu/disk

**Causas possíveis:**

1. **Metrics-server não instalado:**
```bash
kubectl get deployment metrics-server -n kube-system
```

2. **Pod não encontrado:**
```bash
# Verificar label selector
kubectl get pods -n production -l app=api-opb
```

3. **PVC não existe:**
```bash
kubectl get pvc -n production
```

Ver guia completo: [docs/TROUBLESHOOTING.md](TROUBLESHOOTING.md)

### Health check sempre "error"

**Problema:** Badge de health sempre vermelho

**Soluções:**

1. **Verificar URL:**
```bash
# Testar do cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://api-opb.production.svc.cluster.local:8080/health
```

2. **Ajustar configuração:**
```bash
# Corrigir URL na configuração da métrica
curl -X PUT http://localhost:8080/api/v1/application-metrics/$METRIC_ID \
  -H "Content-Type: application/json" \
  -d '{
    "configuration": {
      "health_check_url": "URL_CORRETA",
      "method": "GET",
      "expected_status": 200,
      "timeout_seconds": 10
    }
  }'
```

## 🎨 Customizar Refresh Rate

Por padrão, a interface atualiza a cada 10 segundos.

Para alterar:

1. Edite `web/templates/layout.html`:

```html
<!-- Alterar de "every 10s" para outro valor -->
<div class="projects-grid" 
     hx-get="/api/ui/projects" 
     hx-trigger="load, every 30s"  <!-- 30 segundos -->
     hx-swap="innerHTML">
```

2. Edite `web/templates/project-card.html`:

```html
<div class="application-card" 
     hx-get="/api/ui/applications/{{ .ID }}/metrics" 
     hx-trigger="load, every 30s"  <!-- 30 segundos -->
     ...>
```

## 📊 Dados de Exemplo

Para popular com dados de exemplo:

```bash
# Script de exemplo (criar arquivo populate.sh)
#!/bin/bash

# Ver postman/README.md para workflow completo
# Ou usar: postman/K8s-Monitoring-App.postman_collection.json
```

## 📚 Próximos Passos

1. ✅ Explorar a interface
2. 📖 Ler documentação completa: [web/README.md](../web/README.md)
3. 🎨 Ver guia visual: [docs/WEB_UI.md](WEB_UI.md)
4. 🔧 Configurar mais aplicações
5. 📈 Analisar tendências de métricas

## 🆘 Precisa de Ajuda?

- 📖 [Documentação da API](API.md)
- 🐛 [Troubleshooting](TROUBLESHOOTING.md)
- 💡 [Exemplos](EXAMPLES.md)
- 🧪 [Postman Collection](../postman/README.md)

---

**Pronto! Agora você tem um dashboard completo para monitorar suas aplicações Kubernetes! 🚀**

