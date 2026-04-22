package kanban

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rpcarvs/faz/internal/model"
)

const (
	refreshInterval     = 10 * time.Second
	minRenderableWidth  = 42
	minRenderableHeight = 11
	minModalWidth       = 24
	headerLines         = 2
	columnHeaderLines   = 1
	cardOuterHeight     = 7
	footerGap           = " • "
)

var typeFilterOptions = []string{"all", "task", "bug", "feature", "chore", "decision"}

type catalogLoadedMsg struct {
	catalog Catalog
	err     error
}

// detailsLoadedMsg carries dependency details for the currently inspected issue.
type detailsLoadedMsg struct {
	issueID string
	details issueDetails
	err     error
}

type refreshTickMsg time.Time

// issueDetails holds dependency context for the kanban detail modal.
type issueDetails struct {
	Dependencies []model.Issue
	Dependents   []model.Issue
	Loading      bool
	Err          error
}

// Model manages the read-only kanban TUI state.
type Model struct {
	svc Service

	width  int
	height int
	ready  bool

	catalog Catalog
	err     error

	scopeIndex  int
	selectedCol int
	selectedRow int
	scrollRow   int

	showPicker  bool
	pickerIndex int
	showHelp    bool
	showEpic    bool
	showType    bool
	typeIndex   int
	typeFilter  string

	showDetails      bool
	inspectedIssueID string
	inspectedIssue   model.Issue
	details          map[string]issueDetails
}

// NewModel builds a new kanban TUI model.
func NewModel(svc Service) Model {
	return Model{svc: svc, typeFilter: typeFilterOptions[0]}
}

// Init starts the first data load and background refresh loop.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadCatalogCmd(), tickCmd())
}

