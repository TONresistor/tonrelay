package installer

import (
	"strings"
	"testing"
	"text/template"
)

func TestSystemdUnitTemplate(t *testing.T) {
	opts := Options{
		ConfigPath: "/etc/tonrelay/config.json",
		DataDir:    "/var/lib/tonrelay",
		BinaryPath: "/usr/local/bin/tunnel-node",
		User:       "tonrelay",
	}

	tmpl, err := template.New("unit").Parse(systemdUnitTemplate)
	if err != nil {
		t.Fatalf("template parse: %v", err)
	}

	var b strings.Builder
	if err := tmpl.Execute(&b, opts); err != nil {
		t.Fatalf("template execute: %v", err)
	}

	unit := b.String()

	checks := []string{
		"User=tonrelay",
		"/usr/local/bin/tunnel-node",
		"-config /etc/tonrelay/config.json",
		"-log-disable-file",
		"WorkingDirectory=/var/lib/tonrelay",
		"NoNewPrivileges=yes",
		"ProtectSystem=strict",
		"ReadWritePaths=/var/lib/tonrelay",
		"WantedBy=multi-user.target",
		"LimitCORE=0",
	}

	for _, check := range checks {
		if !strings.Contains(unit, check) {
			t.Errorf("unit file missing: %q", check)
		}
	}
}

func TestSystemdUnitNeverRoot(t *testing.T) {
	opts := Options{
		User: "tonrelay",
	}

	tmpl, _ := template.New("unit").Parse(systemdUnitTemplate)
	var b strings.Builder
	tmpl.Execute(&b, opts)

	unit := b.String()
	if strings.Contains(unit, "User=root") {
		t.Error("systemd unit must never run as root")
	}
}
