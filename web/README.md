# Web UI - K8s Monitoring App

Interface web moderna construída com **HTMX** e **Go Templates** para visualização em tempo real das métricas de aplicações Kubernetes.

## 🎨 Características

### ✨ Interface Moderna
- Design limpo e intuitivo inspirado no mockup fornecido
- Cards organizados por projeto
- Visualização clara de métricas por aplicação
- Indicadores visuais de status (cores e badges)

### 🔄 Atualização em Tempo Real
- **Auto-refresh a cada 10 segundos** usando HTMX polling
- Carregamento progressivo de componentes
- Animações suaves de transição
- Indicador de status de conexão

### 📊 Métricas Visualizadas

Por aplicação, você pode ver:

1. **Pods** 
   - Status visual de cada pod (running/pending/failed)
   - Contador de pods ativos
   - Cores indicativas do estado

2. **Nodes**
   - Badges dos nodes onde os pods estão rodando
   - Identificação visual de cada node

3. **Health Check**
   - Status: ok/error/?
   - URL do health check endpoint
   - Indicador colorido de saúde

4. **Memory (Mem)**
   - Percentual de uso
   - Barra de progresso visual
   - Cores gradientes (verde → amarelo → vermelho)

5. **CPU**
   - Percentual de uso
   - Barra de progresso visual
   - Cores gradientes (verde → amarelo → vermelho)

6. **Disk (PVC)**
   - Percentual de uso do volume
   - Barra de progresso visual
   - Cores gradientes (verde → amarelo → vermelho)

## 🗂️ Estrutura de Arquivos

```
web/
├── static/
│   └── css/
│       └── style.css          # Estilos CSS customizados
├── templates/
│   ├── layout.html            # Template base com header/footer
│   ├── project-card.html      # Card de projeto com aplicações
│   └── application-metrics.html  # Métricas detalhadas da aplicação
└── README.md                  # Esta documentação
```

## 🚀 Como Funciona

### Arquitetura

```
Browser (HTMX)
    ↓
    GET /                           → Dashboard principal (layout.html)
    ↓
    GET /api/ui/projects            → Lista projetos e aplicações
    ↓
    GET /api/ui/applications/:id/metrics  → Métricas de cada aplicação
```

### Fluxo de Carregamento

1. **Página inicial** (`/`)
   - Carrega o layout base
   - HTMX faz request para `/api/ui/projects`

2. **Lista de projetos**
   - Servidor retorna HTML parcial com cards de projetos
   - Cada card contém placeholders para aplicações

3. **Métricas das aplicações**
   - HTMX faz requests paralelos para cada aplicação
   - Cada aplicação carrega suas métricas independentemente
   - **Auto-refresh** a cada 10 segundos

### HTMX Features Utilizadas

```html
<!-- Auto-load e polling -->
hx-get="/api/ui/projects"
hx-trigger="load, every 10s"
hx-swap="innerHTML"

<!-- Carregar métricas de aplicação -->
hx-get="/api/ui/applications/{{ .ID }}/metrics"
hx-trigger="load, every 10s"
hx-swap="innerHTML"
hx-target="this"
```

## 🎨 Design System

### Cores

```css
--primary-color: #2563eb    (Azul - títulos e labels)
--success-color: #10b981    (Verde - status ok, nodes)
--warning-color: #f59e0b    (Amarelo - avisos, health ok)
--danger-color: #ef4444     (Vermelho - erros)
```

### Componentes

#### Project Card
- Borda preta grossa (2px)
- Título em itálico
- Fundo branco com sombra

#### Application Card
- Borda preta grossa (2px)
- Layout em grid responsivo
- Divisão entre métricas principais (esquerda) e side panel (direita)

#### Side Panel (Mem/CPU/Disk)
- Borda verde
- Barras de progresso com gradiente
- Labels em azul
- Valores em percentual

#### Pod Status
- Boxes numerados (1, 2, 3...)
- Verde: Running
- Amarelo: Pending
- Cinza: Unknown

#### Health Indicator
- Badge circular "ok" (amarelo)
- Badge circular "error" (vermelho)
- Badge circular "?" (cinza)