// Update handles input, resizing, refreshes, and modal state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.applyWindowSize(msg.Width, msg.Height)
		m.ensureSelection()
		return m, nil

	case catalogLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.catalog = msg.catalog
		if m.scopeIndex >= len(m.catalog.Scopes) {
			m.scopeIndex = 0
		}
		if m.showDetails {
			m.syncSelectionToInspectedIssue()
		} else {
			m.ensureSelection()
		}
		return m, nil

	case detailsLoadedMsg:
		if m.details == nil {
			m.details = make(map[string]issueDetails)
		}
		if msg.err != nil {
			m.details[msg.issueID] = issueDetails{Err: msg.err}
			return m, nil
		}
		m.details[msg.issueID] = msg.details
		return m, nil

	case refreshTickMsg:
		return m, tea.Batch(m.loadCatalogCmd(), tickCmd())

	case tea.KeyMsg:
		if m.showDetails {
			switch msg.String() {
			case "enter", "esc", "q":
				m.showDetails = false
				m.clearInspectedIssue()
			case "left", "h":
				return m, m.navigateDetailsCol(-1)
			case "right", "l":
				return m, m.navigateDetailsCol(1)
			case "up", "k":
				return m, m.navigateDetails(-1)
			case "down", "j":
				return m, m.navigateDetails(1)
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.showEpic {
			switch msg.String() {
			case "enter", "esc", "q":
				m.showEpic = false
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.showType {
			return m.updateTypePicker(msg)
		}
		if m.showPicker {
			return m.updatePicker(msg)
		}
		if m.showHelp {
			return m.updateHelp(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.cycleScope(1)
			return m, nil
		case "shift+tab":
			m.cycleScope(-1)
			return m, nil
		case "a":
			m.selectScopeByKey(scopeAll)
			return m, nil
		case "e":
			m.showPicker = true
			m.pickerIndex = m.scopeIndex
			return m, nil
		case "o":
			m.showHelp = true
			return m, nil
		case "d":
			m.showEpic = true
			return m, nil
		case "f":
			m.showType = true
			m.typeIndex = 0
			return m, nil
		case "left", "h":
			m.moveCol(-1)
			return m, nil
		case "right", "l":
			m.moveCol(1)
			return m, nil
		case "up", "k":
			m.moveRow(-1)
			return m, nil
		case "down", "j":
			m.moveRow(1)
			return m, nil
		case "enter":
			if issue := m.currentIssue(); issue != nil {
				m.setInspectedIssue(*issue)
				m.showDetails = true
				return m, m.prepareDetailsForIssue(issue.ID)
			}
			return m, nil
		case "r":
			return m, m.loadCatalogCmd()
		}
	}

	return m, nil
}

// View renders the kanban layout, modals, and help footer.
func (m Model) View() string {
	if !m.ready {
		return "Loading kanban..."
	}
	if m.err != nil {
		return fmt.Sprintf("Failed to load kanban: %v", m.err)
	}
	if m.isUndersized() {
		return m.renderUndersized()
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderBoard(),
		m.renderFooter(),
	)

	if m.showPicker {
		return m.overlay(content, m.renderPicker())
	}
	if m.showHelp {
		return m.overlay(content, m.renderHelp())
	}
	if m.showType {
		return m.overlay(content, m.renderTypePicker())
	}
	if m.showEpic {
		return m.overlay(content, m.renderEpicDetails())
	}
	if m.showDetails {
		return m.overlay(content, m.renderDetails())
	}
	return content
}

func (m Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.showPicker = false
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.pickerIndex > 0 {
			m.pickerIndex--
		}
		return m, nil
	case "down", "j":
		if m.pickerIndex < len(m.catalog.Scopes)-1 {
			m.pickerIndex++
		}
		return m, nil
	case "enter":
		m.scopeIndex = m.pickerIndex
		m.showPicker = false
		m.selectedCol = 0
		m.selectedRow = 0
		m.scrollRow = 0
		m.ensureSelection()
		return m, nil
	case "a":
		m.selectScopeByKey(scopeAll)
		m.showPicker = false
		return m, nil
	case "tab":
		m.pickerIndex = (m.pickerIndex + 1) % len(m.catalog.Scopes)
		return m, nil
	}
	return m, nil
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "enter", "o":
		m.showHelp = false
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateTypePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.showType = false
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.typeIndex > 0 {
			m.typeIndex--
		}
		return m, nil
	case "down", "j":
		if m.typeIndex < len(typeFilterOptions)-1 {
			m.typeIndex++
		}
		return m, nil
	case "enter":
		m.typeFilter = typeFilterOptions[m.typeIndex]
		m.showType = false
		m.selectedCol = 0
		m.selectedRow = 0
		m.scrollRow = 0
		m.ensureSelection()
		return m, nil
	case "a":
		m.typeFilter = typeFilterOptions[0]
		m.showType = false
		m.selectedCol = 0
		m.selectedRow = 0
		m.scrollRow = 0
		m.ensureSelection()
		return m, nil
	}
	return m, nil
}

// applyWindowSize records usable dimensions and ignores transient invalid resize events.
func (m *Model) applyWindowSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	m.width = width
	m.height = height
	m.ready = true
}

// isUndersized reports whether the current window is too small for the full board layout.
func (m Model) isUndersized() bool {
	return m.width < minRenderableWidth || m.height < minRenderableHeight
}

func (m *Model) cycleScope(step int) {
	if len(m.catalog.Scopes) == 0 {
		return
	}
	m.scopeIndex = (m.scopeIndex + step + len(m.catalog.Scopes)) % len(m.catalog.Scopes)
	m.selectedCol = 0
	m.selectedRow = 0
	m.scrollRow = 0
	m.ensureSelection()
}

func (m *Model) selectScopeByKey(key string) {
	for i, scope := range m.catalog.Scopes {
		if scope.Key != key {
			continue
		}
		m.scopeIndex = i
		m.selectedCol = 0
		m.selectedRow = 0
		m.scrollRow = 0
		m.ensureSelection()
		return
	}
}

func (m *Model) moveCol(delta int) {
	next := m.selectedCol + delta
	if next < 0 || next > 2 {
		return
	}
	m.selectedCol = next
	m.ensureSelection()
}

