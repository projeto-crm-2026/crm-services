# CRM Services

Sistema de CRM desenvolvido em Go seguindo **Arquitetura Hexagonal**, **Options Pattern** e **Strategy Pattern**.

## 📁 Estrutura do Projeto

```
crm-services/
├── cmd/
│   └── api/
│       └── main.go                 # Entry Point
├── internal/
│   ├── config/
│   │   ├── config.go               # Carregamento de configurações
│   │   └── fields.go               # Structs de configuração
│   ├── domain/
│   │   ├── constant/
│   │   │   ├── constant.go         # Constantes globais
│   │   │   └── errors.go           # Mensagens de erro
│   │   └── entity/
│   │       ├── apikey.go           # Entidade API Key
│   │       ├── chat.go             # Entidade Chat
│   │       ├── chatparticipant.go  # Entidade Participante do Chat
│   │       ├── message.go          # Entidade Mensagem
│   │       ├── user.go             # Entidade Usuário
│   │       └── webhook.go          # Entidade Webhook
│   ├── repo/
│   │   ├── apikeyrepo.go           # Repositório de API Keys
│   │   ├── chatrepo.go             # Repositório de Chats
│   │   ├── messagerepo.go          # Repositório de Mensagens
│   │   ├── repo.go                 # Conexão com banco de dados
│   │   ├── userrepo.go             # Repositório de Usuários
│   │   └── webhookrepo.go          # Repositório de Webhooks
│   ├── server/
│   │   ├── server.go               # Configuração do servidor HTTP
│   │   ├── adapters/
│   │   │   └── widgetadapter.go    # Adapter para validação de widget
│   │   ├── handler/
│   │   │   ├── apikeyhandler.go    # Handler de API Keys
│   │   │   ├── chathandler.go      # Handler de Chat
│   │   │   ├── healthhandler.go    # Handler de Health Check
│   │   │   ├── userhandler.go      # Handler de Usuário
│   │   │   ├── webhookhandler.go   # Handler de Webhooks
│   │   │   └── widgethandler.go    # Handler de Widget
│   │   ├── middleware/
│   │   │   ├── contentjson.go      # Middleware Content-Type JSON
│   │   │   ├── corsmiddleware.go   # Middleware CORS
│   │   │   ├── jwtmiddleware.go    # Middleware de autenticação JWT
│   │   │   └── widgetorigin.go     # Middleware de validação de origem do widget
│   │   ├── model/
│   │   │   ├── apikey.go           # DTOs de API Key
│   │   │   ├── chat.go             # DTOs de Chat
│   │   │   ├── user.go             # DTOs de Usuário
│   │   │   ├── webhook.go          # DTOs de Webhook
│   │   │   └── widget.go           # DTOs de Widget
│   │   ├── route/
│   │   │   └── routes.go           # Definição de rotas
│   │   └── websocket/
│   │       ├── client.go           # Cliente WebSocket
│   │       ├── handler.go          # Handler WebSocket
│   │       ├── hub.go              # Hub de gerenciamento de conexões
│   │       └── websocket-test.html # Página de teste
│   └── service/
│       ├── chatservice/
│       │   └── chatservice.go      # Serviço de Chat
│       ├── userservice/
│       │   └── userservice.go      # Serviço de Usuário
│       ├── webhookservice/
│       │   ├── dispatcher.go       # Dispatcher de eventos
│       │   ├── events.go           # Definição de eventos
│       │   └── webhookservice.go   # Serviço de Webhooks
│       └── widgetservice/
│           ├── exceptions.go       # Exceções do Widget
│           └── widgetservice.go    # Serviço de Widget
├── pkg/
│   ├── jwt/
│   │   └── jwt.go                  # Utilitários JWT para usuários
│   ├── passwordhashing/
│   │   └── passwordhashing.go      # Utilitários de hash de senha
│   └── visitorjwt/
│       └── visitorjwt.go           # Utilitários JWT para visitantes
├── .editorconfig
├── .env
├── .ex.env                         # Exemplo de variáveis de ambiente (copiar e colar no .env)
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── Makefile
├── README.md
└── sonar-project.properties
```

## 🏗️ Arquitetura

### Arquitetura Hexagonal (Ports and Adapters)

O projeto segue a arquitetura hexagonal, separando claramente:

- **Domain (Núcleo)**: Entidades (`internal/domain/`)
- **Ports (Interfaces)**: Interfaces que definem contratos e regras de negócio nos services (`internal/repo/`, `internal/service/`)
- **Adapters (Implementações)**: Implementações concretas (`internal/server/`, `internal/repo/`)

