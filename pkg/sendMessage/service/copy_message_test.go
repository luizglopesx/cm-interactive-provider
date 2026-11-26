package send_service

import (
	"bytes"
	"encoding/json"
	waBinary "go.mau.fi/whatsmeow/binary"
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func TestCopyButtonPreservesCodeAndEscapesJSON(t *testing.T) {
	input := Button{DisplayText: "Copiar \"PIX\"", CopyCode: "synthetic\ncode\""}
	button := buildCopyNativeFlowButton(input)
	if button.GetName() != "cta_copy" {
		t.Fatal("copy must remain cta_copy")
	}
	var params map[string]string
	if err := json.Unmarshal([]byte(button.GetButtonParamsJSON()), &params); err != nil {
		t.Fatal(err)
	}
	if params["copy_code"] != input.CopyCode || params["display_text"] != input.DisplayText || params["id"] == "" {
		t.Fatalf("invalid copy params: %#v", params)
	}
	if len(params) != 3 {
		t.Fatalf("unexpected params: %#v", params)
	}
	explicit := buildCopyNativeFlowButton(Button{Id: "copy-test", CopyCode: "code", DisplayText: "Copy"})
	if err := json.Unmarshal([]byte(explicit.GetButtonParamsJSON()), &params); err != nil || params["id"] != "copy-test" {
		t.Fatal("explicit ID must be preserved")
	}
}

func TestCopyEnvelopePreservesInteractiveAndUsesFreshSecret(t *testing.T) {
	inner := &waE2E.Message{InteractiveMessage: &waE2E.InteractiveMessage{Body: &waE2E.InteractiveMessage_Body{Text: proto.String("Synthetic copy test")}}}
	first, directNodes, err := prepareCopyMessage(inner, false)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(first.GetDocumentWithCaptionMessage().GetMessage(), inner) {
		t.Fatal("interactive payload changed")
	}
	if len(first.GetMessageContextInfo().GetMessageSecret()) != 32 {
		t.Fatal("missing 32-byte message secret")
	}
	if len(directNodes) != 2 || directNodes[0].Tag != "biz" || directNodes[1].Tag != "bot" || directNodes[1].Attrs["biz_bot"] != "1" {
		t.Fatalf("missing direct-chat metadata: %#v", directNodes)
	}
	interactive := directNodes[0].Content.([]waBinary.Node)[0]
	native := interactive.Content.([]waBinary.Node)[0]
	if interactive.Attrs["type"] != "native_flow" || interactive.Attrs["v"] != "1" || native.Attrs["name"] != "mixed" || len(native.Attrs) != 1 {
		t.Fatal("invalid copy business metadata")
	}
	second, groupNodes, err := prepareCopyMessage(inner, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(groupNodes) != 1 || groupNodes[0].Tag != "biz" {
		t.Fatal("group must keep existing metadata without direct-chat bot node")
	}
	if bytes.Equal(first.GetMessageContextInfo().GetMessageSecret(), second.GetMessageContextInfo().GetMessageSecret()) {
		t.Fatal("secret reused")
	}
}
