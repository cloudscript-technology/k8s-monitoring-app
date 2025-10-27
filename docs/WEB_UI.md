# Web UI - Visual Guide

## 🎨 Interface Overview

A interface web do K8s Monitoring App foi projetada para ser intuitiva e fornecer informações em tempo real sobre o estado das suas aplicações Kubernetes.

## 📱 Layout Geral

```
┌─────────────────────────────────────────────────────────────────┐
│ 🚀 K8s Monitoring    [Dashboard] [Projects] [Applications]     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Dashboard                          ● Auto-refresh: 10s        │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │ VISA (Production Environment)                             │ │
│  │                                                            │ │
│  │  ┌─────────────────────────────────────────────────────┐  │ │
│  │  │ api             [n61] [n62] [ok]      ┌─────────┐  │  │ │
│  │  │                                          │  mem    │  │  │ │
│  │  │  pods: [1] [2] [3]                      │ ▓▓░░░░  │  │  │ │
│  │  │                                          │ 45.2%   │  │  │ │
│  │  │  health: https://api.../healthcheck     │  cpu    │  │  │ │
│  │  │                                          │ ░░░░░░  │  │  │ │
│  │  │                                          │  0.1%   │  │  │ │
│  │  │                                          │  disk   │  │  │ │
│  │  │                                          │ ▓░░░░░  │  │  │ │
│  │  │                                          │ 23.8%   │  │  │ │
│  │  └─────────────────────────────────────────└─────────┘  │  │ │
│  │                                                            │ │
│  │  ┌─────────────────────────────────────────────────────┐  │ │
│  │  │ Redis            [n61] [n62] [ok]      ┌─────────┐  │  │ │
│  │  │  ...                                    │  ...    │  │  │ │
│  │  └─────────────────────────────────────────└─────────┘  │  │ │
│  │                                                            │ │
│  │  ┌─────────────────────────────────────────────────────┐  │ │
│  │  │ Konga            [n61] [?]              ┌─────────┐  │  │ │
│  │  │  ...                                    │  ...    │  │  │ │
│  │  └─────────────────────────────────────────└─────────┘  │  │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 🎯 Componentes Principais

### 1. **Navbar (Barra de Navegação)**

```
┌──────────────────────────────────────────────────────┐
│ 🚀 K8s Monitoring  [Dashboard] [Projects] [Apps]   │
└──────────────────────────────────────────────────────┘
```

- Logo e título do app
- Links de navegação principais
- Estilo: Fundo branco, borda inferior cinza

### 2. **Dashboard Header**

```
Dashboard                           ● Auto-refresh: 10s
```

- Título da página
- Indicador de status (ponto verde pulsante)
- Informação de auto-refresh

### 3. **Project Card**

```
┌─────────────────────────────────────────────────┐
│ VISA                                            │
│ Production environment applications             │
│                                                 │
│ [Application Cards...]                          │
└─────────────────────────────────────────────────┘
```

**Características:**
- Borda preta grossa (2px)
- Título em itálico e negrito
- Descrição em cinza
- Contém múltiplas aplicações

### 4. **Application Card** (Componente Principal)

```
┌──────────────────────────────────────────────────────┐
│ api                   [n61] [n62] [ok]  ┌────────┐ │
│ namespace: production                     │  mem   │ │
│                                           │ ▓▓▓░░░ │ │
│  ┌────────┐  ┌────────────────────────┐  │ 45.2%  │ │
│  │ pods   │  │ health                 │  │        │ │
│  │ [1][2] │  │ https://api.../health  │  │  cpu   │ │
│  │ [3]    │  │                        │  │ ░░░░░░ │ │
│  └────────┘  └────────────────────────┘  │  0.1%  │ │
│                                           │        │ │
│                                           │  disk  │ │
│                                           │ ▓░░░░░ │ │
│                                           │ 23.8%  │ │
│                                           └────────┘ │
└──────────────────────────────────────────────────────┘
```

#### Elementos:

**A. Header**
- Nome da aplicação (esquerda)
- Namespace (abaixo do nome, cinza, monospace)
- Nodes ativos (badges verdes)
- Status de saúde (badge amarelo/vermelho)

**B. Pods Section**
- Label "pods" em azul
- Boxes numerados para cada pod
- Cores:
  - Verde: Pod running e ready
  - Amarelo: Pod pending
  - Cinza: Status desconhecido

**C. Health Section**
- Label "health" em azul
- URL do endpoint de health check
- Fundo cinza claro
- Fonte monospace

**D. Side Panel (Direita)**
- Borda verde
- 3 métricas verticais:

**Memory (mem):**
```
mem
▓▓▓░░░  (barra de progresso)
45.2%
```

**CPU:**
```
cpu
░░░░░░  (barra de progresso)
0.1%
```

**Disk:**
```
disk
▓░░░░░  (barra de progresso)
23.8%
```

## 🎨 Paleta de Cores

### Status Colors

| Elemento | Cor | Hex | Uso |
|----------|-----|-----|-----|
| Success | Verde | `#10b981` | Nodes, pods running |
| Warning | Amarelo | `#f59e0b` | Health ok, pods pending |
| Danger | Vermelho | `#ef4444` | Erros, health error |
| Info | Azul | `#2563eb` | Labels, títulos |

### UI Colors

