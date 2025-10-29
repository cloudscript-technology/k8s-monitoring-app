# Configuração de Autenticação OAuth 2.0 com Google

Este guia explica como configurar a autenticação OAuth 2.0 do Google para a aplicação K8s Monitoring App.

## 📋 Visão Geral

A aplicação agora requer autenticação via Google OAuth 2.0. Apenas usuários de domínios Google específicos podem acessar as métricas.

## 🎯 Características de Segurança

- ✅ Autenticação segura via Google OAuth 2.0
- ✅ Restrição por domínio de e-mail
- ✅ Sessões com expiração de 24 horas
- ✅ Cookies seguros (HttpOnly, Secure em produção)
- ✅ Proteção contra CSRF
- ✅ Logout seguro

## 🚀 Configuração Rápida

### Passo 1: Criar Credenciais OAuth no Google Cloud

1. Acesse o [Console do Google Cloud](https://console.cloud.google.com/apis/credentials)
2. Crie um novo projeto ou selecione um existente
3. Vá para **APIs e Serviços** > **Credenciais**
4. Clique em **Criar Credenciais** > **ID do cliente OAuth**
5. Escolha **Aplicativo da Web**
6. Configure:
   - **Nome**: K8s Monitoring App
   - **Origens JavaScript autorizadas**: 
     - `http://localhost:8080` (desenvolvimento local)
     - `https://seu-dominio.com` (produção)
   - **URIs de redirecionamento autorizados**:
     - `http://localhost:8080/auth/callback` (desenvolvimento)
     - `https://seu-dominio.com/auth/callback` (produção)
7. Clique em **Criar**
8. **Copie o Client ID e o Client Secret**

### Passo 2: Configurar Tela de Consentimento OAuth

1. Vá para **APIs e Serviços** > **Tela de consentimento OAuth**
2. Escolha **Interno** (apenas sua organização) ou **Externo** (público)
3. Preencha:
   - **Nome do app**: K8s Monitoring App
   - **E-mail de suporte**: seu e-mail
   - **Informações de contato**: seu e-mail
4. Em **Escopos**, adicione:
   - `userinfo.email`
   - `userinfo.profile`
5. Salve e continue

### Passo 3: Configurar Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```bash
# Autenticação OAuth
GOOGLE_CLIENT_ID=seu-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=seu-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
ALLOWED_GOOGLE_DOMAINS=suaempresa.com

# Banco de Dados (SQLite)
DB_PATH=./data/k8s_monitoring.db

# Outros
ENV=development
METRICS_RETENTION_DAYS=30
METRICS_COLLECTION_INTERVAL=60
```

### Passo 4: Executar a Aplicação

```bash
# Instalar dependências
go mod download

# Executar
go run cmd/main.go
```

### Passo 5: Testar

1. Acesse `http://localhost:8080`
2. Você será redirecionado para a página de login
3. Clique em "Sign in with Google"
4. Autentique com sua conta Google
5. Será redirecionado de volta para o dashboard

## 🔐 Variáveis de Ambiente

| Variável | Obrigatória | Descrição | Exemplo |
|----------|-------------|-----------|---------|
| `GOOGLE_CLIENT_ID` | Sim | ID do cliente OAuth do Google | `123456789-abc.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | Sim | Segredo do cliente OAuth | `GOCSPX-abc123def456` |
| `GOOGLE_REDIRECT_URL` | Sim | URL de callback (deve corresponder ao Console) | `http://localhost:8080/auth/callback` |
| `ALLOWED_GOOGLE_DOMAINS` | Não* | Lista de domínios permitidos (separados por vírgula) | `empresa.com,parceiro.com` |

\* Se `ALLOWED_GOOGLE_DOMAINS` não for definido, todos os domínios serão permitidos.

## 📝 Exemplo de Configuração para Produção

```bash
# Produção
ENV=production
GOOGLE_CLIENT_ID=123456789-xyz.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xyz789uvw012
GOOGLE_REDIRECT_URL=https://monitoring.suaempresa.com/auth/callback
ALLOWED_GOOGLE_DOMAINS=suaempresa.com

# Banco de Dados
DB_PATH=/data/k8s_monitoring.db
```

## 🔄 Fluxo de Autenticação

```
1. Usuário acessa uma rota protegida (ex: /)
   ↓
2. Middleware verifica se há sessão válida
   ↓
3. Sem sessão? → Redireciona para /auth/login
   ↓
4. Usuário clica em "Sign in with Google"
   ↓
5. Redirecionamento para Google OAuth
   ↓
6. Usuário autentica no Google
   ↓
7. Google redireciona para /auth/callback com código
   ↓
8. Aplicação troca código por token de acesso
   ↓
9. Aplicação obtém informações do usuário
   ↓
10. Valida o domínio do e-mail
    ↓
11. Domínio permitido? → Cria sessão
    ↓
12. Define cookie de sessão
    ↓
13. Redireciona para o dashboard
```

## 🚪 Rotas

### Rotas Públicas (Sem Autenticação)
- `/health` - Health check
- `/auth/login` - Página de login
- `/auth/google` - Inicia OAuth
- `/auth/callback` - Callback OAuth
- `/auth/logout` - Logout
- `/auth/error` - Página de erro
- `/static/*` - Arquivos estáticos

### Rotas Protegidas (Requer Autenticação)
- `/` - Dashboard
- `/api/v1/*` - Todos os endpoints da API REST
- `/api/ui/*` - Endpoints da API da UI

## 🗄️ Tabela de Sessões

A migration cria automaticamente a tabela `sessions`:

```sql
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_expiry DATETIME,
    created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
    expires_at DATETIME NOT NULL
);
```

## 🛠️ Solução de Problemas

### Erro: "OAuth not configured"

**Causa**: Variáveis de ambiente não configuradas

**Solução**:
```bash
# Verifique se as variáveis estão definidas
echo $GOOGLE_CLIENT_ID
echo $GOOGLE_CLIENT_SECRET

# Se não estiverem, defina-as
export GOOGLE_CLIENT_ID="seu-client-id"
export GOOGLE_CLIENT_SECRET="seu-secret"
```

### Erro: "Domain not allowed"

**Causa**: O domínio do e-mail do usuário não está na lista permitida

**Solução**:
```bash
# Adicione o domínio
export ALLOWED_GOOGLE_DOMAINS="empresa.com,outro-dominio.com"

# Ou permita todos os domínios (não recomendado para produção)
unset ALLOWED_GOOGLE_DOMAINS
```

### Erro: "Redirect URI mismatch"

**Causa**: A URI de redirecionamento não corresponde à configurada no Google Cloud Console

**Solução**:
1. Verifique a URL no arquivo `.env`:
   ```bash
   GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
   ```
2. Vá ao Console do Google Cloud
3. Edite seu ID do cliente OAuth
4. Adicione a URI exata em "URIs de redirecionamento autorizados"

### Erro: "Session not found or expired"

**Causa**: A sessão expirou (após 24 horas de inatividade)

**Solução**: Faça login novamente

## 🔒 Segurança

### Sessões
- IDs de sessão criptograficamente seguros (32 bytes aleatórios)
- Armazenadas no banco de dados PostgreSQL
- Expiração automática após 24 horas
- Renovação automática a cada requisição

### Cookies
- **HttpOnly**: Previne acesso via JavaScript (XSS)
- **Secure**: Apenas HTTPS em produção
- **SameSite=Lax**: Proteção contra CSRF
- Tempo de vida limitado

### Validação de Domínio
- Verifica o domínio do e-mail contra `ALLOWED_GOOGLE_DOMAINS`
- E-mail deve ser verificado pelo Google
- Suporta G Suite/Google Workspace (hosted domain)

## 📊 Monitoramento de Sessões

### Ver sessões ativas

```sql
-- Contar sessões ativas
SELECT COUNT(*) FROM sessions WHERE expires_at > NOW();

-- Listar sessões ativas
SELECT user_email, created_at, expires_at 
FROM sessions 
WHERE expires_at > NOW()
ORDER BY created_at DESC;
```

### Limpar sessões expiradas manualmente

```sql
DELETE FROM sessions WHERE expires_at <= NOW();
```

## 🚀 Deploy em Produção

### Checklist de Produção

- [ ] Criar credenciais OAuth específicas para produção
- [ ] Configurar URI de redirecionamento de produção
- [ ] Definir `ENV=production`
- [ ] Usar HTTPS (obrigatório para cookies seguros)
- [ ] Configurar `ALLOWED_GOOGLE_DOMAINS` apropriadamente
- [ ] Testar fluxo completo de autenticação
- [ ] Monitorar logs de autenticação

### Kubernetes/Docker

```yaml
# ConfigMap ou Secret
apiVersion: v1
kind: Secret
metadata:
  name: k8s-monitoring-oauth
type: Opaque
stringData:
  GOOGLE_CLIENT_ID: "seu-client-id"
  GOOGLE_CLIENT_SECRET: "seu-client-secret"
  GOOGLE_REDIRECT_URL: "https://monitoring.empresa.com/auth/callback"
  ALLOWED_GOOGLE_DOMAINS: "empresa.com"
```

```yaml
# Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-monitoring-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: k8s-monitoring-app:latest
        env:
        - name: ENV
          value: "production"
        envFrom:
        - secretRef:
            name: k8s-monitoring-oauth
```

## 📚 Recursos Adicionais

- [Documentação OAuth 2.0 do Google](https://developers.google.com/identity/protocols/oauth2)
- [Console do Google Cloud](https://console.cloud.google.com/)
- [Guia completo em inglês](OAUTH_SETUP.md)

## 🆘 Suporte

Para problemas:
1. Verifique os logs da aplicação para mensagens detalhadas de erro
2. Verifique se todas as variáveis de ambiente estão configuradas
3. Confirme que a configuração no Google Cloud Console está correta
4. Consulte o [guia completo de troubleshooting](OAUTH_SETUP.md#troubleshooting)

## ✅ Teste Rápido

```bash
# 1. Configurar variáveis
export GOOGLE_CLIENT_ID="seu-id"
export GOOGLE_CLIENT_SECRET="seu-secret"
export GOOGLE_REDIRECT_URL="http://localhost:8080/auth/callback"
export ALLOWED_GOOGLE_DOMAINS="empresa.com"
export DB_CONNECTION_STRING="postgres://user:senha@localhost:5432/k8s_monitoring"

# 2. Executar aplicação
go run cmd/main.go

# 3. Testar
# Abra o navegador em: http://localhost:8080
# Você deve ser redirecionado para a página de login
```

## 🎉 Pronto!

Agora sua aplicação está protegida com autenticação OAuth 2.0 do Google. Apenas usuários dos domínios autorizados poderão acessar as métricas.
