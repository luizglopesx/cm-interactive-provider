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

## Test Results — 2026-05-08 (sessão tarde)

### Environment
- License activated: ✅ `active`
- Instance connected: ✅ `LoggedIn: true` — `Senhor Colchão` (`551733233694:63@s.whatsapp.net`)
- Recipient: `5517981189332`

### Results

| Endpoint | API response | Delivered (mobile)? | Notes |
|---|---|---|---|
| `POST /send/text` | ✅ success | ✅ YES | Texto simples |
| `POST /send/button` (reply) | ✅ success | ✅ YES | Após injeção do `<biz>` node |
| `POST /send/button` (copy/PIX) | ✅ success | ✅ YES | Após injeção do `<biz>` node |
| `POST /send/list` | ❌ `server returned error 405` | ❌ NO | WA server bloqueia ListMessage legacy (`not-allowed`) — feature descontinuada nessa conta |
| `POST /send/carousel` | ✅ success | ✅ YES | Após alinhamento estrutural com 2.3.7 |

> WhatsApp Desktop/Web NÃO renderiza `CarouselMessage` por design do cliente. Carrossel
> é mobile-only (Android/iOS); botões e lista chegam no Desktop normalmente. Mesmo
> comportamento do Evolution 2.3.7 (Baileys).

### Root cause + fixes aplicados

**1. Botões / Lista / Carrossel não chegavam (silenciados pelo servidor WhatsApp)**

O WhatsApp filtra `NativeFlowMessage` quando o stanza não tem o nó `<biz>` que sinaliza
contexto Business legítimo. ACK é retornado, mas a mensagem nunca chega no destinatário.

**Fix:** injetar `AdditionalNodes` no `whatsmeow.SendRequestExtra` em
`pkg/sendMessage/service/send_service.go`:

- `SendButton` (linha ~1844): injeta `<biz><interactive type="native_flow" v="1"><native_flow v="9" name="mixed"/></interactive></biz>`
- `SendMessage` (linha ~2358, switch por `messageType`):
  - `"InteractiveMessage"` → mesmo nó `biz/interactive`
  - `"ListMessage"` → `<biz><list type="product_list" v="2"/></biz>`

Equivalente Node.js no Evolution 2.3.7:
`src/api/integrations/channel/whatsapp/helpers/interactiveMessage.helper.ts` →
`buildInteractiveBizNode()` / `buildListBizNode()`.

**2. Carrossel ainda silenciado mesmo com `<biz>`**

Mesmo após (1), o carrossel retornava success mas não chegava. Diferenças estruturais
encontradas comparando com o `carouselMessage` que funciona no 2.3.7:

| Campo | Antes (Go) | Depois (Go = 2.3.7) |
|---|---|---|
| `msg.MessageContextInfo` | `{ DeviceListMetadata: {} }` | omitido |
| `interactiveMessage.Footer` (top-level) | preenchido com `data.Footer` | omitido |
| `interactiveMessage.ContextInfo` (top-level) | sempre `{}` | só setado se há `quoted` |
| `cards[].Header` | sempre criado com `Title`/`Subtitle`/`HasMediaAttachment:false` | só criado se há `imageUrl`/`videoUrl`, e descartado se upload falhar |

Aplicado em `SendCarousel` (`pkg/sendMessage/service/send_service.go:2545+`).

### Lista — investigação final

Trajetória de erros e o que foi descoberto:

