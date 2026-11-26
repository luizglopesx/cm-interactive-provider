package send_service

import (
	"encoding/json"
	"testing"
)

func TestBuildNativeFlowListMessage(t *testing.T) {
	data := &ListStruct{
		Title:       "Teste",
		Description: "Escolha",
		ButtonText:  "Ver opções",
		FooterText:  "Loja",
		Sections: []Section{{Title: "Seção", Rows: []Row{
			{Title: "Chegou", Description: "sim", RowId: "lista_chegou"},
			{Title: "Sem id"},
		}}},
	}
	msg, err := buildNativeFlowListMessage(data, "123")
	if err != nil {
		t.Fatal(err)
	}
	if msg.ListMessage != nil || msg.DocumentWithCaptionMessage != nil {
		t.Fatal("lista não pode usar o ListMessage legado")
	}
	im := msg.GetInteractiveMessage()
	if im.GetBody().GetText() != "*Teste*\n\nEscolha" || im.GetFooter().GetText() != "Loja" {
		t.Fatalf("body/footer inesperados: %q / %q", im.GetBody().GetText(), im.GetFooter().GetText())
	}
	buttons := im.GetNativeFlowMessage().GetButtons()
	if len(buttons) != 1 || buttons[0].GetName() != "single_select" {
		t.Fatalf("esperado um botão single_select, veio %v", buttons)
	}
	var params nativeFlowListParams
	if err := json.Unmarshal([]byte(buttons[0].GetButtonParamsJSON()), &params); err != nil {
		t.Fatal(err)
	}
	if params.Title != "Ver opções" || len(params.Sections) != 1 || len(params.Sections[0].Rows) != 2 {
		t.Fatalf("params inesperados: %+v", params)
	}
	if params.Sections[0].Rows[0].ID != "lista_chegou" || params.Sections[0].Rows[1].ID != "row_0_1" {
		t.Fatalf("ids inesperados: %+v", params.Sections[0].Rows)
	}

	// Rodapé vazio não cria Footer (iOS).
	data.FooterText = ""
	msg, _ = buildNativeFlowListMessage(data, "123")
	if msg.GetInteractiveMessage().Footer != nil {
		t.Fatal("rodapé vazio não deveria criar Footer")
	}
}
