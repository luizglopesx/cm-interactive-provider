# CI/CD com GHCR + Portainer Webhook

## Visão geral

O pipeline de CI/CD faz o build da imagem Docker, publica no GitHub Container Registry (GHCR) e dispara automaticamente o redeploy no Portainer via webhook.

## Fluxo completo

```
Push na branch main ou interactive-go-test
        ↓
GitHub Actions: build da imagem Docker (linux/amd64)
        ↓
Push da imagem para GHCR (ghcr.io/<repositório>)
        ↓
curl POST no webhook do Portainer
        ↓
Portainer faz redeploy automático do stack
```

## Workflow

Arquivo: `.github/workflows/publish-to-ghcr.yml`

- Ativado por push nas branches `main` e `interactive-go-test`
- Builda apenas para `linux/amd64` (arm64 foi removido — ver decisões abaixo)
- Tags geradas:
  - `latest` sempre
  - `<versão>` (lida do arquivo `VERSION`) quando o push é na `main`
  - `<nome-da-branch>` quando o push é em outra branch
- Usa cache do GitHub Actions (`type=gha`) para acelerar builds subsequentes

## Configuração do webhook

O webhook do Portainer é configurado via secret do GitHub:

| Secret               | Valor                          |
|----------------------|--------------------------------|
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

## Servidor

| Propriedade  | Valor   |
|--------------|---------|
| Arquitetura  | x86_64 (amd64) |
| Gerenciador  | Portainer (Docker Swarm — manager1) |

## Decisões tomadas

### Remoção do build multi-plataforma (commit `a0095aa`)

- **O que foi removido:** build para `linux/arm64` via QEMU e o step `docker/setup-qemu-action@v3`
- **Motivo:** a emulação QEMU fazia o Action demorar 20+ minutos por run
- **Impacto:** nenhum — o servidor roda `x86_64`, não precisa de ARM
- **Resultado:** build passou de 20+ min para ~2m43s

### Webhook configurado como secret (não hardcoded)

- A URL do webhook fica em `secrets.PORTAINER_WEBHOOK_URL` no GitHub
- O step é silencioso se o secret não estiver configurado, sem quebrar o pipeline
