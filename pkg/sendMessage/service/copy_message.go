package send_service

import (
	cryptorand "crypto/rand"
	"encoding/json"

	"github.com/evolution-foundation/evolution-go/pkg/utils"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// Copy CTA keeps the public copyCode contract and adds its internal native-flow ID.
func buildCopyNativeFlowButton(button Button) *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton {
	id := button.Id
	if id == "" {
		id = "copy_" + utils.GenerateRandomString(12)
	}
	encoded, _ := json.Marshal(map[string]string{"display_text": button.DisplayText, "id": id, "copy_code": button.CopyCode})
	return &waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{Name: proto.String("cta_copy"), ButtonParamsJSON: proto.String(string(encoded))}
}

// Align copy CTA rendering with upstream without changing reply, carousel or native Pix paths.
func prepareCopyMessage(message *waE2E.Message, group bool) (*waE2E.Message, []waBinary.Node, error) {
	secret := make([]byte, 32)
	if _, err := cryptorand.Read(secret); err != nil {
		return nil, nil, err
	}
	wrapped := &waE2E.Message{
		DocumentWithCaptionMessage: &waE2E.FutureProofMessage{Message: message},
		MessageContextInfo:         &waE2E.MessageContextInfo{MessageSecret: secret},
	}
	nodes := []waBinary.Node{{Tag: "biz", Content: []waBinary.Node{{
		Tag: "interactive", Attrs: waBinary.Attrs{"type": "native_flow", "v": "1"},
		Content: []waBinary.Node{{Tag: "native_flow", Attrs: waBinary.Attrs{"name": "mixed"}}},
	}}}}
	if !group {
		nodes = append(nodes, waBinary.Node{Tag: "bot", Attrs: waBinary.Attrs{"biz_bot": "1"}})
	}
	return wrapped, nodes, nil
}
