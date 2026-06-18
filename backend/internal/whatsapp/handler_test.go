package whatsapp

import (
	"testing"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func TestExtractInboundMediaDocument(t *testing.T) {
	ts := time.Unix(1710000000, 0)
	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			FileName: proto.String("invoice.pdf"),
			Mimetype: proto.String("application/pdf"),
			Caption:  proto.String("cek invoice"),
		},
	}

	media := extractInboundMedia(msg, ts)
	if media == nil {
		t.Fatal("expected document media")
	}
	if media.messageType != "document" {
		t.Fatalf("messageType = %q, want document", media.messageType)
	}
	if media.caption != "cek invoice" {
		t.Fatalf("caption = %q, want cek invoice", media.caption)
	}
	if media.fileName != "invoice.pdf" {
		t.Fatalf("fileName = %q, want invoice.pdf", media.fileName)
	}
	if media.mimeType != "application/pdf" {
		t.Fatalf("mimeType = %q, want application/pdf", media.mimeType)
	}
	if media.download == nil {
		t.Fatal("expected downloadable document")
	}
}

func TestMediaFileNameFallsBackFromMimeType(t *testing.T) {
	got := mediaFileName("document", time.Unix(1710000000, 0), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "")
	want := "document_1710000000.xlsx"
	if got != want {
		t.Fatalf("mediaFileName() = %q, want %q", got, want)
	}
}

func TestForwardableWithoutTextIncludesMedia(t *testing.T) {
	for _, messageType := range []string{"image", "document", "audio", "video", "sticker", "location", "live_location"} {
		if !isWebhookForwardableWithoutText(messageType) {
			t.Fatalf("%s should be forwardable without text", messageType)
		}
	}

	if isWebhookForwardableWithoutText("text") {
		t.Fatal("text should not be forwardable without text")
	}
}