```
┌─────────────────────────────────────────────────────────────┐
│                      Adapters (HTTP)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Handlers   │  │ Middlewares │  │     WebSocket       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                    Ports (Interfaces)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ UserService │  │ ChatService │  │   WidgetService     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                      Domain (Core)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Entities  │  │  Constants  │  │   Business Rules    │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                   Adapters (Database)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  UserRepo   │  │  ChatRepo   │  │   APIKeyRepo        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Options Pattern

Utilizado na criação do servidor para configuração flexível:

```go
// Exemplo de uso do Options Pattern
srv := server.NewServer(
    server.WithLogger(logger),
    server.WithConfig(cfg),
    server.WithDB(dbConn),
    server.WithHealthHandler(healthHandler),
    server.WithUserHandler(userHandler),
    server.WithChatHandler(chatHandler),
    server.WithWidgetHandler(widgetHandler),
    server.WithContentJSONMiddleware(contentJsonMiddleware),
    server.WithJWTMiddleware(jwtMiddleware),
    server.WithCorsMiddleware(corsMiddleware),
    server.WithWidgetAuthMiddleware(widgetAuthMiddleware),
)
```

### Strategy Pattern

Utilizado nos middlewares de autenticação, permitindo diferentes estratégias:

- **JWTMiddleware**: Autenticação para usuários CRM
- **WidgetAuthMiddleware**: Autenticação para widgets externos

## 🚀 Como Executar

### Pré-requisitos

- Docker e Docker Compose
- Go 1.25+ (para desenvolvimento local)

### Usando Docker

```bash
# Iniciar todos os serviços
make up

# Ver logs
make logs

# Parar serviços
make down

# Limpar volumes
make clean
```

### Desenvolvimento Local

```bash
# Inicia os serviços em modo log
make dev

```

## 📡 API Endpoints

### Autenticação

| Método | Endpoint    | Descrição         |
|--------|-------------|-------------------|
| POST   | `/register` | Registrar usuário |
| POST   | `/login`    | Login             |
| POST   | `/logout`   | Logout            |

### API Keys (Autenticado)

| Método | Endpoint          | Descrição         |
|--------|-------------------|-------------------|
| POST   | `/api-keys`       | Criar API Key     |
| GET    | `/api-keys`       | Listar API Keys   |
| DELETE | `/api-keys/{id}`  | Deletar API Key   |

### Widget (Com X-Widget-Key)

| Método | Endpoint                          | Descrição              |
|--------|-----------------------------------|------------------------|
| POST   | `/widget/init`                    | Inicializar sessão     |
| POST   | `/widget/chat`                    | Criar chat             |
| GET    | `/widget/chat/{chatID}/messages`  | Obter mensagens        |

### Chat (Autenticado)

| Método | Endpoint                   | Descrição          |
|--------|----------------------------|--------------------|
| GET    | `/chats`                   | Listar chats       |
| GET    | `/chats/{chatID}`          | Obter chat         |
| GET    | `/chats/{chatID}/messages` | Obter mensagens    |

### Webhooks - Outgoing (Autenticado)

| Método | Endpoint                      | Descrição                    |
|--------|-------------------------------|------------------------------|
| GET    | `/webhooks/events`            | Listar eventos disponíveis   |
| POST   | `/webhooks`                   | Criar webhook                |
| GET    | `/webhooks`                   | Listar webhooks              |
| GET    | `/webhooks/{webhookID}`       | Obter webhook                |
| PUT    | `/webhooks/{webhookID}`       | Atualizar webhook            |
| DELETE | `/webhooks/{webhookID}`       | Deletar webhook              |
| GET    | `/webhooks/{webhookID}/logs`  | Obter logs de envio          |

### Webhooks - Incoming Tokens (Autenticado)

| Método | Endpoint                      | Descrição                    |
|--------|-------------------------------|------------------------------|
| POST   | `/webhooks/tokens`            | Criar token de entrada       |
| GET    | `/webhooks/tokens`            | Listar tokens                |
| DELETE | `/webhooks/tokens/{tokenID}`  | Deletar token                |

### Webhooks - Incoming (Com X-Webhook-Token)

| Método | Endpoint            | Descrição                    |
|--------|---------------------|------------------------------|
| POST   | `/webhook/incoming` | Receber webhook externo      |

## 🔔 Webhooks

O sistema suporta webhooks bidirecionais:

### Outgoing Webhooks (CRM → Seu Servidor)

Configure webhooks para receber notificações quando eventos ocorrerem no CRM.

#### Eventos Disponíveis

| Evento                  | Descrição                           |
|-------------------------|-------------------------------------|
| `message.received`      | Mensagem recebida de visitante      |
| `message.sent`          | Mensagem enviada por agente         |
| `chat.created`          | Novo chat criado                    |
| `chat.closed`           | Chat fechado                        |
| `visitor.connected`     | Visitante conectou ao WebSocket     |
| `visitor.disconnected`  | Visitante desconectou do WebSocket  |

#### Criar Webhook

```bash
curl -X POST http://localhost:8080/webhooks \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "Minha Integração",
    "url": "https://meu-servidor.com/webhook",
    "events": ["message.received", "chat.created"]
  }'
