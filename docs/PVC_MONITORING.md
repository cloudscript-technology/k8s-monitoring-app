# PVC Monitoring - Guia Completo

## 📊 Visão Geral

A aplicação monitora o uso real de disco dos **Persistent Volume Claims (PVCs)** executando o comando `df` dentro dos pods que montam esses volumes.

## 🔍 Como Funciona

### Processo de Coleta

1. **Descoberta Automática do Mount Path**
   - O sistema inspeciona o pod spec para encontrar o volume que referencia o PVC
   - Identifica automaticamente o `mountPath` onde o PVC está montado no container

2. **Execução do Comando df**
   - Seleciona um pod em execução que usa o PVC (via label selector)
   - Executa `df -B1 <mount_path>` dentro do container
   - Parseia a saída para obter bytes usados e disponíveis

3. **Cálculo de Métricas**
   - Capacidade total (da API do Kubernetes)
   - Espaço usado (do comando df)
   - Porcentagem de uso

### Fluxo Detalhado

```
┌─────────────────────────────────────────────────────────────┐
│ 1. API Request com configuração                             │
│    - pvc_name: "my-data-pvc"                                │
│    - pod_label_selector: "app=myapp"                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Descoberta do Mount Path (Auto)                          │
│    - Busca pods com label selector                          │
│    - Encontra volume que referencia o PVC                   │
│    - Identifica mountPath: "/data"                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Seleção do Pod                                           │
│    - Filtra pods com status "Running"                       │
│    - Seleciona o primeiro pod disponível                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Execução do df                                           │
│    - kubectl exec pod-123 -- df -B1 /data                   │
│    - Filesystem  1B-blocks    Used   Available  Use%        │
│    - /dev/sda1   10737418240  5368709120  ...   50%         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Armazenamento das Métricas                               │
│    {                                                         │
│      "pvc_capacity_bytes": 10737418240,                     │
│      "pvc_used_bytes": 5368709120,                          │
│      "pvc_percent": 50.0                                    │
│    }                                                         │
└─────────────────────────────────────────────────────────────┘
```

## ⚙️ Configuração

### Campos Obrigatórios

- **`pvc_name`**: Nome do PVC no Kubernetes
- **`pod_label_selector`**: Label selector para encontrar pods que usam o PVC

### Campos Opcionais

- **`container_name`**: Nome específico do container (padrão: primeiro container)
- **`pvc_mount_path`**: Path onde o PVC está montado (auto-descoberto se omitido)

### Exemplo de Configuração Mínima

```json
{
  "application_id": "app-uuid",
  "type_id": "pvcusage-type-uuid",
  "configuration": {
    "pvc_name": "postgres-data",
    "pod_label_selector": "app=postgres"
  }
}
```

### Exemplo de Configuração Completa

```json
{
  "application_id": "app-uuid",
  "type_id": "pvcusage-type-uuid",
  "configuration": {
    "pvc_name": "postgres-data",
    "pod_label_selector": "app=postgres,tier=database",
    "container_name": "postgres",
    "pvc_mount_path": "/var/lib/postgresql/data"
  }
}
```