func (m *Model) moveRow(delta int) {
	column := m.currentColumn()
	if len(column) == 0 {
		m.selectedRow = 0
		m.scrollRow = 0
		return
	}
	next := m.selectedRow + delta
	if next < 0 {
		next = 0
	}
	if next >= len(column) {
		next = len(column) - 1
	}
	m.selectedRow = next
	m.adjustScroll()
}

// navigateDetails moves within the current column and keeps the details modal open.
func (m *Model) navigateDetails(delta int) tea.Cmd {
	before := m.inspectedIssueRef()
	if before == nil {
		return nil
	}
	m.syncSelectionToInspectedIssue()
	beforeID := before.ID
	m.moveRow(delta)
	after := m.currentIssue()
	if after == nil || after.ID == beforeID {
		return nil
	}
	m.setInspectedIssue(*after)
	return m.prepareDetailsForIssue(after.ID)
}

// navigateDetailsCol moves to an adjacent modal column without cycling.
func (m *Model) navigateDetailsCol(delta int) tea.Cmd {
	before := m.inspectedIssueRef()
	if before == nil {
		return nil
	}
	m.syncSelectionToInspectedIssue()
	nextCol := m.selectedCol + delta
	if nextCol < 0 || nextCol > 2 {
		return nil
	}
	columns := m.currentColumns()
	switch nextCol {
	case 0:
		if len(columns.Todo) == 0 {
			return nil
		}
	case 1:
		if len(columns.Claimed) == 0 {
			return nil
		}
	case 2:
		if len(columns.Done) == 0 {
			return nil
		}
	}
	m.moveCol(delta)
	after := m.currentIssue()
	if after == nil || after.ID == before.ID {
		return nil
	}
	m.setInspectedIssue(*after)
	return m.prepareDetailsForIssue(after.ID)
}

func (m *Model) ensureSelection() {
	column := m.currentColumn()
	if len(column) == 0 {
		m.selectedRow = 0
		m.scrollRow = 0
		return
	}
	if m.selectedRow >= len(column) {
		m.selectedRow = len(column) - 1
	}
	if m.selectedRow < 0 {
		m.selectedRow = 0
	}
	m.adjustScroll()
}

func (m *Model) adjustScroll() {
	rowsVisible := m.visibleRows()
	if rowsVisible < 1 {
		m.scrollRow = 0
		return
	}
	if m.selectedRow < m.scrollRow {
		m.scrollRow = m.selectedRow
	}
	if m.selectedRow >= m.scrollRow+rowsVisible {
		m.scrollRow = m.selectedRow - rowsVisible + 1
	}
	if m.scrollRow < 0 {
		m.scrollRow = 0
	}
}

func (m Model) visibleRows() int {
	rows := (m.boardHeightBudget() - columnHeaderLines) / cardOuterHeight
	if rows < 1 {
		return 0
	}
	return rows
}

func (m Model) currentScope() Scope {
	if len(m.catalog.Scopes) == 0 || m.scopeIndex >= len(m.catalog.Scopes) {
		return Scope{Key: scopeAll, Title: "All Epics"}
	}
	return m.catalog.Scopes[m.scopeIndex]
}

func (m Model) currentColumns() scopeColumns {
	scope := m.currentScope()
	return m.applyTypeFilter(m.catalog.Columns[scope.Key])
}

func (m Model) applyTypeFilter(columns scopeColumns) scopeColumns {
	selected := strings.TrimSpace(strings.ToLower(m.typeFilter))
	if selected == "" || selected == typeFilterOptions[0] {
		return columns
	}
	return scopeColumns{
		Todo:    filterByType(columns.Todo, selected),
		Claimed: filterByType(columns.Claimed, selected),
		Done:    filterByType(columns.Done, selected),
	}
}

func (m Model) currentColumn() []model.Issue {
	columns := m.currentColumns()
	switch m.selectedCol {
	case 1:
		return columns.Claimed
	case 2:
		return columns.Done
	default:
		return columns.Todo
	}
}

