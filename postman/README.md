# Postman Collection - K8s Monitoring App

Esta collection do Postman contém todos os endpoints da API do K8s Monitoring App para facilitar testes e desenvolvimento.

## Arquivos

- `K8s-Monitoring-App.postman_collection.json` - Collection com todos os endpoints
- `K8s-Monitoring-App.postman_environment.json` - Environment com variáveis configuradas

## Como Importar

### No Postman Desktop

1. Abra o Postman
2. Clique em **Import** (ou use Ctrl+O / Cmd+O)
3. Selecione os dois arquivos JSON:
   - `K8s-Monitoring-App.postman_collection.json`
   - `K8s-Monitoring-App.postman_environment.json`
4. Clique em **Import**

### Via URL (se estiver em repositório Git)

1. Abra o Postman
2. Clique em **Import**
3. Cole a URL raw dos arquivos JSON do GitHub/GitLab
4. Clique em **Continue** e depois **Import**

## Estrutura da Collection

### 1. Health Check
- `GET /health` - Verificar status da aplicação

### 2. Projects
- `GET /api/v1/projects` - Listar projetos
- `GET /api/v1/projects/:id` - Obter projeto por ID
- `POST /api/v1/projects` - Criar projeto
- `PUT /api/v1/projects/:id` - Atualizar projeto
- `DELETE /api/v1/projects/:id` - Deletar projeto

### 3. Applications
- `GET /api/v1/applications` - Listar todas as aplicações
- `GET /api/v1/projects/:project_id/applications` - Listar aplicações por projeto
- `GET /api/v1/applications/:id` - Obter aplicação por ID
- `POST /api/v1/applications` - Criar aplicação
- `PUT /api/v1/applications/:id` - Atualizar aplicação
- `DELETE /api/v1/applications/:id` - Deletar aplicação

### 4. Metric Types
- `GET /api/v1/metric-types` - Listar tipos de métricas disponíveis
- `GET /api/v1/metric-types/:id` - Obter tipo de métrica por ID

### 5. Application Metrics
- `GET /api/v1/application-metrics` - Listar todas as métricas
- `GET /api/v1/applications/:application_id/metrics` - Listar métricas por aplicação
- `GET /api/v1/application-metrics/:id` - Obter métrica por ID
- `POST /api/v1/application-metrics` - Configurar métrica
- `PUT /api/v1/application-metrics/:id` - Atualizar métrica
- `DELETE /api/v1/application-metrics/:id` - Deletar métrica

### 6. 🆕 Metric Values (Collected Data)
- `GET /api/v1/applications/:application_id/latest-metrics` - Obter últimos valores coletados
- `GET /api/v1/application-metrics/:application_metric_id/values` - Histórico de valores
- `GET /api/v1/metric-values/:id` - Obter valor específico

Estes endpoints são **read-only** pois os valores são coletados automaticamente pelo cron job a cada minuto.

#### Exemplos de Configuração de Métricas

##### Health Check
```json
{
    "application_id": "{{application_id}}",
    "type_id": "{{healthcheck_type_id}}",
    "configuration": {
        "health_check_url": "http://service.namespace.svc.cluster.local:8080/health",
        "method": "GET",
        "expected_status": 200,
        "timeout_seconds": 10
    }
}
```

##### Pod Status
```json
{
    "application_id": "{{application_id}}",
    "type_id": "{{podstatus_type_id}}",
    "configuration": {
        "pod_label_selector": "app=myapp",
        "container_name": "main"
    }
}
```

##### Pod Memory Usage
```json
{
    "application_id": "{{application_id}}",
    "type_id": "{{podmemoryusage_type_id}}",
    "configuration": {
        "pod_label_selector": "app=myapp",
        "container_name": "main"
    }
}
```

##### Pod CPU Usage
```json
{
    "application_id": "{{application_id}}",
    "type_id": "{{podcpuusage_type_id}}",
    "configuration": {
        "pod_label_selector": "app=myapp",
        "container_name": "main"
    }
}
```