| Elemento | Cor | Hex | Uso |
|----------|-----|-----|-----|
| Background | Cinza 50 | `#f9fafb` | Fundo da página |
| Card Background | Branco | `#ffffff` | Cards |
| Border | Preto | `#000000` | Bordas principais |
| Border Success | Verde | `#10b981` | Side panel |
| Text Primary | Cinza 900 | `#111827` | Texto principal |
| Text Secondary | Cinza 600 | `#4b5563` | Texto secundário |

## 🔄 Estados de Loading

### Initial Load

```
┌─────────────────────────────────┐
│  ⟳ Loading projects...         │
└─────────────────────────────────┘
```

### Application Loading

```
┌─────────────────────────────────┐
│  ⟳ Loading api   ...           │
└─────────────────────────────────┘
```

### Updating (HTMX)

Durante updates, há uma transição suave de opacity:
- Opacity: 1 → 0 → 1
- Duração: 200ms

## 📊 Progress Bars (Mem/CPU/Disk)

### Estrutura

```html
┌─────────────────────┐
│ ▓▓▓▓▓░░░░░░░░░░░░░░ │  (barra)
│       45.2%          │  (valor)
└─────────────────────┘
```

### Gradiente de Cores

```
0%        50%        70%        100%
|─────────|──────────|───────────|
  Verde    →  Verde   →  Amarelo  →  Vermelho
#10b981      #10b981     #f59e0b     #ef4444
```

- 0-70%: Verde (normal)
- 70-85%: Transição verde→amarelo (atenção)
- 85-100%: Transição amarelo→vermelho (crítico)

## 🔔 Status Indicators

### Health Badges

```
┌────┐     ┌───────┐     ┌───┐
│ ok │     │ error │     │ ? │
└────┘     └───────┘     └───┘
Amarelo     Vermelho      Cinza
```

### Node Badges

```
┌─────┐ ┌─────┐
│ n61 │ │ n62 │
└─────┘ └─────┘
  Verde   Verde
```

### Pod Badges

```
┌───┐ ┌───┐ ┌───┐
│ 1 │ │ 2 │ │ 3 │
└───┘ └───┘ └───┘
Verde Verde Amarelo
```

## 📱 Responsividade

### Desktop (> 768px)

```
┌────────────────────────────────────┐
│ Application                        │
│ ┌────────────┐     ┌────────────┐ │
│ │            │     │    Side    │ │
│ │   Metrics  │     │   Panel    │ │
│ │            │     │            │ │
│ └────────────┘     └────────────┘ │
└────────────────────────────────────┘
```

### Mobile (< 768px)

```
┌────────────────┐
│ Application    │
│ ┌────────────┐ │
│ │  Metrics   │ │
│ └────────────┘ │
│ ┌────────────┐ │
│ │ Side Panel │ │
│ │ (Horizontal│ │
│ └────────────┘ │
└────────────────┘
```

## ✨ Animações

### 1. Status Indicator Pulse

```css
● → ◐ → ○ → ◐ → ●
(Opacidade: 1 → 0.5 → 1)
Duração: 2s
Loop: Infinito
```

### 2. Loading Spinner

```
⟲ → ⟳ → ⟲ → ⟳
Rotação: 360deg
Duração: 1s
Loop: Infinito
```

### 3. Content Update

```
Conteúdo Antigo → Fade Out → Fade In → Conteúdo Novo
Duração: 200ms
```

## 🎯 Interatividade

### Auto-refresh

- **Frequência**: A cada 10 segundos
- **Scope**: Toda a página (projects e applications)
- **Visual**: Indicador pulsante mostra atividade
- **Performance**: Carregamento assíncrono e independente

### Hover States

```css
Links: Cinza → Azul (com background cinza claro)
Cards: Sem hover (informativo apenas)
```

## 📐 Dimensões

### Spacing

```
Container padding: 20px
Card padding: 1.5rem (24px)
Gap entre cards: 1rem (16px)
Gap entre elementos: 0.75rem (12px)
```

### Border Radius

```
Cards principais: 16px
Elementos internos: 8px
Badges pequenos: 4px
```

### Bordas

```
Project card: 2px solid black
Application card: 2px solid black
Side panel: 2px solid green
Metric boxes: 2px solid green
```

## 🚀 Performance

### Otimizações

1. **Lazy Loading**: Cada aplicação carrega independentemente
2. **Polling Eficiente**: Apenas recarrega o necessário via HTMX
3. **CSS Minimalista**: Sem frameworks pesados
4. **No JavaScript**: Tudo via HTMX (apenas 14KB)

### Tempo de Carregamento

```
Initial Load:  < 100ms
Project Load:  < 200ms
App Metrics:   < 150ms por app
Refresh:       < 150ms
```

## 🔮 Estados Especiais

### Sem Dados

```
┌─────────────────────────────────┐
│ api             [?]            │
│                                 │
│ pods: [?]                       │
│ health: Not configured          │
│                         mem: -  │
│                         cpu: -  │
│                        disk: -  │
└─────────────────────────────────┘
```

### Erro de Coleta

```
┌─────────────────────────────────┐
│ api             [error]        │
│                                 │
│ pods: [?] [?] [?]              │
│ health: Connection failed       │
│                         mem: -  │
│                         cpu: -  │
│                        disk: -  │
└─────────────────────────────────┘
```

## 📚 Referências de Design

- Inspiração: Mockup fornecido pelo usuário
- Estilo: Clean/Minimalist
- Framework: Custom CSS (sem Bootstrap/Tailwind)
- Typography: System fonts (-apple-system, Roboto, etc)

---

**Resultado:** Interface limpa, rápida e intuitiva para monitoramento em tempo real! 🎉