```

Resposta:
```json
{
  "id": 1,
  "name": "Minha Integração",
  "url": "https://meu-servidor.com/webhook",
  "secret": "whsec_abc123...",
  "events": ["message.received", "chat.created"],
  "is_active": true,
  "fail_count": 0,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Payload do Webhook

Quando um evento ocorre, o CRM envia um POST para sua URL:

```json
{
  "id": "evt_uuid-do-evento",
  "type": "message.received",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "chat_id": 1,
    "message_id": 42,
    "content": "Olá, preciso de ajuda!",
    "visitor_id": "visitor-uuid",
    "type": "text",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

#### Headers do Webhook

| Header               | Descrição                              |
|----------------------|----------------------------------------|
| `Content-Type`       | `application/json`                     |
| `X-Webhook-Signature`| Assinatura HMAC-SHA256 do payload      |
| `X-Webhook-Event`    | Tipo do evento                         |
| `X-Webhook-ID`       | ID único do evento                     |

#### Verificando Assinatura

```javascript
const crypto = require('crypto');

function verifySignature(payload, signature, secret) {
  const expectedSignature = 'sha256=' + crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}
```

### Incoming Webhooks (Seu Servidor → CRM)

Envie comandos para o CRM através de webhooks de entrada.

#### Criar Token de Entrada

```bash
curl -X POST http://localhost:8080/webhooks/tokens \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name": "Bot de Automação"}'
```

Resposta:
```json
{
  "id": 1,
  "token": "whit_abc123...",
  "name": "Bot de Automação",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Ações Disponíveis

| Ação            | Descrição                    |
|-----------------|------------------------------|
| `send_message`  | Enviar mensagem no chat      |
| `close_chat`    | Fechar chat                  |

#### Enviar Mensagem via Webhook

```bash
curl -X POST http://localhost:8080/webhook/incoming \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Token: whit_abc123..." \
  -d '{
    "action": "send_message",
    "chat_id": 1,
    "content": "Olá! Esta é uma mensagem automática."
  }'
```

Resposta:
```json
{
  "status": "ok"
}
```

A mensagem será enviada automaticamente para todos os clientes conectados via WebSocket.

## 💬 Exemplo de Uso do WebSocket

### 1. Conectando como Agente CRM (JavaScript)

```javascript
// Após fazer login (cookie auth_token será enviado automaticamente)
const chatId = 1;
const ws = new WebSocket(`ws://localhost:8080/ws/chat/${chatId}?visitor_id=agent-123`);

ws.onopen = () => {
    console.log('Conectado como agente CRM');
    
    // Enviar mensagem
    ws.send(JSON.stringify({
        type: 'message',
        content: 'Olá! Como posso ajudar?',
        visitor_id: 'agent-123'
    }));
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Mensagem recebida:', data.content);
};

ws.onclose = () => {
    console.log('Desconectado');
};
```

### 2. Conectando como Visitante (JavaScript)

```javascript
const chatId = 1;
const visitorId = 'visitor-uuid-aqui';
const ws = new WebSocket(`ws://localhost:8080/ws/widget/${chatId}?visitor_id=${visitorId}`);

ws.onopen = () => {
    console.log('Conectado como visitante');
    
    // Enviar mensagem
    ws.send(JSON.stringify({
        type: 'message',
        content: 'Olá! Preciso de ajuda.',
        visitor_id: visitorId
    }));
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Mensagem recebida:', data.content);
};
```

### 3. Implementação Completa de Chat (HTML/JavaScript)

```html
<!DOCTYPE html>
<html>
<head>
    <title>Chat Widget</title>
    <style>
        .chat-container { max-width: 400px; margin: 20px auto; }
        .messages { height: 300px; overflow-y: auto; border: 1px solid #ddd; padding: 10px; }
        .message { margin: 5px 0; padding: 8px; border-radius: 8px; }
        .message.sent { background: #007bff; color: white; text-align: right; }
        .message.received { background: #f1f1f1; }
        .input-area { display: flex; gap: 10px; margin-top: 10px; }
        .input-area input { flex: 1; padding: 10px; }
        .input-area button { padding: 10px 20px; }
    </style>
</head>
<body>
    <div class="chat-container">
        <h3>Chat de Suporte</h3>
        <div class="messages" id="messages"></div>
        <div class="input-area">
            <input type="text" id="messageInput" placeholder="Digite sua mensagem...">
            <button onclick="sendMessage()">Enviar</button>
        </div>
    </div>

    <script>
        const API_BASE = 'http://localhost:8080';
        const WS_BASE = 'ws://localhost:8080';
        const WIDGET_KEY = 'pk_sua-chave-publica';
        
        let ws = null;
        let visitorId = null;
        let chatId = null;

        // Inicializar widget
        async function initWidget() {
            const response = await fetch(`${API_BASE}/widget/init`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Widget-Key': WIDGET_KEY
                },
                body: JSON.stringify({ visitor_id: '' })
            });
            
            const data = await response.json();
            visitorId = data.visitor_id;
            console.log('Widget inicializado, visitor:', visitorId);
            
            await createChat();
        }

        // Criar chat
        async function createChat() {
            const response = await fetch(`${API_BASE}/widget/chat`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Widget-Key': WIDGET_KEY
                },
                body: JSON.stringify({ visitor_id: visitorId })
            });
            
            const data = await response.json();
            chatId = data.id;
            console.log('Chat criado:', chatId);
            
            connectWebSocket();
        }

        // Conectar WebSocket
        function connectWebSocket() {
            ws = new WebSocket(`${WS_BASE}/ws/widget/${chatId}?visitor_id=${visitorId}`);
            
            ws.onopen = () => {
                console.log('WebSocket conectado');
                addMessage('Sistema', 'Conectado ao chat!', 'received');
            };
            
            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                const isSent = data.visitor_id === visitorId;
                addMessage(
                    isSent ? 'Você' : 'Agente',
                    data.content,
                    isSent ? 'sent' : 'received'
                );
            };
            
            ws.onclose = () => {
                console.log('WebSocket desconectado');
                addMessage('Sistema', 'Desconectado do chat.', 'received');
            };
        }

        // Enviar mensagem
        function sendMessage() {
            const input = document.getElementById('messageInput');
            const content = input.value.trim();
            
            if (!content || !ws || ws.readyState !== WebSocket.OPEN) return;
            
            ws.send(JSON.stringify({
                type: 'message',
                content: content,
                visitor_id: visitorId
            }));
            
            input.value = '';
        }

        // Adicionar mensagem na tela
        function addMessage(sender, content, type) {
            const container = document.getElementById('messages');
            const div = document.createElement('div');
            div.className = `message ${type}`;
            div.innerHTML = `<strong>${sender}:</strong> ${content}`;
            container.appendChild(div);
            container.scrollTop = container.scrollHeight;
        }

        // Enter para enviar
        document.getElementById('messageInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') sendMessage();
        });

        // Inicializar ao carregar página
        initWidget();
    </script>
