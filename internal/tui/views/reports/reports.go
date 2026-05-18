package reports

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/tui/styles"
)

type Period int

const (
	PeriodWeekly Period = iota
	PeriodMonthly
	PeriodDaily
)

type BackMsg struct{}

const nameColWidth = 20

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type Model struct {
	period           Period
	dailyTotals      []repository.ProjectTotal
	weeklyBreakdown  []repository.DailyProjectBreakdown
	monthlyBreakdown []repository.DailyProjectBreakdown
	now              time.Time
}

// New creates a reports model. Defaults to weekly mode.
func New(
	daily []repository.ProjectTotal,
	weekly []repository.DailyProjectBreakdown,
	monthly []repository.DailyProjectBreakdown,
) Model {
	return Model{
		period:           PeriodWeekly,
		dailyTotals:      daily,
		weeklyBreakdown:  weekly,
		monthlyBreakdown: monthly,
		now:              time.Now().UTC(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "d":
		m.period = PeriodDaily
	case "w":
		m.period = PeriodWeekly
	case "m":
		m.period = PeriodMonthly
	case "esc", "q":
		return m, func() tea.Msg { return BackMsg{} }
	}
	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString(styles.Title.Render("reports — "+m.periodHeader()) + "\n\n")

	switch m.period {
	case PeriodDaily:
		renderDaily(&sb, m.dailyTotals)
	case PeriodWeekly:
		ws := weekStart(m.now)
		renderChart(&sb, m.weeklyBreakdown, ws, 7)
	case PeriodMonthly:
		ms := monthStart(m.now)
		days := ms.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()
		renderChart(&sb, m.monthlyBreakdown, ms, days)
	}

	sb.WriteString("\n" + styles.StatusBar.Render("  [d] daily  [w] weekly  [m] monthly  [esc] back"))
	return sb.String()
}

func (m Model) periodHeader() string {
	switch m.period {
	case PeriodDaily:
		return m.now.Format("Mon 2 Jan 2006")
	case PeriodWeekly:
		ws := weekStart(m.now)
		we := ws.AddDate(0, 0, 6)
		if ws.Month() == we.Month() {
			return fmt.Sprintf("%s %d – %s %d %s",
				ws.Format("Mon"), ws.Day(),
				we.Format("Mon"), we.Day(),
				ws.Format("Jan 2006"))
		}
		return fmt.Sprintf("%s – %s", ws.Format("Mon 2 Jan"), we.Format("Mon 2 Jan 2006"))
	case PeriodMonthly:
		return m.now.Format("Jan 2006")
	}
	return ""
}

func renderDaily(sb *strings.Builder, totals []repository.ProjectTotal) {
	if len(totals) == 0 {
		sb.WriteString(styles.Muted.Render("  No sessions today.") + "\n")
		return
	}
	for _, pt := range totals {
		fmt.Fprintf(sb, "  %-*s  %s\n",
			nameColWidth, pt.ProjectName,
			styles.Accent.Render(styles.FormatDuration(pt.Total)))
	}
}

type projectRow struct {
	name  string
	days  []time.Duration
	total time.Duration
}

func buildRows(breakdown []repository.DailyProjectBreakdown, periodStart time.Time, numDays int) []projectRow {
	indexByID := map[int64]int{}
	var rows []projectRow

	for _, b := range breakdown {
		idx, ok := indexByID[b.ProjectID]
		if !ok {
			idx = len(rows)
			indexByID[b.ProjectID] = idx
			rows = append(rows, projectRow{
				name: b.ProjectName,
				days: make([]time.Duration, numDays),
			})
		}
		dayIdx := int(b.Date.Sub(periodStart) / (24 * time.Hour))
		if dayIdx >= 0 && dayIdx < numDays {
			rows[idx].days[dayIdx] += b.Total
			rows[idx].total += b.Total
		}
	}
	return rows
}

func renderChart(sb *strings.Builder, breakdown []repository.DailyProjectBreakdown, periodStart time.Time, numDays int) {
	rows := buildRows(breakdown, periodStart, numDays)
	if len(rows) == 0 {
		sb.WriteString(styles.Muted.Render("  No sessions in this period.") + "\n")
		return
	}
	for _, row := range rows {
		spark := sparkline(row.days)
		fmt.Fprintf(sb, "  %-*s  %s   %s\n",
			nameColWidth, row.name,
			styles.Accent.Render(spark),
			styles.Accent.Render(styles.FormatDuration(row.total)))
	}
}

func sparkline(values []time.Duration) string {
	var maxVal time.Duration
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	runes := make([]rune, len(values))
	for i, v := range values {
		switch {
		case maxVal == 0:
			runes[i] = sparkChars[0]
		case v == maxVal:
			runes[i] = sparkChars[7]
		default:
			idx := int(float64(v) / float64(maxVal) * 7)
			runes[i] = sparkChars[idx]
		}
	}
	return string(runes)
}

// weekStart returns Monday midnight UTC of the week containing t.
func weekStart(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7 // Sunday
	}
	mon := t.AddDate(0, 0, -(wd - 1))
	return time.Date(mon.Year(), mon.Month(), mon.Day(), 0, 0, 0, 0, time.UTC)
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