func (m Model) currentIssue() *model.Issue {
	column := m.currentColumn()
	if len(column) == 0 || m.selectedRow < 0 || m.selectedRow >= len(column) {
		return nil
	}
	issue := column[m.selectedRow]
	return &issue
}

func (m Model) currentIssueDetails() issueDetails {
	issue := m.inspectedIssueRef()
	if issue == nil || m.details == nil {
		return issueDetails{}
	}
	return m.details[issue.ID]
}

// inspectedIssueRef returns the issue currently shown in the details modal.
func (m Model) inspectedIssueRef() *model.Issue {
	if m.inspectedIssueID == "" {
		return m.currentIssue()
	}
	if issue := m.findIssueByID(m.inspectedIssueID); issue != nil {
		return issue
	}
	if m.inspectedIssue.ID == m.inspectedIssueID {
		issue := m.inspectedIssue
		return &issue
	}
	return nil
}

// inspectedColumnIndex reports the active modal column for the inspected issue.
func (m Model) inspectedColumnIndex() int {
	if col, _, ok := m.locateIssueInCurrentScope(m.inspectedIssueID); ok {
		return col
	}
	issue := m.inspectedIssueRef()
	if issue == nil {
		return m.selectedCol
	}
	switch issue.Status {
	case "in_progress":
		return 1
	case "closed":
		return 2
	default:
		return 0
	}
}

// inspectedColumnTitle returns the board column title for the inspected issue.
func (m Model) inspectedColumnTitle() string {
	switch m.inspectedColumnIndex() {
	case 1:
		return "CLAIMED"
	case 2:
		return "DONE"
	default:
		return "TO DO"
	}
}

// locateIssueInCurrentScope returns the issue position in the currently visible board scope.
func (m Model) locateIssueInCurrentScope(issueID string) (int, int, bool) {
	if issueID == "" {
		return 0, 0, false
	}
	columns := m.currentColumns()
	if row, ok := findIssueIndex(columns.Todo, issueID); ok {
		return 0, row, true
	}
	if row, ok := findIssueIndex(columns.Claimed, issueID); ok {
		return 1, row, true
	}
	if row, ok := findIssueIndex(columns.Done, issueID); ok {
		return 2, row, true
	}
	return 0, 0, false
}

// findIssueByID locates an issue across the current catalog snapshot.
func (m Model) findIssueByID(issueID string) *model.Issue {
	for _, columns := range m.catalog.Columns {
		if issue := findIssueInColumn(columns.Todo, issueID); issue != nil {
			return issue
		}
		if issue := findIssueInColumn(columns.Claimed, issueID); issue != nil {
			return issue
		}
		if issue := findIssueInColumn(columns.Done, issueID); issue != nil {
			return issue
		}
	}
	return nil
}

func (m Model) loadCatalogCmd() tea.Cmd {
	return func() tea.Msg {
		catalog, err := LoadCatalog(m.svc)
		return catalogLoadedMsg{catalog: catalog, err: err}
	}
}

func (m Model) loadDetailsCmd(issueID string) tea.Cmd {
	return func() tea.Msg {
		dependencies, err := m.svc.Dependencies(issueID)
		if err != nil {
			return detailsLoadedMsg{issueID: issueID, err: err}
		}
		dependents, err := m.svc.Dependents(issueID)
		if err != nil {
			return detailsLoadedMsg{issueID: issueID, err: err}
		}
		return detailsLoadedMsg{
			issueID: issueID,
			details: issueDetails{
				Dependencies: dependencies,
				Dependents:   dependents,
			},
		}
	}
}

// prepareDetailsForIssue ensures detail state exists and triggers loading when needed.
func (m *Model) prepareDetailsForIssue(issueID string) tea.Cmd {
	if m.details == nil {
		m.details = make(map[string]issueDetails)
	}
	if details, ok := m.details[issueID]; ok {
		if details.Loading || details.Err != nil || details.Dependencies != nil || details.Dependents != nil {
			return nil
		}
	}
	m.details[issueID] = issueDetails{Loading: true}
	return m.loadDetailsCmd(issueID)
}

