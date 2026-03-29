package monitor

import (
	"fmt"
	"net"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/TONresistor/tonrelay/internal/config"
	"github.com/TONresistor/tonrelay/internal/service"
	tunnelconfig "github.com/ton-blockchain/adnl-tunnel/config"
)

// Colors
var (
	colorDim    = lipgloss.Color("241")
	colorText   = lipgloss.Color("252")
	colorAccent = lipgloss.Color("75")
	colorOk     = lipgloss.Color("78")
	colorErr    = lipgloss.Color("203")
	colorBorder = lipgloss.Color("238")
)

// Styles
var (
	dimStyle = lipgloss.NewStyle().Foreground(colorDim)
	valStyle = lipgloss.NewStyle().Foreground(colorText)
	okStyle  = lipgloss.NewStyle().Foreground(colorOk)
	errStyle = lipgloss.NewStyle().Foreground(colorErr)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)
)

type tickMsg time.Time

type DashboardModel struct {
	monitor *Monitor
	cfg     *tunnelconfig.Config
	version string
	width   int
	height  int
}

func NewDashboardModel(cfg *tunnelconfig.Config, version string) DashboardModel {
	mon := New()
	mon.Start()
	return DashboardModel{monitor: mon, cfg: cfg, version: version}
}

func (m DashboardModel) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.monitor.Stop()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m, tickCmd()
	}
	return m, nil
}

func (m DashboardModel) View() string {
	w := m.width
	if w < 40 {
		w = 80
	}

	metrics := m.monitor.GetMetrics()
	svcStatus, _ := service.GetStatus()

	// -- Header --
	title := headerStyle.Render("tonrelay")
	version := dimStyle.Render(m.version)
	headerLine := title + " " + version
	separator := dimStyle.Render(strings.Repeat("─", max(0, w-lipgloss.Width(headerLine)-1)))
	header := headerLine + " " + separator

	// -- Status row (two panels side by side) --
	halfW := (w - 3) / 2 // gap between panels

	statusPanel := m.renderStatus(svcStatus, halfW)
	relayPanel := m.renderRelay(metrics, halfW)
	statusRow := lipgloss.JoinHorizontal(lipgloss.Top, statusPanel, " ", relayPanel)

	// -- Traffic panel (full width) --
	trafficPanel := m.renderTraffic(metrics, w-2)

	// -- Payments panel (if enabled) --
	var paymentsPanel string
	if m.cfg != nil && m.cfg.PaymentsEnabled {
		paymentsPanel = m.renderPayments(w - 2)
	}

	// -- Error line --
	var errorLine string
	if metrics.LastError != "" {
		errorLine = errStyle.Render("err: " + metrics.LastError)
	}

	// -- Footer --
	footer := dimStyle.Render("q quit  |  refreshes every 3s")

	// -- Compose --
	parts := []string{header, "", statusRow, "", trafficPanel}
	if paymentsPanel != "" {
		parts = append(parts, "", paymentsPanel)
	}
	if errorLine != "" {
		parts = append(parts, "", errorLine)
	}
	parts = append(parts, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m DashboardModel) renderStatus(svc *service.Status, w int) string {
	var rows []string

	if svc != nil && svc.Active {
		rows = append(rows, kv("state", okStyle.Render("running"), w-4))
		if svc.ActiveTime != "" {
			if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", svc.ActiveTime); err == nil {
				up := time.Since(t).Truncate(time.Second)
				rows = append(rows, kv("uptime", formatDuration(up), w-4))
			}
		}
		if svc.MainPID != "" && svc.MainPID != "0" {
			rows = append(rows, kv("pid", svc.MainPID, w-4))
		}
	} else {
		rows = append(rows, kv("state", errStyle.Render("stopped"), w-4))
	}

	content := strings.Join(rows, "\n")
	return panelStyle.Width(w - 2).Render(
		headerStyle.Render("service") + "\n" + content,
	)
}

func (m DashboardModel) renderRelay(metrics Metrics, w int) string {
	var rows []string

	if m.cfg != nil {
		id := config.GetADNLID(m.cfg)
		if len(id) > 16 {
			id = id[:16] + "..."
		}
		rows = append(rows, kv("adnl", id, w-4))
		if m.cfg.ExternalIP != "" {
			_, port, _ := net.SplitHostPort(m.cfg.TunnelListenAddr)
			if port == "" {
				port = "17330"
			}
			rows = append(rows, kv("endpoint", m.cfg.ExternalIP+":"+port, w-4))
		}
	}

	if metrics.DHTPublished {
		rows = append(rows, kv("dht", okStyle.Render("published"), w-4))
	} else {
		rows = append(rows, kv("dht", errStyle.Render("waiting"), w-4))
	}

	content := strings.Join(rows, "\n")
	return panelStyle.Width(w - 2).Render(
		headerStyle.Render("relay") + "\n" + content,
	)
}

func (m DashboardModel) renderTraffic(metrics Metrics, w int) string {
	colW := (w - 3) / 2

	col1 := miniPanel("packets", []string{
		kv("routed", FormatPackets(metrics.PacketsRouted), colW-4),
		kv("in", FormatPackets(metrics.PacketsRecv), colW-4),
		kv("out", FormatPackets(metrics.PacketsSent), colW-4),
	}, colW)

	col2 := miniPanel("network", []string{
		kv("inbound", fmt.Sprintf("%d", metrics.InboundSect), colW-4),
		kv("gateways", fmt.Sprintf("%d", metrics.OutGateways), colW-4),
		kv("routes", fmt.Sprintf("%d", metrics.ActiveRoutes+metrics.ActiveRoutesP), colW-4),
	}, colW)

	return lipgloss.JoinHorizontal(lipgloss.Top, col1, " ", col2)
}

func (m DashboardModel) renderPayments(w int) string {
	var rows []string
	if m.cfg != nil {
		rows = append(rows, kv("route", fmt.Sprintf("%d nano/pkt", m.cfg.Payments.MinPricePerPacketRoute), w-6))
		rows = append(rows, kv("out", fmt.Sprintf("%d nano/pkt", m.cfg.Payments.MinPricePerPacketInOut), w-6))
	}
	content := strings.Join(rows, "\n")
	return panelStyle.Width(w).Render(
		headerStyle.Render("payments") + "\n" + content,
	)
}

// -- helpers --

func kv(key, value string, w int) string {
	keyStr := dimStyle.Render(key)
	keyW := lipgloss.Width(keyStr)
	gap := w - keyW - lipgloss.Width(value)
	if gap < 1 {
		gap = 1
	}
	return keyStr + strings.Repeat(" ", gap) + valStyle.Render(value)
}

func miniPanel(title string, rows []string, w int) string {
	content := strings.Join(rows, "\n")
	return panelStyle.Width(w).Render(
		headerStyle.Render(title) + "\n" + content,
	)
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	return fmt.Sprintf("%dm %ds", mins, secs)
}

