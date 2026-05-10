# CI/CD com GHCR + Portainer Webhook

## Visão geral

Fork do [evolution-foundation/evolution-go](https://github.com/evolution-foundation/evolution-go) com adição de endpoints interativos (botões, lista, carrossel). O pipeline de CI/CD faz o build da imagem Docker, publica no GitHub Container Registry (GHCR) e dispara automaticamente o redeploy no Portainer via webhook.

Repositório do fork: https://github.com/luizglopesx/cm-interactive-provider

---

## O que o fork adiciona ao upstream

### Endpoints `/send` disponíveis

| Endpoint | Descrição | Origem |
|---|---|---|
| `POST /send/text` | Texto simples | upstream |
| `POST /send/link` | Link com preview | upstream |
| `POST /send/media` | Imagem, vídeo, áudio, documento | upstream |
| `POST /send/poll` | Enquete | upstream |
| `POST /send/sticker` | Figurinha | upstream |
| `POST /send/location` | Localização | upstream |
| `POST /send/contact` | Contato | upstream |
| `POST /send/button` | Botões interativos | **fork** |
| `POST /send/list` | Lista de opções | **fork** |
| `POST /send/carousel` | Carrossel | **fork** |
| `POST /send/status/text` | Status texto | upstream |
| `POST /send/status/media` | Status mídia | upstream |

Demais grupos de rotas: instâncias, usuários, mensagens, chats, grupos, chamadas, comunidades, labels, newsletters e enquetes.

---

## Fluxo de auto-atualização

```
Push na branch main ou interactive-go-test
        ↓
GitHub Actions: build da imagem Docker (linux/amd64, ~2m43s)
        ↓
Push da imagem para GHCR (ghcr.io/<repositório>)
        ↓
curl POST no webhook do Portainer
        ↓
Portainer faz redeploy automático do stack
```

---

## Workflow

Arquivo: `.github/workflows/publish-to-ghcr.yml`

- Ativado por push nas branches `main` e `interactive-go-test`
- Builda apenas para `linux/amd64` (arm64 foi removido — ver decisões abaixo)
- Tags geradas:
  - `latest` sempre
  - `<versão>` (lida do arquivo `VERSION`) quando o push é na `main`
  - `<nome-da-branch>` quando o push é em outra branch
- Usa cache do GitHub Actions (`type=gha`) para acelerar builds subsequentes

---

## Configuração do webhook

O webhook do Portainer é configurado via secret do GitHub:

| Secret | Valor |
|---|---|
| `PORTAINER_WEBHOOK_URL` | URL do webhook do stack no Portainer |

O step no workflow verifica se o secret está configurado antes de disparar:

```yaml
- name: Trigger Portainer redeploy
  env:
    WEBHOOK_URL: ${{ secrets.PORTAINER_WEBHOOK_URL }}
  run: |
    if [ -n "$WEBHOOK_URL" ]; then
      curl -X POST "$WEBHOOK_URL"
      echo "Portainer redeploy triggered"
    else
      echo "No webhook configured — skipping Portainer update"
    fi
```

---

## Servidor

| Propriedade | Valor |
|---|---|
| Arquitetura | x86_64 (amd64) |
| Gerenciador | Portainer (Docker Swarm — manager1) |

---

## Remotes do git

| Remote | URL | Finalidade |
|---|---|---|
| `origin` | https://github.com/evolution-foundation/evolution-go.git | upstream oficial |
| `luiz` | https://github.com/luizglopesx/cm-interactive-provider | nosso fork |

---

## Como sincronizar com o upstream

Quando sair uma atualização no repositório original:

```bash
# 1. Buscar as mudanças do upstream
git fetch origin

# 2. Mudar para a main
git checkout main

# 3. Mergear o upstream
git merge origin/main
```

Se houver **conflitos** (provável nos arquivos abaixo), resolver manualmente mantendo as adições do fork:

- `pkg/routes/routes.go` — registro das rotas button, list e carousel
- `pkg/sendMessage/handler/send_handler.go` — handlers dos novos endpoints
- `pkg/sendMessage/service/send_service.go` — lógica de envio

Após resolver os conflitos:

```bash
# 4. Commitar a resolução
git add .
git commit

# 5. Push para o fork — já dispara o CI/CD e redeploy automático
git push luiz main
```

---

## Decisões tomadas

### Remoção do build multi-plataforma (commit `a0095aa`)

- **O que foi removido:** build para `linux/arm64` via QEMU e o step `docker/setup-qemu-action@v3`
- **Motivo:** a emulação QEMU fazia o Action demorar 20+ minutos por run
- **Impacto:** nenhum — o servidor roda `x86_64`, não precisa de ARM
- **Resultado:** build passou de 20+ min para ~2m43s

### Webhook configurado como secret (não hardcoded)

- A URL do webhook fica em `secrets.PORTAINER_WEBHOOK_URL` no GitHub
- O step é silencioso se o secret não estiver configurado, sem quebrar o pipeline
