# Testando a Coleta de Métricas de PVC

Este guia mostra como testar a nova funcionalidade de coleta de métricas de PVC.

## 🧪 Setup de Teste

### 1. Criar um PVC e Pod de Teste

```yaml
# test-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-data-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: test-app
  namespace: default
  labels:
    app: test-app
spec:
  containers:
  - name: nginx
    image: nginx:alpine
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: test-data-pvc
```

Aplicar:
```bash
kubectl apply -f test-pvc.yaml

# Aguardar pod ficar Running
kubectl wait --for=condition=Ready pod/test-app --timeout=60s
```

### 2. Criar Dados de Teste no PVC

```bash
# Criar alguns arquivos para gerar uso de disco
kubectl exec test-app -- sh -c 'dd if=/dev/zero of=/data/testfile bs=1M count=100'

# Verificar o uso
kubectl exec test-app -- df -h /data
```

## 📊 Configurar Monitoramento

### 1. Obter IDs Necessários

```bash
API_URL="http://localhost:8080/api/v1"

# Criar projeto de teste
PROJECT_ID=$(curl -s -X POST $API_URL/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Project",
    "description": "Testing PVC monitoring"
  }' | jq -r '.id')

echo "Project ID: $PROJECT_ID"

# Criar aplicação de teste
APP_ID=$(curl -s -X POST $API_URL/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"name\": \"Test App\",
    \"namespace\": \"default\"
  }" | jq -r '.id')

echo "Application ID: $APP_ID"

# Obter tipo de métrica PvcUsage
PVC_TYPE_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PvcUsage") | .id')

echo "PVC Type ID: $PVC_TYPE_ID"
```

### 2. Configurar Métrica de PVC

```bash
# Configuração com auto-discovery do mount path
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$PVC_TYPE_ID\",
    \"configuration\": {
      \"pvc_name\": \"test-data-pvc\",
      \"pod_label_selector\": \"app=test-app\"
    }
  }" | jq .

# Salvar o ID da métrica
METRIC_ID=$(curl -s $API_URL/application-metrics?application_id=$APP_ID | jq -r '.[0].id')
echo "Metric ID: $METRIC_ID"
```

## ✅ Verificar Coleta

### Método 1: Logs da Aplicação

```bash
# Seguir logs da aplicação
tail -f logs/k8s-monitoring.log | grep -E "PvcUsage|metric collection"

# Você deve ver algo como:
# {"level":"info","msg":"Starting metric collection"}
# {"level":"info","msg":"Collecting PvcUsage for test-app"}
# {"level":"info","msg":"Metric collection completed"}
```

### Método 2: API

```bash
# Aguardar 1-2 minutos para a primeira coleta
sleep 120

# Buscar valores da métrica
curl -s "$API_URL/application-metric-values?application_metric_id=$METRIC_ID" | jq .

# Resposta esperada:
# [
#   {
#     "id": "uuid",
#     "application_metric_id": "uuid",
#     "value": {
#       "pvc_capacity_bytes": 1073741824,
#       "pvc_used_bytes": 104857600,
#       "pvc_percent": 9.77
#     },
#     "created_at": "2024-01-15T10:30:00Z"
#   }
# ]
```

### Método 3: Teste Manual da Função

```bash
# Testar execução do df diretamente
kubectl exec test-app -- df -B1 /data

# Saída esperada:
# Filesystem           1B-blocks      Used Available Use% Mounted on
# /dev/sda1          1073741824 104857600 968884224  10% /data
```

## 🔍 Validações

### 1. Verificar Auto-Discovery do Mount Path

```bash
# Ver os volumes montados no pod
kubectl get pod test-app -o json | jq '.spec.volumes[] | select(.persistentVolumeClaim.claimName=="test-data-pvc")'

# Ver o mountPath
kubectl get pod test-app -o json | jq '.spec.containers[0].volumeMounts[] | select(.name=="data")'
```

### 2. Verificar Dados Coletados

```bash
# Buscar último valor
curl -s "$API_URL/application-metric-values?application_metric_id=$METRIC_ID&limit=1" | jq '.[-1].value'

# Comparar com df real
kubectl exec test-app -- df -B1 /data | tail -1

# Os valores devem ser consistentes
```

### 3. Testar com Aumento de Uso

