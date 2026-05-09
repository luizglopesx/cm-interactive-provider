# Evolution Go local interactive test

## Status

- Branch: `interactive-go-test`
- Docker image: `evolution-go-interactive:local`
- Container: `evolution-go-interactive-test`
- Local URL: `http://localhost:8093`
- Global API key: `BQYHJGJHJ123`
- Auth DB: `evogo_auth_test`
- Users DB: `evogo_users_test`

The current upstream code already includes:

- `POST /send/button`
- `POST /send/list`
- `POST /send/carousel`
- a local `whatsmeow-lib` replacement module
- interactive/list/carousel message builders

API endpoints are gated by the official license middleware until `/license/status`
returns `active`.

## Dev setup (primeira vez)

```bash
git clone https://github.com/luizglopesx/cm-interactive-provider
cd cm-interactive-provider/evolution-go-interactive
git checkout interactive-go-test

# Build
docker build -t evolution-go-interactive:local .

# Run
docker rm -f evolution-go-interactive-test 2>/dev/null
docker run -d \
  --name evolution-go-interactive-test \
  --add-host=host.docker.internal:host-gateway \
  -p 8093:8093 \
  --env-file .env \
  evolution-go-interactive:local

# Ativar licenca
curl -i -sS "http://localhost:8093/license/register?redirect_uri=http://localhost:8093/manager/license/callback"

# Criar instancia
curl -X POST "http://localhost:8093/instance/create" \
  -H "Content-Type: application/json" \
  -H "apikey: BQYHJGJHJ123" \
  -d '{"name":"interactive-go-test","token":"interactive-go-token-123"}'

# Conectar e escanear QR
curl -X POST "http://localhost:8093/instance/connect" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123"

curl "http://localhost:8093/instance/qr" \
  -H "apikey: interactive-go-token-123"

# Verificar conexao
curl "http://localhost:8093/instance/status" \
  -H "apikey: interactive-go-token-123"
```

## Test Results — 2026-05-09 (sessão manhã)

### Environment
- License activated: ✅ `active`
- Instance connected: ✅ `LoggedIn: true` — `Senhor Colchão` (`551733233694:63@s.whatsapp.net`)
- Recipient: `5517981189332`

### Results

| Endpoint | API response | Mobile | Desktop WA | Notes |
|---|---|---|---|---|
| `POST /send/text` | ✅ | ✅ | ✅ | Texto simples |
| `POST /send/button` (reply) | ✅ | ✅ | ✅ | Após injeção do `<biz>` node |
| `POST /send/button` (copy/PIX) | ✅ | ✅ | ✅ | |
| `POST /send/button` (thumbnailUrl) | ✅ | ✅ | ✅ | Header com imagem — novo |
| `POST /send/list` | ✅ | ✅ | ✅ (testar) | Após correção do biz node type |
| `POST /send/carousel` | ✅ | ✅ | ❌ | Desktop WA não renderiza carrossel |

## Fixes aplicados (2026-05-08 e 2026-05-09)

### 1. Biz node para InteractiveMessage (botões/carrossel)

O WhatsApp filtra `NativeFlowMessage` quando o stanza não tem o nó `<biz>` que sinaliza
contexto Business legítimo.

**Fix:** injetar `AdditionalNodes` no `whatsmeow.SendRequestExtra` em
`pkg/sendMessage/service/send_service.go` (linha ~1853: `SendButton`, linha ~2372: `SendMessage`).

- `InteractiveMessage` → `<biz><interactive type="native_flow" v="1"><native_flow v="9" name="mixed"/></interactive></biz>`
- `ListMessage` → whatsmeow já injeta automaticamente (ver item 3)

Equivalente Node.js: `buildInteractiveBizNode()` em `interactiveMessage.helper.ts`.

### 2. Estrutura do carrossel alinhada com 2.3.7

| Campo | Antes (Go) | Depois (Go = 2.3.7) |
|---|---|---|
| `msg.MessageContextInfo` | `{ DeviceListMetadata: {} }` | omitido |
| `interactiveMessage.Footer` (top-level) | preenchido | omitido |
| `interactiveMessage.ContextInfo` (top-level) | sempre `{}` | só se há `quoted` |
| `cards[].Header` | sempre criado | só se há `imageUrl`/`videoUrl` |

### 3. ListMessage biz node type (`product_list`)

O endpoint `/send/list` sempre esteve implementado. O erro `405` vinha do servidor
WhatsApp porque o atributo `type` do nó `<biz>` estava errado.

**Causa:** `whatsmeow-lib/send.go:getButtonAttributes()` derivava o tipo do enum protobuf
(`SINGLE_SELECT` → `"single_select"`), mas o WhatsApp espera `"product_list"`.

**Fix:** `whatsmeow-lib/send.go:1014` — alterado para `"product_list"` hard-coded.

O Evolution 2.3.7 já fazia isso certo no `buildListBizNode()`:
```typescript
{ tag: 'list', attrs: { type: 'product_list', v: '2' } }
```

**Commit:** `c074d14`

### 4. thumbnailUrl no botão