##### PVC Usage
```json
{
    "application_id": "{{application_id}}",
    "type_id": "{{pvcusage_type_id}}",
    "configuration": {
        "pvc_name": "my-pvc"
    }
}
```

##### Pod Active Nodes
```json
{
    "application_id": "{{application_id}}",
    "type_id": "{{podactivenodes_type_id}}",
    "configuration": {
        "pod_label_selector": "app=myapp"
    }
}
```

### 7. Complete Workflow
Pasta com sequência completa de requests para testar o fluxo:
1. Criar projeto
2. Criar aplicação
3. Obter tipos de métricas
4. Configurar health check
5. Listar métricas da aplicação
6. **Obter métricas coletadas** (aguardar 1-2 minutos após step 4)

## Variáveis de Ambiente

O environment inclui as seguintes variáveis:

### Variáveis Base
- `base_url` - URL base da API (padrão: `http://localhost:8080`)

### Variáveis Automáticas (preenchidas pelos scripts)
- `project_id` - ID do projeto criado
- `application_id` - ID da aplicação criada
- `application_metric_id` - ID da métrica criada
- `healthcheck_type_id` - ID do tipo HealthCheck
- `podstatus_type_id` - ID do tipo PodStatus
- `podmemoryusage_type_id` - ID do tipo PodMemoryUsage
- `podcpuusage_type_id` - ID do tipo PodCpuUsage
- `pvcusage_type_id` - ID do tipo PvcUsage
- `podactivenodes_type_id` - ID do tipo PodActiveNodes
- `workflow_project_id` - ID do projeto no workflow
- `workflow_application_id` - ID da aplicação no workflow
- `workflow_healthcheck_id` - ID do health check no workflow
- `metric_value_id` - ID de um valor de métrica coletado

## Como Usar

### Teste Rápido

1. **Selecione o Environment**
   - No canto superior direito, selecione "K8s Monitoring App - Local"

2. **Verifique a URL base**
   - Certifique-se que `base_url` está correto (localhost:8080 ou seu servidor)

3. **Execute Health Check**
   - Vá para "Health Check" → "Health Check"
   - Clique em **Send**
   - Deve retornar `200 OK` com resposta "ok"

### Workflow Completo

#### Opção 1: Executar pasta "Complete Workflow"

1. Clique com botão direito na pasta "Complete Workflow"
2. Selecione **Run folder**
3. Clique em **Run K8s Monitoring App**
4. Veja os resultados de todas as requisições em sequência

#### Opção 2: Executar manualmente

1. **Criar Projeto**
   ```
   POST /api/v1/projects
   ```
   O script de teste salva automaticamente o `project_id` no environment

2. **Listar Metric Types**
   ```
   GET /api/v1/metric-types
   ```
   O script salva automaticamente os IDs de todos os tipos

3. **Criar Aplicação**
   ```
   POST /api/v1/applications
   ```
   Usa automaticamente o `project_id` salvo

4. **Configurar Métricas**
   - Escolha um dos requests de criação de métrica
   - Ex: "Create HealthCheck Metric"
   - Ajuste a configuração conforme seu ambiente

5. **Verificar Métricas Configuradas**
   ```
   GET /api/v1/applications/{{application_id}}/metrics
   ```

6. **🆕 Visualizar Métricas Coletadas**
   ```
   GET /api/v1/applications/{{application_id}}/latest-metrics
   ```
   **Importante:** Aguarde 1-2 minutos após configurar as métricas para que o cron job colete os dados

## Scripts de Teste Automáticos

A collection inclui scripts JavaScript que:

### Scripts no "Create Project"
- Salva o `project_id` automaticamente quando projeto é criado

### Scripts no "List Metric Types"
- Extrai e salva todos os IDs de tipos de métricas automaticamente
- Facilita uso posterior nas configurações