// setInspectedIssue anchors the open modal to one explicit issue.
func (m *Model) setInspectedIssue(issue model.Issue) {
	m.inspectedIssueID = issue.ID
	m.inspectedIssue = issue
}

// clearInspectedIssue resets the explicit modal target.
func (m *Model) clearInspectedIssue() {
	m.inspectedIssueID = ""
	m.inspectedIssue = model.Issue{}
}

// syncSelectionToInspectedIssue keeps board cursor state aligned with the modal target when possible.
func (m *Model) syncSelectionToInspectedIssue() {
	col, row, ok := m.locateIssueInCurrentScope(m.inspectedIssueID)
	if !ok {
		m.ensureSelection()
		return
	}
	m.selectedCol = col
	m.selectedRow = row
	m.adjustScroll()
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func (m Model) renderHeader() string {
	scope := m.currentScope()
	contentWidth := maxInt(1, m.width-2)
	headerStyle := lipgloss.NewStyle().
		Width(maxInt(1, m.width)).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("24"))

	subtitleStyle := lipgloss.NewStyle().
		Width(maxInt(1, m.width)).
		Padding(0, 1).
		Foreground(lipgloss.Color("245"))

	title := "faz kanban"
	scopeTitle := "Epic: " + scope.Title
	subtitle := fmt.Sprintf("Auto-refresh every 10s. Type: %s.", strings.ToUpper(m.typeFilter))
	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(truncateLine(fmt.Sprintf("%s  |  %s", title, scopeTitle), contentWidth)),
		subtitleStyle.Render(truncateLine(subtitle, contentWidth)),
	)
}

func (m Model) renderBoard() string {
	columns := m.currentColumns()
	if m.visibleRows() == 0 {
		return m.renderCompactBoardNotice()
	}
	widths, gap := boardColumnLayout(m.width)
	todo := m.renderColumn("TO DO", columns.Todo, 0, widths[0], lipgloss.Color("178"))
	claimed := m.renderColumn("CLAIMED", columns.Claimed, 1, widths[1], lipgloss.Color("39"))
	done := m.renderColumn("DONE", columns.Done, 2, widths[2], lipgloss.Color("71"))

	row := lipgloss.JoinHorizontal(lipgloss.Top, todo, strings.Repeat(" ", gap), claimed, strings.Repeat(" ", gap), done)
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, row)
}

func (m Model) renderColumn(title string, issues []model.Issue, colIndex, width int, accent lipgloss.Color) string {
	headerWidth := maxInt(1, width-2)
	header := lipgloss.NewStyle().
		Width(maxInt(1, width)).
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("230")).
		Background(accent).
		Render(truncateLine(fmt.Sprintf("%s (%d)", title, len(issues)), headerWidth))

	rowsVisible := m.visibleRows()
	start := 0
	if colIndex == m.selectedCol {
		start = m.scrollRow
	}
	if start > len(issues) {
		start = len(issues)
	}
	end := start + rowsVisible
	if end > len(issues) {
		end = len(issues)
	}

	body := make([]string, 0, rowsVisible)
	if len(issues) == 0 {
		empty := lipgloss.NewStyle().
			Width(maxInt(1, width-2)).
			Height(5).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 1).
			Foreground(lipgloss.Color("244")).
			Render("No tasks")
		body = append(body, lipgloss.PlaceHorizontal(width, lipgloss.Center, empty))
	} else {
		for rowIndex := start; rowIndex < end; rowIndex++ {
			body = append(body, m.renderCard(issues[rowIndex], colIndex == m.selectedCol && rowIndex == m.selectedRow, width))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, append([]string{header}, body...)...)
}