</body>
</html>
```

## 🔧 Exemplo de Uso do Widget

### Passo 1: Criar API Key

Primeiro, faça login e crie uma API Key para seu domínio:

```bash
# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email": "seu@email.com", "password": "senha123"}'

# Criar API Key
curl -X POST http://localhost:8080/api-keys \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name": "Meu Site", "domain": "meusite.com"}'
```

Resposta:
```json
{
  "id": 1,
  "public_key": "pk_abc123...",
  "secret_key": "sk_xyz789...",
  "name": "Meu Site",
  "domain": "meusite.com",
  "is_active": true
}
```

### Passo 2: Integrar Widget no Site (futuramente criar sdk)

```html
<!-- Adicione no seu site -->
<script>
(function() {
    const WIDGET_KEY = 'pk_abc123...'; // Sua public_key
    const API_BASE = 'https://seu-servidor.com';
    
    // Código do widget aqui (exemplo anterior)
})();
</script>
```

## 🧪 Testes

### Test suit

Pré-requisitos: Python3 e Makefile

```bash
# Iniciar servidor de teste
make test-ui

# Acesse http://localhost:3000/websocket-test.html
```

O test suite inclui:
- Autenticação (Register/Login)
- Gerenciamento de API Keys
- Inicialização de Widget
- Criação de Chat
- WebSocket (CRM Agent e Widget Client)
- **Outgoing Webhooks** (criar, listar, logs)
- **Incoming Webhooks** (tokens e envio de mensagens)

### Testando Webhooks

1. Acesse [webhook.site](https://webhook.site) e copie a URL
2. Crie um webhook com a URL copiada
3. Envie mensagens via WebSocket como visitante
4. Verifique os eventos recebidos no webhook.site

## 📝 Licença

Este projeto está sob a licença MIT.