Adicionado campo `thumbnailUrl` no `ButtonStruct`. Quando informado, o servidor:
1. Baixa a imagem da URL
2. Faz upload para os servidores do WhatsApp (`client.Upload`)
3. Gera thumbnail JPEG de 72px para compatibilidade com iOS
4. Injeta no header do InteractiveMessage com `hasMediaAttachment: true`

**Arquivo:** `pkg/sendMessage/service/send_service.go`
- `ButtonStruct.ThumbnailUrl` (struct)
- `SendButton` (lógica de download/upload/header)

**Commit:** `41258ab`

## Validações de payload (paridade com Evolution 2.3.7)

`POST /send/button`:
- ≥ 1 botão obrigatório
- `reply` máx 3, e não pode misturar com outros tipos
- CTA (`url`/`call`/`copy`) máx 2
- `pix` máx 1 e não pode misturar com outros tipos
- `thumbnailUrl` (opcional): URL de imagem pública exibida como header

`POST /send/carousel`:
- ≥ 1 card e ≤ 10 cards
- Cada card: ≥ 1 botão e ≤ 3 botões
- `pix` não é suportado dentro de cards de carrossel

## Paridade Evolution Go ↔ 2.3.7

| 2.3.7 | Evolution Go | Status |
|---|---|---|
| `sendText` | `/send/text` | ✅ |
| `sendMedia` | `/send/media` | ✅ |
| `sendSticker` | `/send/sticker` | ✅ |
| `sendLocation` | `/send/location` | ✅ |
| `sendContact` | `/send/contact` | ✅ |
| `sendPoll` | `/send/poll` | ✅ |
| `sendButtons` | `/send/button` | ✅ |
| `sendButtons` (thumbnailUrl) | `/send/button` (thumbnailUrl) | ✅ |
| `sendList` | `/send/list` | ✅ |
| `sendCarousel` | `/send/carousel` | ✅ |
| `sendStatus` | `/send/status/text` + `/send/status/media` | ✅ |
| `sendReaction` | `/message/react` | ✅ |
| `sendTemplate` | — | ❌ |
| `sendPtv` | — | ❌ |
| `sendWhatsAppAudio` | — | ❌ |

## Sender → recipient delivery matrix

| Sender | Cliente do recipient | Botão | Botão c/ img | Lista | Carrossel |
|---|---|---|---|---|---|
| Senhor Colchão (Business App) | Mobile (Android/iOS) | ✅ | ✅ | ✅ | ✅ |
| Senhor Colchão (Business App) | WhatsApp Desktop/Mac | ✅ | ✅ | ✅ (testar) | ❌ |

## Test commands

### Botão reply

```bash
curl -X POST "http://localhost:8093/send/button" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","title":"Escolha uma opcao","description":"Teste de botoes","footer":"Evolution Go","buttons":[{"type":"reply","displayText":"Comprar","id":"buy"},{"type":"reply","displayText":"Atendente","id":"agent"}]}'
```

### Botão com imagem (thumbnailUrl)

```bash
curl -X POST "http://localhost:8093/send/button" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","title":"Oferta Especial","description":"Confira nossos planos!","footer":"Evolution GO","thumbnailUrl":"https://picsum.photos/seed/btnheader/400/300","buttons":[{"type":"reply","displayText":"Quero saber mais","id":"cta_saber_mais"},{"type":"reply","displayText":"Falar com consultor","id":"cta_consultor"}]}'
```

### Botão PIX (copy)

```bash
curl -X POST "http://localhost:8093/send/button" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","title":"Pagamento via PIX","description":"Clique para copiar a chave PIX","footer":"Senhor Colchao","buttons":[{"type":"copy","displayText":"Copiar chave PIX","copyCode":"279949240007"}]}'
```

### Lista

```bash
curl -X POST "http://localhost:8093/send/list" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","title":"Menu de atendimento","description":"Escolha uma opcao","footerText":"Senhor Colchao","buttonText":"Ver opcoes","sections":[{"title":"Vendas","rows":[{"title":"Comprar colchao","description":"Ver modelos","rowId":"sales_mattress"},{"title":"Consultar pedido","description":"Status do pedido","rowId":"sales_order"}]},{"title":"Suporte","rows":[{"title":"Falar com atendente","description":"Atendimento humano","rowId":"support_agent"}]}]}'
```

### Carrossel

```bash
curl -X POST "http://localhost:8093/send/carousel" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","body":"Modelos em destaque","footer":"Senhor Colchao","cards":[{"header":{"imageUrl":"https://picsum.photos/600/400?random=11"},"body":{"text":"Colchao Premium - conforto alto."},"footer":"A partir de R$ 999","buttons":[{"type":"REPLY","displayText":"Quero esse","id":"premium"}]},{"header":{"imageUrl":"https://picsum.photos/600/400?random=12"},"body":{"text":"Colchao Luxo - reforcado casal."},"footer":"A partir de R$ 1299","buttons":[{"type":"REPLY","displayText":"Ver luxo","id":"luxo"}]}]}'
```
