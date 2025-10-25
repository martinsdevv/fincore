# **Fincore API**

Uma API completa de **gestão financeira corporativa** desenvolvida em **Go**, com foco em boas práticas, escalabilidade e padrões utilizados em empresas reais.  
O projeto simula o backend de um sistema financeiro corporativo, com módulos de **autenticação**, **gestão de contas e transações**, **relatórios financeiros** e **integração com APIs externas** (como taxas de câmbio).

---

## 🎯 **Objetivo**
Demonstrar **experiência profissional sólida em Go**, incluindo:
- Arquitetura limpa (Clean Architecture)
- Boas práticas REST
- Padrões de projeto (Repository, Service, DTO)
- Uso de Go routines e channels
- Autenticação JWT
- Testes unitários e de integração
- Deploy com Docker e CI/CD
- Integração com banco PostgreSQL

---

## 🧱 **Arquitetura do Projeto**

```
Fincore/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── accounts/
│   ├── transactions/
│   ├── reports/
│   ├── common/ (middlewares, utils)
│   └── config/
├── pkg/
│   ├── database/
│   ├── logger/
│   └── external/
├── tests/
│   ├── integration/
│   └── unit/
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

---

## ⚙️ **Funcionalidades**

### 🔐 1. Autenticação
- Registro e login de usuários.
- Criptografia de senha com bcrypt.
- Autenticação via **JWT**.
- Middleware de autorização.

### 💰 2. Contas e Transações
- CRUD completo de contas financeiras.
- Registro de transações (entrada/saída).
- Cálculo de saldo em tempo real.
- Transferências entre contas.
- Histórico de transações com paginação.

### 📊 3. Relatórios
- Relatórios de despesas por categoria e período.
- Análise de fluxo de caixa.
- Exportação para CSV e PDF.

### 🌍 4. Integração Externa
- Consulta de taxas de câmbio via **API externa** (ex: https://exchangerate.host).
- Conversão automática de valores em diferentes moedas.

### ⚡ 5. Performance
- Processamento assíncrono de relatórios com **Go routines e channels**.
- Cache em memória usando **sync.Map** para evitar reprocessamentos.
- Filas simples para tarefas de background.

---

## 🧪 **Testes**
- Testes unitários para services e handlers com `testing` e `testify`.
- Testes de integração com banco PostgreSQL em container.
- Mock de dependências com interfaces.

---

## 🐳 **Infraestrutura**
- **Docker Compose** para subir ambiente completo.
- Serviços:
  - `api`: aplicação Go
  - `db`: PostgreSQL
  - `cache`: Redis
- Volume persistente para banco.
- Variáveis de ambiente via `.env`.

---

## 🚀 **Deploy e CI/CD**
- Pipeline no **GitHub Actions**:
  - Lint e testes automáticos.
  - Build da imagem Docker.
  - Deploy automático em ambiente de staging.
- Logs estruturados com `zerolog`.
- Monitoramento com Prometheus e Grafana.

---

## 🗂️ **Passos de Desenvolvimento**

1. Configurar ambiente e dependências do Go.
2. Criar estrutura de pastas e `main.go`.
3. Implementar módulo `config` para variáveis de ambiente.
4. Implementar conexão com banco PostgreSQL (`pkg/database`).
5. Criar módulo `auth` com registro e login (JWT + bcrypt).
6. Criar módulo `accounts` (CRUD completo).
7. Criar módulo `transactions` e lógica de saldo.
8. Adicionar relatórios (`reports`), cache e goroutines.
9. Integrar API externa de câmbio.
10. Implementar testes unitários e de integração.
11. Configurar Docker e Compose.
12. Adicionar CI/CD com GitHub Actions.
13. Deploy final e documentação no README.

---

## 💼 **Resumo Profissional (para incluir no portfólio)**

> Desenvolvimento de uma API corporativa em Go para gestão financeira, aplicando Clean Architecture, autenticação JWT, PostgreSQL e Docker.  
> Implementação de pipelines de CI/CD com GitHub Actions e integração com serviços externos para conversão de moedas.  
> O projeto demonstra experiência em design de sistemas escaláveis, uso de goroutines, caching e testes automatizados.