1. **Antes da correção:** `405` (bloqueado pelo servidor).
2. **Após injetar `<biz><list type="product_list" v="2"/></biz>` via AdditionalNodes:** `479`
   (`smax-invalid` — stanza inválida). Causa: o whatsmeow já injeta automaticamente
   `<biz><list v="2" type="single_select"/></biz>` em
   [`whatsmeow-lib/send.go:1137-1145`](../whatsmeow-lib/send.go#L1137) via
   `getButtonTypeFromMessage`/`getButtonAttributes`. O nosso AdditionalNode duplica o `<biz>`,
   o que o servidor rejeita.
3. **Após remover o AdditionalNode para ListMessage** (mantendo só o automático): de volta ao
   `405 not-allowed` — referência em
   [`whatsmeow-lib/errors.go:187`](../whatsmeow-lib/errors.go#L187):
   `ErrIQNotAllowed = &IQError{Code: 405, Text: "not-allowed"}`.

**Conclusão:** o `405` é um bloqueio real do servidor da Meta — o formato `ListMessage` legacy
está sendo descontinuado em favor de `NativeFlowMessage` (botões/carrossel). Contas pessoais
e WhatsApp Business App não conseguem mais enviá-lo. Mesmo bloqueio afeta o Evolution 2.3.7
(Baileys) na mesma conta.

**Workarounds:**
- Para menus pequenos (até 3 opções): usar `/send/button` com botões REPLY.
- Para menus maiores (até 10 cards swipeable, mobile-only): usar `/send/carousel`.
- Para listas verdadeiras: precisaria de conta na WhatsApp Business Platform (Cloud API),
  que não passa por essa restrição.

### Sender → recipient delivery matrix

| Sender | Cliente do recipient | Botão | Lista | Carrossel |
|---|---|---|---|---|
| Senhor Colchão (Business App) | Mobile (Android/iOS) | ✅ | ❌ (479) | ✅ |
| Senhor Colchão (Business App) | WhatsApp Desktop/Web | (botão chega) | n/a | ❌ (limitação do cliente) |

## Health and license

```bash
curl -i -sS http://localhost:8093/server/ok
curl -i -sS http://localhost:8093/license/status
curl -i -sS "http://localhost:8093/license/register?redirect_uri=http://localhost:8093/manager/license/callback"
```

Manager:

```text
http://localhost:8093/manager/login
```

Use:

```text
API URL: http://localhost:8093
GLOBAL_API_KEY: BQYHJGJHJ123
```

## Create and connect instance

After license activation:

```bash
curl -X POST "http://localhost:8093/instance/create" \
  -H "Content-Type: application/json" \
  -H "apikey: BQYHJGJHJ123" \
  -d '{"name":"interactive-go-test","token":"interactive-go-token-123"}'

curl -X POST "http://localhost:8093/instance/connect" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123"

curl "http://localhost:8093/instance/qr" \
  -H "apikey: interactive-go-token-123"

curl "http://localhost:8093/instance/status" \
  -H "apikey: interactive-go-token-123"
```

## Test buttons

```bash
curl -X POST "http://localhost:8093/send/button" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","title":"Escolha uma opcao","description":"Teste de botoes no Evolution Go","footer":"Evolution Go","buttons":[{"type":"reply","displayText":"Comprar","id":"buy"},{"type":"reply","displayText":"Atendente","id":"agent"}]}'
```

## Test copy button for PIX

```bash
curl -X POST "http://localhost:8093/send/button" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","title":"Pagamento via PIX","description":"Clique para copiar a chave PIX","footer":"Senhor Colchao","buttons":[{"type":"copy","displayText":"Copiar chave PIX","copyCode":"279949240007"}]}'
```

## Test list

```bash
curl -X POST "http://localhost:8093/send/list" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","title":"Menu de atendimento","description":"Escolha uma opcao abaixo","footerText":"Senhor Colchao","buttonText":"Ver opcoes","sections":[{"title":"Vendas","rows":[{"title":"Comprar colchao","description":"Ver modelos disponiveis","rowId":"sales_mattress"},{"title":"Consultar pedido","description":"Acompanhar status","rowId":"sales_order_status"}]},{"title":"Suporte","rows":[{"title":"Falar com atendente","description":"Abrir atendimento humano","rowId":"support_agent"}]}]}'
```

## Test carousel

> Atenção: cada card usa `body: { "text": "..." }` (objeto), enquanto o `body` no
> nível raiz e `footer` são strings simples.

```bash
curl -X POST "http://localhost:8093/send/carousel" \
  -H "Content-Type: application/json" \
  -H "apikey: interactive-go-token-123" \
  -d '{"number":"5517981189332","body":"Modelos em destaque","footer":"Senhor Colchao","cards":[{"header":{"title":"Colchao Premium","imageUrl":"https://picsum.photos/600/400?random=11"},"body":{"text":"Conforto alto e entrega rapida."},"footer":"A partir de R$ 999","buttons":[{"type":"REPLY","displayText":"Quero esse","id":"premium"},{"type":"COPY","displayText":"Copiar PIX","copyCode":"279949240007"}]},{"header":{"title":"Colchao Luxo","imageUrl":"https://picsum.photos/600/400?random=12"},"body":{"text":"Modelo reforcado para casal."},"footer":"A partir de R$ 1299","buttons":[{"type":"REPLY","displayText":"Ver luxo","id":"luxo"},{"type":"URL","displayText":"Abrir site","id":"https://senhorcolchao.com"}]}]}'
```