func (m Model) renderCard(issue model.Issue, selected bool, width int) string {
	borderColor := lipgloss.Color("238")
	bgColor := lipgloss.Color("235")
	titleColor := lipgloss.Color("252")
	if selected {
		borderColor = lipgloss.Color("246")
		bgColor = lipgloss.Color("240")
		titleColor = lipgloss.Color("230")
	}

	meta := fmt.Sprintf("P%d • %s", issue.Priority, issue.Type)
	if m.currentScope().Key == scopeAll {
		if issue.ParentID != nil {
			meta = fmt.Sprintf("%s • %s", meta, truncateLine(m.catalog.EpicTitles[*issue.ParentID], width-6))
		} else {
			meta = fmt.Sprintf("%s • No Epic", meta)
		}
	}
	title := fitLines(issue.Title, maxInt(1, width-4), 2)

	cardStyle := lipgloss.NewStyle().
		Width(maxInt(1, width-2)).
		Height(5).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(bgColor).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(selected)
	metaColor := lipgloss.Color("243")
	if selected {
		metaColor = lipgloss.Color("251")
	}
	metaStyle := lipgloss.NewStyle().Foreground(metaColor)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		metaStyle.Render(truncateLine(meta, maxInt(1, width-4))),
	)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, cardStyle.Render(content))
}

func (m Model) renderFooter() string {
	contentWidth := maxInt(1, m.width-2)
	lines, _ := m.footerLines(contentWidth)
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, truncateLine(line, contentWidth))
	}
	return lipgloss.NewStyle().
		Width(maxInt(1, m.width)).
		Padding(0, 1).
		Foreground(lipgloss.Color("244")).
		Render(strings.Join(rendered, "\n"))
}

// renderUndersized shows a stable fallback instead of rendering a clipped board.
func (m Model) renderUndersized() string {
	lines := []string{
		truncateLine("faz kanban", maxInt(1, m.width)),
		truncateLine(fmt.Sprintf("Window too small (%dx%d)", m.width, m.height), maxInt(1, m.width)),
		truncateLine(fmt.Sprintf("Need at least %dx%d", minRenderableWidth, minRenderableHeight), maxInt(1, m.width)),
		truncateLine("Enlarge the terminal.", maxInt(1, m.width)),
		truncateLine("q quit", maxInt(1, m.width)),
	}
	if len(lines) > m.height {
		lines = lines[:m.height]
	}
	return strings.Join(lines, "\n")
}

// renderCompactBoardNotice keeps the view height-safe when cards cannot fit.
func (m Model) renderCompactBoardNotice() string {
	return lipgloss.NewStyle().
		Width(maxInt(1, m.width)).
		Foreground(lipgloss.Color("244")).
		Render(truncateLine("Window height too small for kanban cards.", maxInt(1, m.width)))
}

func (m Model) renderPicker() string {
	width := minInt(60, maxInt(minModalWidth, m.width-10))
	lines := []string{"Select scope", ""}
	for i, scope := range m.catalog.Scopes {
		prefix := "  "
		if i == m.pickerIndex {
			prefix = "> "
		}
		lines = append(lines, prefix+scope.Title)
	}
	lines = append(lines, "", "Enter select • Esc close • a all")
	box := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
	return box
}

func (m Model) renderHelp() string {
	width := minInt(72, maxInt(minModalWidth, m.width-10))
	lines := []string{
		"Keybindings",
		"",
		"Board:",
		"  Tab / Shift+Tab  cycle epic scope",
		"  e                open epic list",
		"  a                all epics view",
		"  f                issue type filter",
		"  d                epic details",
		"  arrows / h j k l move selection",
		"  Enter            open task details",
		"  r                refresh now",
		"  o                keybinding help",
		"  q                quit",
		"",
		"Task details modal:",
		"  up/down          move within current column",
		"  left/right       move to adjacent column",
		"  Enter / Esc / q  close details",
		"",
		"Epic picker:",
		"  up/down          move",
		"  Enter            select",
		"  a                all epics view",
		"  Esc / q          close picker",
		"",
		"Enter, Esc, q, or o closes this view.",
	}
	box := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
	return box
}

func (m Model) renderTypePicker() string {
	width := minInt(44, maxInt(minModalWidth, m.width-10))
	lines := []string{"Select issue type", ""}
	for i, issueType := range typeFilterOptions {
		prefix := "  "
		if i == m.typeIndex {
			prefix = "> "
		}
		lines = append(lines, prefix+strings.ToUpper(issueType))
	}
	lines = append(lines, "", "Enter select • Esc close • a all")
	box := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
	return box
}