### Scripts no "Create Application"
- Salva o `application_id` automaticamente

### Scripts no "Create Metric"
- Salva o `application_metric_id` automaticamente

## Customização

### Mudar URL Base

Para usar em outro ambiente (staging, produção):

1. No environment "K8s Monitoring App - Local"
2. Edite a variável `base_url`
3. Exemplos:
   - Local: `http://localhost:8080`
   - Port Forward: `http://localhost:8080`
   - Ingress: `https://monitoring.example.com`

### Criar Novo Environment

1. Duplique o environment existente
2. Renomeie para "K8s Monitoring App - Production"
3. Altere `base_url` para URL de produção
4. Adicione autenticação se necessário

### Adicionar Autenticação

Se sua API tiver autenticação:

1. Vá para a collection
2. Clique em "..." → "Edit"
3. Vá para a aba "Authorization"
4. Configure o tipo (Bearer Token, Basic Auth, etc.)
5. Use variáveis do environment para tokens:
   ```
   {{auth_token}}
   ```

## Troubleshooting

### "Could not get response"

**Problema**: Aplicação não está rodando

**Solução**:
```bash
# Verificar se a aplicação está rodando
curl http://localhost:8080/health

# Ou inicie a aplicação
go run cmd/main.go
```

### "project not found" ao criar aplicação

**Problema**: `project_id` não foi salvo ou é inválido

**Solução**:
1. Execute primeiro "Create Project"
2. Verifique na aba "Console" se o ID foi salvo
3. Ou defina manualmente no environment

### Variáveis não estão sendo substituídas

**Problema**: Environment não está selecionado

**Solução**:
1. Verifique no canto superior direito
2. Selecione "K8s Monitoring App - Local"
3. A variável deve aparecer em laranja: `{{base_url}}`

### Erro 404 em endpoints

**Problema**: URL incorreta ou aplicação não iniciada

**Solução**:
1. Verifique `base_url` no environment
2. Confirme que a aplicação está rodando
3. Verifique os logs da aplicação

## Exportar Resultados

Para compartilhar resultados dos testes:

1. Execute a pasta "Complete Workflow"
2. Clique em **Export Results**
3. Escolha formato (JSON ou HTML)
4. Compartilhe o arquivo

## Exemplos de Uso

### Teste Básico
```
1. Health Check
2. List Metric Types (para preencher IDs)
3. Create Project
4. Create Application
5. Create HealthCheck Metric
6. List Application Metrics
```

### Teste Completo de Monitoramento
```
1. Create Project
2. Create Application
3. List Metric Types
4. Create HealthCheck Metric
5. Create PodStatus Metric
6. Create PodMemoryUsage Metric
7. Create PodCpuUsage Metric
8. List Application Metrics (ver todas configuradas)
```

### Atualização de Configuração
```
1. List Application Metrics
2. Get Application Metric by ID (copiar ID)
3. Update Application Metric (alterar configuração)
4. Get Application Metric by ID (verificar mudança)
```

## Integração com CI/CD

Você pode executar a collection via Newman (CLI do Postman):

```bash
# Instalar Newman
npm install -g newman

# Executar collection
newman run K8s-Monitoring-App.postman_collection.json \
  -e K8s-Monitoring-App.postman_environment.json \
  --reporters cli,json

# Com variáveis customizadas
newman run K8s-Monitoring-App.postman_collection.json \
  -e K8s-Monitoring-App.postman_environment.json \
  --env-var "base_url=https://staging.example.com"
```

## Recursos Adicionais

- [Documentação Completa da API](../docs/API.md)
- [Guia de Desenvolvimento Local](../docs/LOCAL_DEVELOPMENT.md)
- [Exemplos de Uso](../docs/EXAMPLES.md)

## Suporte

Para problemas ou dúvidas:
- Verifique os logs da aplicação
- Consulte a documentação da API
- Crie uma issue no repositório