## 🔐 Permissões RBAC Necessárias

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-monitoring-app
rules:
  # Ler PVCs
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list"]
  
  # Ler pods
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
  
  # Executar comandos em pods (para df)
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create"]
```

## 📋 Requisitos do Container

O container onde o `df` será executado deve:

1. ✅ Ter o comando `df` disponível
2. ✅ Ter o PVC montado no filesystem
3. ✅ Estar no estado "Running"

### Containers Compatíveis

- ✅ Imagens baseadas em Alpine Linux
- ✅ Imagens baseadas em Debian/Ubuntu
- ✅ Imagens baseadas em CentOS/RHEL
- ✅ Containers com BusyBox
- ⚠️ Containers "distroless" (não possuem shell/df)

### Workaround para Containers Distroless

Se seu container não tem `df`, você pode:

1. Adicionar um sidecar container com ferramentas de debug
2. Usar um init container para popular dados
3. Especificar um container diferente no mesmo pod que tenha `df`

## 📊 Formato das Métricas

### Resposta da API

```json
{
  "pvc_capacity_bytes": 10737418240,
  "pvc_used_bytes": 5368709120,
  "pvc_percent": 50.0
}
```

### Conversões Úteis

```
1 KB = 1,024 bytes
1 MB = 1,048,576 bytes
1 GB = 1,073,741,824 bytes
1 TB = 1,099,511,627,776 bytes
```

## 🔧 Troubleshooting

### Erro: "no pods found with label selector"

**Causa:** Nenhum pod encontrado com o label selector fornecido

**Solução:**
```bash
# Verifique os labels dos pods
kubectl get pods -n <namespace> --show-labels

# Ajuste o pod_label_selector na configuração
```

### Erro: "no running pods found"

**Causa:** Os pods existem mas não estão em estado Running

**Solução:**
```bash
# Verifique o status dos pods
kubectl get pods -n <namespace> -l app=myapp

# Aguarde os pods ficarem Running ou corrija problemas
```

### Erro: "could not find mount path for PVC"

**Causa:** O PVC não está montado em nenhum dos pods encontrados

**Solução:**
```bash
# Verifique quais PVCs estão montados
kubectl get pods -n <namespace> <pod-name> -o yaml | grep -A5 volumes

# Verifique se o PVC existe e está bound
kubectl get pvc -n <namespace>

# Se necessário, especifique o mount_path manualmente
```

### Erro: "failed to execute df command"

**Causa:** O container não possui o comando `df`

**Solução:**
```json
// Especifique um container diferente que tenha df
{
  "pvc_name": "my-pvc",
  "pod_label_selector": "app=myapp",
  "container_name": "sidecar-debug"  // Container com ferramentas
}
```

### Erro: "invalid df output"

**Causa:** O formato da saída do df não foi reconhecido

**Solução:**
```bash
# Execute df manualmente para verificar a saída
kubectl exec -n <namespace> <pod-name> -- df -B1 /data

# Verifique se o filesystem está montado corretamente
```

## 🎯 Casos de Uso

### Banco de Dados PostgreSQL

```json
{
  "pvc_name": "postgres-data-pvc",
  "pod_label_selector": "app=postgres",
  "container_name": "postgres"
}
```

### Aplicação com Multiple Containers

```json
{
  "pvc_name": "shared-data",
  "pod_label_selector": "app=myapp,component=backend",
  "container_name": "app",  // Container específico que monta o volume
  "pvc_mount_path": "/app/data"
}
```

### StatefulSet

```json
{
  "pvc_name": "data-redis-0",  // PVC do primeiro pod do StatefulSet
  "pod_label_selector": "app=redis,statefulset.kubernetes.io/pod-name=redis-0"
}
```

## 📈 Exemplo de Resposta Completa

```json
{
  "id": "metric-value-uuid",
  "application_metric_id": "app-metric-uuid",
  "value": {
    "pvc_capacity_bytes": 10737418240,
    "pvc_used_bytes": 5368709120,
    "pvc_percent": 50.0
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

## 🚀 Performance

- **Overhead**: Mínimo - apenas executa `df` uma vez por minuto
- **Impacto no Pod**: Negligível - comando df é muito leve
- **Latência**: < 500ms por coleta típica

## 🔒 Segurança

- ✅ Usa RBAC do Kubernetes - requer permissões explícitas
- ✅ Apenas comandos read-only (df)
- ✅ Não modifica dados do filesystem
- ✅ Não requer privilégios elevados no container

## 📚 Referências

- [Kubernetes PVC Documentation](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Pod Exec API](https://kubernetes.io/docs/tasks/debug-application-cluster/get-shell-running-container/)

