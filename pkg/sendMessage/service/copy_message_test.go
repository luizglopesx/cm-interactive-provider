package send_service

import (
	"bytes"
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func TestHasCopyButton(t *testing.T) {
	if !hasCopyButton([]Button{{Type: "url"}, {Type: "copy"}}) {
		t.Fatal("copy button not detected")
	}
	if hasCopyButton([]Button{{Type: "reply"}, {Type: "url"}}) {
		t.Fatal("false copy detection")
	}
}

func TestCopyMessageSecretKeepsInteractiveUnwrapped(t *testing.T) {
	newMsg := func() *waE2E.Message {
		return &waE2E.Message{InteractiveMessage: &waE2E.InteractiveMessage{Body: &waE2E.InteractiveMessage_Body{Text: proto.String("Synthetic copy test")}}}
	}
	first, second := newMsg(), newMsg()
	inner := proto.Clone(first.GetInteractiveMessage())
	if err := withCopyMessageSecret(first); err != nil {
		t.Fatal(err)
	}
	if err := withCopyMessageSecret(second); err != nil {
		t.Fatal(err)
	}
	if first.GetDocumentWithCaptionMessage() != nil {
		t.Fatal("copy message must not be wrapped (server error 473)")
	}
	if !proto.Equal(first.GetInteractiveMessage(), inner) {
		t.Fatal("interactive payload changed")
	}
	if len(first.GetMessageContextInfo().GetMessageSecret()) != 32 {
		t.Fatal("missing 32-byte message secret")
	}
	if bytes.Equal(first.GetMessageContextInfo().GetMessageSecret(), second.GetMessageContextInfo().GetMessageSecret()) {
		t.Fatal("secret reused")
	}
}
