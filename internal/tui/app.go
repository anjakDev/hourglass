package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/anjakDev/hourglass/internal/timer"
	"github.com/anjakDev/hourglass/internal/tui/styles"
	"github.com/anjakDev/hourglass/internal/tui/views/activetimer"
	"github.com/anjakDev/hourglass/internal/tui/views/newproject"
	pv "github.com/anjakDev/hourglass/internal/tui/views/projects"
	"github.com/anjakDev/hourglass/internal/tui/views/sessionlog"
)

type viewID int

const (
	viewProjects   viewID = iota
	viewNewProject        // new-project text-input form
	viewActiveTimer       // live running timer
	viewSessionLog        // read-only today's session list
)

// Internal messages produced by tea.Cmd closures after DB operations.
type projectsLoadedMsg struct{ items []repository.ProjectTotal }
type projectCreatedMsg struct{}
type projectArchivedMsg struct{}
type sessionStartedMsg struct {
	sessionID   int64
	projectID   int64
	projectName string
	startedAt   time.Time
	err         error
}
type sessionStoppedMsg struct{}
type sessionLogLoadedMsg struct {
	projectName string
	sessions    []repository.Session
}

// App is the root Bubbletea model. It owns the repos and timer, manages
// the active view, and handles all DB-writing commands.
type App struct {
	active      viewID
	projectRepo *repository.ProjectRepo
	sessionRepo *repository.SessionRepo
	timer       *timer.Timer

	activeSessionID int64
	activeProjectID int64
	width, height   int

	projects    pv.Model
	newProject  newproject.Model
	activeTimer activetimer.Model
	sessionLog  sessionlog.Model
}

// New constructs the root app model. Repos are shared with all cmd closures.
func New(pr *repository.ProjectRepo, sr *repository.SessionRepo) App {
	return App{
		active:      viewProjects,
		projectRepo: pr,
		sessionRepo: sr,
		timer:       timer.New(),
		projects:    pv.New(),
		newProject:  newproject.New(),
	}
}

// Init loads today's project totals on startup.
func (a App) Init() tea.Cmd { return a.loadProjects() }

// loadProjects fetches all active projects and merges in today's session
// totals. Projects with no sessions today appear with Total == 0.
func (a App) loadProjects() tea.Cmd {
	return func() tea.Msg {
		projs, err := a.projectRepo.List()
		if err != nil {
			return projectsLoadedMsg{}
		}
		totals, _ := a.sessionRepo.TodayTotalsByProject()
		byID := make(map[int64]time.Duration, len(totals))
		for _, t := range totals {
			byID[t.ProjectID] = t.Total
		}
		items := make([]repository.ProjectTotal, len(projs))
		for i, p := range projs {
			items[i] = repository.ProjectTotal{
				ProjectID:   p.ID,
				ProjectName: p.Name,
				Total:       byID[p.ID],
			}
		}
		return projectsLoadedMsg{items: items}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		return a, nil

	case projectsLoadedMsg:
		a.projects = a.projects.SetItems(msg.items)
		return a, nil

	// ── projects view ─────────────────────────────────────────────────────

	case pv.NewProjectMsg:
		a.newProject = newproject.New()
		a.active = viewNewProject
		return a, a.newProject.Init()

	case pv.ArchiveMsg:
		id := msg.ProjectID
		return a, func() tea.Msg {
			_ = a.projectRepo.Archive(id)
			return projectArchivedMsg{}
		}

	case projectArchivedMsg:
		return a, a.loadProjects()

	case pv.StartSessionMsg:
		now := time.Now().UTC()
		if err := a.timer.Start(now); err != nil {
			return a, nil // already running
		}
		id := msg.ProjectID
		return a, func() tea.Msg {
			sid, err := a.sessionRepo.StartSession(id, now)
			if err != nil {
				return sessionStartedMsg{err: err}
			}
			proj, err := a.projectRepo.GetByID(id)
			if err != nil {
				return sessionStartedMsg{err: err}
			}
			return sessionStartedMsg{
				sessionID:   sid,
				projectID:   id,
				projectName: proj.Name,
				startedAt:   now,
			}
		}

	case sessionStartedMsg:
		if msg.err != nil {
			_, _ = a.timer.Stop(time.Now().UTC()) // undo in-memory start
			return a, nil
		}
		a.activeSessionID = msg.sessionID
		a.activeProjectID = msg.projectID
		a.activeTimer = activetimer.New(msg.projectName, a.timer, msg.startedAt)
		a.active = viewActiveTimer
		return a, a.activeTimer.Init()

	case pv.ShowSessionLogMsg:
		id := msg.ProjectID
		return a, func() tea.Msg {
			proj, err := a.projectRepo.GetByID(id)
			if err != nil {
				return sessionLogLoadedMsg{projectName: "unknown"}
			}
			all, err := a.sessionRepo.ListToday()
			if err != nil {
				return sessionLogLoadedMsg{projectName: proj.Name}
			}
			var filtered []repository.Session
			for _, s := range all {
				if s.ProjectID == id {
					filtered = append(filtered, s)
				}
			}
			return sessionLogLoadedMsg{projectName: proj.Name, sessions: filtered}
		}

	case sessionLogLoadedMsg:
		a.sessionLog = sessionlog.New(msg.projectName, msg.sessions)
		a.active = viewSessionLog
		return a, nil

	// ── newproject view ───────────────────────────────────────────────────

	case newproject.CreatedMsg:
		name := msg.Name
		return a, func() tea.Msg {
			_, _ = a.projectRepo.Create(name, "", "")
			return projectCreatedMsg{}
		}

	case projectCreatedMsg:
		a.active = viewProjects
		return a, a.loadProjects()

	case newproject.CancelMsg:
		a.active = viewProjects
		return a, nil

	// ── activetimer view ──────────────────────────────────────────────────

	case activetimer.StopSessionMsg:
		now := time.Now().UTC()
		sess, err := a.timer.Stop(now)
		if err != nil {
			return a, nil // timer wasn't running
		}
		sid := a.activeSessionID
		endedAt := sess.EndedAt
		breakSec := int64(sess.TotalBreak.Seconds())
		return a, func() tea.Msg {
			_ = a.sessionRepo.StopSession(sid, endedAt, breakSec)
			return sessionStoppedMsg{}
		}

	case sessionStoppedMsg:
		a.activeSessionID = 0
		a.activeProjectID = 0
		a.active = viewProjects
		return a, a.loadProjects()

	// ── sessionlog view ───────────────────────────────────────────────────

	case sessionlog.BackMsg:
		a.active = viewProjects
		return a, nil
	}

	return a.forwardToActiveView(msg)
}

func (a App) forwardToActiveView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.active {
	case viewProjects:
		next, cmd := a.projects.Update(msg)
		a.projects = next.(pv.Model)
		return a, cmd
	case viewNewProject:
		next, cmd := a.newProject.Update(msg)
		a.newProject = next.(newproject.Model)
		return a, cmd
	case viewActiveTimer:
		next, cmd := a.activeTimer.Update(msg)
		a.activeTimer = next.(activetimer.Model)
		return a, cmd
	case viewSessionLog:
		next, cmd := a.sessionLog.Update(msg)
		a.sessionLog = next.(sessionlog.Model)
		return a, cmd
	}
	return a, nil
}

func (a App) View() string {
	switch a.active {
	case viewProjects:
		return a.projects.View()
	case viewNewProject:
		return a.newProject.View()
	case viewActiveTimer:
		return a.activeTimer.View()
	case viewSessionLog:
		return a.sessionLog.View()
	default:
		return styles.Muted.Render("loading…")
	}
}
