package send_service

import (
	cryptorand "crypto/rand"

	"go.mau.fi/whatsmeow/proto/waE2E"
)

func hasCopyButton(buttons []Button) bool {
	for _, b := range buttons {
		if b.Type == "copy" {
			return true
		}
	}
	return false
}

// iOS só exibe a mensagem interativa com um MessageSecret de 32 bytes. Mantém o
// InteractiveMessage sem embrulho: DocumentWithCaptionMessage faz o servidor
// recusar o copy CTA com erro 473 (evolution-go#110).
func withCopyMessageSecret(message *waE2E.Message) error {
	secret := make([]byte, 32)
	if _, err := cryptorand.Read(secret); err != nil {
		return err
	}
	message.MessageContextInfo = &waE2E.MessageContextInfo{MessageSecret: secret}
	return nil
}