func (m Model) renderDetails() string {
	issue := m.inspectedIssueRef()
	if issue == nil {
		return ""
	}
	width := minInt(84, maxInt(minModalWidth, m.width-12))
	details := m.currentIssueDetails()
	columnLabel := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("60")).
		Render(m.inspectedColumnTitle())
	parentTitle := "None"
	if issue.ParentID != nil {
		parentTitle = m.catalog.EpicTitles[*issue.ParentID]
	}
	lines := []string{
		issue.Title,
		"",
		fmt.Sprintf("ID: %s", issue.ID),
		fmt.Sprintf("Type: %s", issue.Type),
		fmt.Sprintf("Priority: P%d", issue.Priority),
		fmt.Sprintf("Status: %s", issue.Status),
		fmt.Sprintf("Epic: %s", parentTitle),
		"",
		issue.Description,
		"",
	}
	switch {
	case details.Loading:
		lines = append(lines, "Loading dependencies...")
	case details.Err != nil:
		lines = append(lines, fmt.Sprintf("Failed to load dependencies: %v", details.Err))
	default:
		lines = append(lines, m.renderIssueLinks("Blocked by", details.Dependencies, width)...)
		lines = append(lines, "")
		lines = append(lines, m.renderIssueLinks("Blocks", details.Dependents, width)...)
	}
	lines = append(lines,
		"",
		"Enter or Esc closes this view.",
	)
	box := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
	return lipgloss.JoinVertical(lipgloss.Center, columnLabel, "", box)
}

func (m Model) renderEpicDetails() string {
	width := minInt(84, maxInt(minModalWidth, m.width-12))
	epic, reason := m.resolveInspectedEpic()
	lines := []string{"Epic Details", ""}
	if epic == nil {
		lines = append(lines,
			reason,
			"",
			"Tip: select an epic scope with Tab or e,",
			"or highlight a task linked to an epic in All Epics.",
		)
	} else {
		lines = append(lines,
			epic.Title,
			"",
			fmt.Sprintf("ID: %s", epic.ID),
			fmt.Sprintf("Type: %s", epic.Type),
			fmt.Sprintf("Priority: P%d", epic.Priority),
			fmt.Sprintf("Status: %s", epic.Status),
			"",
			epic.Description,
		)
	}
	lines = append(lines, "", "Enter or Esc closes this view.")
	box := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(1, 2).
		Background(lipgloss.Color("24")).
		Foreground(lipgloss.Color("230")).
		Render(strings.Join(lines, "\n"))
	return box
}

// resolveInspectedEpic chooses which epic to render from the current board context.
func (m Model) resolveInspectedEpic() (*model.Issue, string) {
	scope := m.currentScope()
	switch scope.Key {
	case scopeNoEpic:
		return nil, "No epic is associated with this scope."
	case scopeAll:
		issue := m.currentIssue()
		if issue == nil {
			return nil, "No task is selected."
		}
		if issue.ParentID == nil {
			return nil, "Selected task is not linked to an epic."
		}
		epic, ok := m.catalog.Epics[*issue.ParentID]
		if !ok {
			return nil, "Epic details are unavailable for the selected task."
		}
		return &epic, ""
	default:
		epic, ok := m.catalog.Epics[scope.Key]
		if !ok {
			return nil, "Selected scope does not map to a known epic."
		}
		return &epic, ""
	}
}

func findIssueInColumn(issues []model.Issue, issueID string) *model.Issue {
	for _, issue := range issues {
		if issue.ID == issueID {
			found := issue
			return &found
		}
	}
	return nil
}

// findIssueIndex returns the row index for an issue within one kanban column.
func findIssueIndex(issues []model.Issue, issueID string) (int, bool) {
	for i, issue := range issues {
		if issue.ID == issueID {
			return i, true
		}
	}
	return 0, false
}

