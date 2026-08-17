package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestTelegramEnabled(t *testing.T) {
	tests := []struct {
		name  string
		token string
		chat  string
		want  bool
	}{
		{name: "both empty", want: false},
		{name: "empty token", token: "", chat: "12345", want: false},
		{name: "empty chat", token: "123:ABC", chat: "", want: false},
		{name: "whitespace only", token: "  ", chat: "   ", want: false},
		{name: "non-numeric chat", token: "123:ABC", chat: "not-a-chat-id", want: false},
		{name: "username chat", token: "123:ABC", chat: "@mychannel", want: false},
		{name: "valid positive chat", token: "123:ABC", chat: "12345", want: true},
		{name: "valid negative group chat", token: "123:ABC", chat: "-1001234567890", want: true},
		{name: "padded numeric chat", token: " 123:ABC ", chat: " 12345 ", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AlertmanagerConfig{}
			cfg.TelegramBotToken.Value = tt.token
			cfg.TelegramChatID.Value = tt.chat
			if got := cfg.TelegramEnabled(); got != tt.want {
				t.Fatalf("TelegramEnabled() = %v, want %v (token=%q chat=%q)", got, tt.want, tt.token, tt.chat)
			}
		})
	}
}

func TestAlertmanagerTemplateTelegramConfigs(t *testing.T) {
	tmplPath := filepath.Join("..", "rocketpool", "assets", "install", "alerting", "alertmanager.tmpl")
	src, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("read alertmanager template: %v", err)
	}

	tmpl, err := template.New("alertmanager.tmpl").Parse(string(src))
	if err != nil {
		t.Fatalf("parse alertmanager template: %v", err)
	}

	cfg := &AlertmanagerConfig{}
	cfg.TelegramBotToken.Value = "123:ABC"
	cfg.TelegramChatID.Value = "-1001234567890"

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		t.Fatalf("execute alertmanager template: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "telegram_configs:") {
		t.Fatalf("expected telegram_configs when Telegram is enabled:\n%s", out)
	}
	if !strings.Contains(out, `bot_token: "123:ABC"`) {
		t.Fatalf("expected quoted bot token:\n%s", out)
	}
	if !strings.Contains(out, "chat_id: -1001234567890") {
		t.Fatalf("expected unquoted chat ID:\n%s", out)
	}
	if strings.Count(out, "telegram_configs:") != 2 {
		t.Fatalf("expected telegram_configs on both receivers, found %d:\n%s", strings.Count(out, "telegram_configs:"), out)
	}
	if strings.Count(out, "send_resolved: false") < 1 {
		t.Fatalf("expected send_resolved: false on the info receiver:\n%s", out)
	}

	disabled := &AlertmanagerConfig{}
	disabled.TelegramBotToken.Value = "123:ABC"
	disabled.TelegramChatID.Value = "not-a-chat-id"
	buf.Reset()
	if err := tmpl.Execute(&buf, disabled); err != nil {
		t.Fatalf("execute alertmanager template with invalid chat ID: %v", err)
	}
	if strings.Contains(buf.String(), "telegram_configs:") {
		t.Fatalf("did not expect telegram_configs for invalid chat ID:\n%s", buf.String())
	}
}