## 📝 Customização

### Ajustar Intervalo de Refresh

No template, altere o `every 10s`:

```html
<div hx-trigger="load, every 30s">  <!-- 30 segundos -->
```

### Adicionar Novos Tipos de Métricas

1. Adicione a lógica no handler `GetApplicationMetrics`
2. Acesse a métrica no template:
```go
{{ $newMetric := index .MetricsByType "NewMetricType" }}
```

3. Renderize no template:
```html
<div class="metric-box">
    {{ if $newMetric }}
        {{ if $newMetric.LatestValue }}
            <!-- Exibir valor -->
        {{ end }}
    {{ end }}
</div>
```

### Personalizar CSS

Edite `web/static/css/style.css`:

```css
/* Suas customizações */
.project-card {
    border-color: var(--primary-color);
    border-width: 3px;
}
```

## 🔧 Desenvolvimento

### Requisitos

- Go 1.21+
- Servidor rodando na porta 8080
- Banco de dados PostgreSQL configurado

### Executar Localmente

```bash
# Na raiz do projeto
go run cmd/main.go

# Acesse
http://localhost:8080
```

### Hot Reload (Opcional)

Para desenvolvimento, use `air` para hot reload:

```bash
# Instalar air
go install github.com/cosmtrek/air@latest

# Executar
air
```

## 🌐 Endpoints

### Web UI

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/` | GET | Dashboard principal |
| `/static/*` | GET | Arquivos estáticos (CSS, JS, imagens) |

### API UI (HTMX Partials)

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/api/ui/projects` | GET | Lista todos projetos com aplicações |
| `/api/ui/applications/:id/metrics` | GET | Métricas de uma aplicação específica |

### REST API

Todas as rotas da REST API continuam disponíveis em `/api/v1/*` para integração programática.

## 📱 Responsividade

O layout é responsivo e se adapta a diferentes tamanhos de tela:

- **Desktop**: Layout em grid com side panel à direita
- **Mobile**: Side panel move para baixo, layout vertical

Breakpoint: `768px`

## 🎯 Boas Práticas

### Performance

1. **Lazy Loading**: Cada aplicação carrega suas métricas independentemente
2. **Polling Inteligente**: Apenas recarrega o necessário
3. **CSS Otimizado**: Uso de variáveis CSS e classes reutilizáveis

### UX

1. **Loading States**: Spinners durante carregamento
2. **Transições Suaves**: Animações CSS para melhor feedback
3. **Status Indicator**: Indicador pulsante mostrando conexão ativa
4. **Cores Semânticas**: Verde (ok), Amarelo (warning), Vermelho (erro)

### Acessibilidade

1. **Contraste de cores** adequado
2. **Textos legíveis** com tamanhos apropriados
3. **Hierarquia visual** clara

## 🐛 Troubleshooting

### Métricas não aparecem

1. Verifique se as aplicações estão cadastradas
2. Confirme que as métricas estão sendo coletadas (ver logs)
3. Aguarde 1-2 minutos para primeira coleta

### Auto-refresh não funciona

1. Verifique se o HTMX está carregado (console do browser)
2. Confirme que o servidor está respondendo em `/api/ui/*`
3. Veja erros no console do navegador (F12)

### Layout quebrado

1. Verifique se `/static/css/style.css` está acessível
2. Limpe o cache do navegador (Ctrl+F5)
3. Confirme que o servidor está servindo arquivos estáticos

## 🔮 Próximas Melhorias

- [ ] Filtros por projeto/aplicação
- [ ] Histórico de métricas (gráficos)
- [ ] Notificações de alertas
- [ ] Modo escuro (dark mode)
- [ ] Exportar métricas (CSV/JSON)
- [ ] Configuração de thresholds visuais
- [ ] Dashboard customizável (drag-and-drop)

## 📚 Referências

- [HTMX Documentation](https://htmx.org/)
- [Go Templates](https://pkg.go.dev/html/template)
- [Echo Framework](https://echo.labstack.com/)

---

Desenvolvido com ❤️ usando HTMX + Go