func (m Model) overlay(base, modal string) string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) renderIssueLinks(title string, issues []model.Issue, width int) []string {
	lines := []string{title + ":"}
	if len(issues) == 0 {
		return append(lines, "  None")
	}
	for _, issue := range issues {
		lines = append(lines, "  - "+truncateLine(fmt.Sprintf("%s %s", issue.ID, issue.Title), width-8))
	}
	return lines
}

// boardHeightBudget returns the lines available for the board after header and footer.
func (m Model) boardHeightBudget() int {
	budget := m.height - headerLines - m.footerLineCount()
	if budget < columnHeaderLines {
		return columnHeaderLines
	}
	return budget
}

// footerLineCount returns how many terminal rows the footer currently needs.
func (m Model) footerLineCount() int {
	contentWidth := maxInt(1, m.width-2)
	lines, _ := m.footerLines(contentWidth)
	if len(lines) < 1 {
		return 1
	}
	return len(lines)
}

func boardColumnLayout(totalWidth int) ([3]int, int) {
	gap := 2
	if totalWidth < 72 {
		gap = 1
	}
	available := totalWidth - 2*gap
	if available < 3 {
		available = 3
	}
	base := available / 3
	remainder := available % 3
	widths := [3]int{base, base, base}
	for i := 0; i < remainder; i++ {
		widths[i]++
	}
	for i := range widths {
		widths[i] = maxInt(1, widths[i])
	}
	return widths, gap
}

// footerLines builds a responsive footer and reports whether condensed mode is active.
func (m Model) footerLines(width int) ([]string, bool) {
	fullItems := []string{
		"Tab next epic",
		"Shift+Tab previous",
		"e epic list",
		"a all",
		"f type filter",
		"d epic details",
		"arrows move",
		"Enter details",
		"r refresh",
		"q quit",
	}
	full, ok := wrapFooterItems(fullItems, width, 2)
	if ok {
		return full, false
	}
	condensedItems := []string{
		"Tab",
		"e",
		"a",
		"f",
		"d",
		"arrows",
		"Enter " + string(rune(0x23CE)),
		"q",
		"o all keys",
	}
	condensed, ok := wrapFooterItems(condensedItems, width, 2)
	if ok {
		return condensed, true
	}
	// At extreme widths keep core controls visible in one safe line.
	return []string{truncateLine("Tab • e • a • q • o", width)}, true
}

// filterByType keeps only issues matching one issue type.
func filterByType(issues []model.Issue, selected string) []model.Issue {
	filtered := make([]model.Issue, 0, len(issues))
	for _, issue := range issues {
		if strings.EqualFold(issue.Type, selected) {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// wrapFooterItems packs footer items into at most maxLines lines.
func wrapFooterItems(items []string, width, maxLines int) ([]string, bool) {
	if width < 1 || len(items) == 0 || maxLines < 1 {
		return nil, false
	}
	lines := make([]string, 0, maxLines)
	current := ""
	for _, item := range items {
		if lipgloss.Width(item) > width {
			return nil, false
		}
		next := item
		if current != "" {
			next = current + footerGap + item
		}
		if lipgloss.Width(next) <= width {
			current = next
			continue
		}
		lines = append(lines, current)
		if len(lines) >= maxLines {
			return nil, false
		}
		current = item
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) > maxLines {
		return nil, false
	}
	return lines, true
}

func fitLines(text string, width, maxLines int) string {
	if width < 1 {
		width = 1
	}
	lines := lipgloss.NewStyle().Width(width).MaxWidth(width).Render(text)
	parts := strings.Split(lines, "\n")
	if len(parts) <= maxLines {
		return strings.Join(parts, "\n")
	}
	parts = parts[:maxLines]
	last := truncateLine(parts[len(parts)-1], width)
	parts[len(parts)-1] = strings.TrimRight(last, " ") + "…"
	return strings.Join(parts, "\n")
}

func truncateLine(text string, width int) string {
	if width <= 1 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	return string(runes[:width-1]) + "…"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