```bash
# Adicionar mais dados
kubectl exec test-app -- sh -c 'dd if=/dev/zero of=/data/testfile2 bs=1M count=200'

# Aguardar próxima coleta (1 minuto)
sleep 70

# Verificar novo valor
curl -s "$API_URL/application-metric-values?application_metric_id=$METRIC_ID&limit=2" | jq '.[].value.pvc_percent'

# A porcentagem deve ter aumentado
```

## 🎯 Cenários de Teste

### Teste 1: Auto-Discovery Funciona

```bash
# Configurar sem pvc_mount_path
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$PVC_TYPE_ID\",
    \"configuration\": {
      \"pvc_name\": \"test-data-pvc\",
      \"pod_label_selector\": \"app=test-app\"
    }
  }"

# Deve funcionar - mount path descoberto automaticamente
```

### Teste 2: Mount Path Explícito

```bash
# Configurar com pvc_mount_path explícito
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$PVC_TYPE_ID\",
    \"configuration\": {
      \"pvc_name\": \"test-data-pvc\",
      \"pod_label_selector\": \"app=test-app\",
      \"pvc_mount_path\": \"/data\"
    }
  }"

# Deve funcionar - usa o path fornecido
```

### Teste 3: Multiple Containers

```yaml
# test-multi-container.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-multi
  labels:
    app: test-multi
spec:
  containers:
  - name: app
    image: nginx:alpine
    volumeMounts:
    - name: data
      mountPath: /data
  - name: sidecar
    image: busybox
    command: ["sleep", "3600"]
    volumeMounts:
    - name: data
      mountPath: /shared
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: test-data-pvc
```

```bash
kubectl apply -f test-multi-container.yaml

# Configurar especificando container
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$PVC_TYPE_ID\",
    \"configuration\": {
      \"pvc_name\": \"test-data-pvc\",
      \"pod_label_selector\": \"app=test-multi\",
      \"container_name\": \"app\"
    }
  }"

# Deve usar o container 'app' e descobrir mount path /data
```

### Teste 4: Erro - Pod Não Encontrado

```bash
# Configurar com label selector errado
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$PVC_TYPE_ID\",
    \"configuration\": {
      \"pvc_name\": \"test-data-pvc\",
      \"pod_label_selector\": \"app=nao-existe\"
    }
  }"

# Logs devem mostrar erro: "no pods found with label selector"
```

### Teste 5: Erro - PVC Não Montado

```bash
# Configurar com PVC que não está montado no pod
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$PVC_TYPE_ID\",
    \"configuration\": {
      \"pvc_name\": \"pvc-inexistente\",
      \"pod_label_selector\": \"app=test-app\"
    }
  }"

# Logs devem mostrar erro: "could not find mount path for PVC"
```

## 📊 Resultados Esperados

### Métricas Típicas (1Gi PVC com 100MB usado)

```json
{
  "pvc_capacity_bytes": 1073741824,    // ~1GB
  "pvc_used_bytes": 104857600,         // ~100MB
  "pvc_percent": 9.77                  // ~10%
}
```

### Conversões

```bash
# Converter bytes para GB
echo "scale=2; 1073741824 / 1024 / 1024 / 1024" | bc
# Output: 1.00

echo "scale=2; 104857600 / 1024 / 1024" | bc
# Output: 100.00 MB
```

## 🧹 Limpeza

```bash
# Deletar recursos de teste
kubectl delete pod test-app test-multi
kubectl delete pvc test-data-pvc

# Deletar dados da API (opcional)
curl -X DELETE "$API_URL/applications/$APP_ID"
curl -X DELETE "$API_URL/projects/$PROJECT_ID"
```

## 📈 Monitoramento Contínuo

```bash
# Script para monitorar mudanças em tempo real
watch -n 10 "curl -s '$API_URL/application-metric-values?application_metric_id=$METRIC_ID&limit=1' | jq '.[-1].value'"
```

## 🎓 Checklist de Validação

- [ ] PVC e Pod criados com sucesso
- [ ] Métrica configurada na API
- [ ] Primeira coleta realizada (aguardar 1-2 minutos)
- [ ] Valores de capacidade corretos
- [ ] Valores de uso corretos (comparar com df manual)
- [ ] Porcentagem calculada corretamente
- [ ] Auto-discovery do mount path funcionando
- [ ] Logs não mostram erros
- [ ] UI exibe métricas corretamente (se aplicável)
- [ ] Testes de erro funcionam como esperado

