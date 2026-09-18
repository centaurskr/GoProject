// ui.go — bubbletea Model / Update / View 구현
//
// goTbmon의 전체 TUI 로직을 담는다.
//
// 화면 구조:
//   modeList    — 프로세스 목록 (기본 화면)
//   modeDetail  — 프로세스 상세 정보 (h 키)
//   modeTree    — 그룹 트리 (g 키)
//   modeDeps    — 전체 의존성 트리 (d 키)
//   modeEnv     — 전역 환경변수 목록 (e 키)
//   modeHelp    — 도움말 (? 키)
//   modeConfirm — y/N 확인 다이얼로그
//   modeInput   — / 검색 입력
//
// 업데이트 흐름:
//   Init() → doFetch() → statusMsg → Update() → View()
//   키 입력 → handleKey() → 명령 Cmd → cmdResultMsg / envMsg → Update()
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const defaultWidth = 80

// ── 메시지 타입 ──────────────────────────────────────────────────────────────

type statusMsg struct {
	procs []ProcessSnapshot
	err   error
}

type cmdResultMsg struct {
	text  string
	isErr bool
}

type multiResultMsg struct {
	results []cmdResult
}

type envMsg struct {
	vars map[string]string
	err  error
}

type cmdResult struct {
	id  string
	ok  bool
	msg string
}

// ── 화면 모드 ────────────────────────────────────────────────────────────────

type viewMode int

const (
	modeList    viewMode = iota
	modeConfirm          // y/N 확인
	modeInput            // / 검색
	modeHelp             // ? 도움말
	modeDetail           // h 프로세스 상세
	modeTree             // g 그룹 트리
	modeEnv              // e 전역 환경변수
	modeDeps             // d 전체 의존성 트리
)

type pendingAction struct {
	label    string
	ids      []string
	fn       func(id string) (ControlResponse, error)
	fnSingle func() (ControlResponse, error)
}

// ── 모델 ─────────────────────────────────────────────────────────────────────

type Model struct {
	client       *Client
	host         string            // 접속 호스트 (타이틀 표시용)
	procs        []ProcessSnapshot
	cursor       int // filtered 기준 절대 인덱스
	scrollOffset int // 뷰포트 첫 행의 filtered 인덱스
	sel          map[string]bool
	width        int
	height       int
	mode         viewMode
	pending      pendingAction
	statusLine   string
	statusErr    bool
	lastUpdate   time.Time
	filter       string
	input        textinput.Model
	loading      bool
	detailProc   *ProcessSnapshot   // h 키로 열람 중인 프로세스
	detailScroll int               // 상세 화면 스크롤 오프셋
	helpScroll   int               // 도움말 스크롤 오프셋
	treeScroll   int               // 그룹 트리 스크롤 오프셋
	envVars      map[string]string // e 키로 조회한 전역 환경변수
	envScroll    int               // 환경변수 화면 스크롤 오프셋
	depsScroll   int               // 의존성 트리 화면 스크롤 오프셋
}

// NewModel: 초기 Model을 생성한다. host는 타이틀 표시용, addr은 TCP 접속 주소.
func NewModel(host, addr string) Model {
	ti := textinput.New()
	ti.Placeholder = "검색..."
	ti.CharLimit = 64

	return Model{
		client: NewClient(addr),
		host:   host,
		input:  ti,
		width:  defaultWidth,
		sel:    make(map[string]bool),
	}
}

// ── Init: 시작 시 1회 목록 로드 ───────────────────────────────────────────────

// Init: 프로그램 시작 시 1회 프로세스 목록 로드 명령을 반환한다.
func (m Model) Init() tea.Cmd {
	return doFetch(m.client)
}

// doFetch: goTbnova에 status 명령을 보내고 결과를 statusMsg로 반환하는 Cmd를 만든다.
func doFetch(c *Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.Status("")
		return statusMsg{procs: resp.Processes, err: err}
	}
}

// ── 뷰포트 계산 ───────────────────────────────────────────────────────────────

// 고정 오버헤드:
//   타이틀(1) + 통계(1) + 헤더+구분선(2) + 하단구분선(1) + 상태줄(1) + 힌트(1) = 7
//   검색바 활성 시 +1

const fixedRows = 7

// viewportSize: 목록 화면에서 프로세스 행을 표시할 수 있는 줄 수를 반환한다.
func (m Model) viewportSize() int {
	extra := 0
	if m.mode == modeInput || m.filter != "" {
		extra = 1
	}
	vs := m.height - fixedRows - extra
	if vs < 3 {
		vs = 3
	}
	return vs
}

// 커서가 뷰포트 안에 들어오도록 scrollOffset 조정
func (m *Model) adjustScroll() {
	vs := m.viewportSize()
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+vs {
		m.scrollOffset = m.cursor - vs + 1
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// clampCursor: 커서를 [0, n) 범위로 제한하고 scrollOffset을 조정한다.
func (m *Model) clampCursor(n int) {
	if n == 0 {
		m.cursor = 0
		m.scrollOffset = 0
		return
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.adjustScroll()
}

// ── Update ───────────────────────────────────────────────────────────────────

// Update: bubbletea 메시지 디스패처. 비동기 결과 메시지와 키 입력을 처리한다.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustScroll()
		return m, nil

	case statusMsg:
		m.loading = false
		if msg.err != nil {
			m.setStatus("접속 오류: "+msg.err.Error(), true)
		} else {
			// 사라진 ID 선택 해제
			exists := make(map[string]bool, len(msg.procs))
			for _, p := range msg.procs {
				exists[p.ID] = true
			}
			for id := range m.sel {
				if !exists[id] {
					delete(m.sel, id)
				}
			}
			m.procs = msg.procs
			m.lastUpdate = time.Now()
			m.clampCursor(len(m.filtered()))
			m.setStatus(fmt.Sprintf("목록 갱신 완료 (%d개)", len(msg.procs)), false)
		}
		return m, nil

	case cmdResultMsg:
		m.loading = false
		m.setStatus(msg.text, msg.isErr)
		// 명령 완료 후 자동으로 목록 재요청
		return m, doFetch(m.client)

	case multiResultMsg:
		m.loading = false
		m.sel = make(map[string]bool)
		ok, fail := 0, 0
		var errs []string
		for _, r := range msg.results {
			if r.ok {
				ok++
			} else {
				fail++
				errs = append(errs, r.id+": "+r.msg)
			}
		}
		if fail == 0 {
			m.setStatus(fmt.Sprintf("완료 %d건 — 목록 갱신 중...", ok), false)
		} else {
			m.setStatus(fmt.Sprintf("완료 %d건  실패 %d건: %s", ok, fail, strings.Join(errs, ", ")), true)
		}
		// 명령 완료 후 자동으로 목록 재요청
		return m, doFetch(m.client)

	case envMsg:
		m.loading = false
		if msg.err != nil {
			m.setStatus("env 조회 오류: "+msg.err.Error(), true)
		} else {
			m.envVars = msg.vars
			m.envScroll = 0
			m.mode = modeEnv
			m.setStatus(fmt.Sprintf("전역 환경변수 %d개 조회 완료", len(msg.vars)), false)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.mode == modeInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.filter = m.input.Value()
		m.clampCursor(len(m.filtered()))
		return m, cmd
	}

	return m, nil
}

// ── Key Handler ──────────────────────────────────────────────────────────────

// handleKey: 현재 viewMode에 따라 키 입력을 라우팅한다.
// 서브 화면(detail/tree/env/deps/help)은 ↑↓ 스크롤 외의 키를 목록 복귀로 처리한다.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 도움말 모드: ↑↓ 스크롤, 그 외 아무 키나 누르면 복귀
	if m.mode == modeHelp {
		switch msg.String() {
		case "up", "k":
			if m.helpScroll > 0 {
				m.helpScroll--
			}
		case "down", "j":
			m.helpScroll++
		default:
			m.mode = modeList
			m.helpScroll = 0
			m.setStatus("", false)
		}
		return m, nil
	}

	// 상세 모드: ↑↓ 스크롤, 그 외 아무 키나 누르면 복귀
	if m.mode == modeDetail {
		switch msg.String() {
		case "up", "k":
			if m.detailScroll > 0 {
				m.detailScroll--
			}
		case "down", "j":
			m.detailScroll++
		default:
			m.mode = modeList
			m.detailProc = nil
			m.detailScroll = 0
			m.setStatus("", false)
		}
		return m, nil
	}

	// 트리 모드: ↑↓ 스크롤, 그 외 아무 키나 누르면 복귀
	if m.mode == modeTree {
		switch msg.String() {
		case "up", "k":
			if m.treeScroll > 0 {
				m.treeScroll--
			}
		case "down", "j":
			m.treeScroll++
		default:
			m.mode = modeList
			m.treeScroll = 0
			m.setStatus("", false)
		}
		return m, nil
	}

	// 환경변수 모드: ↑↓ 스크롤, 그 외 아무 키나 누르면 복귀
	if m.mode == modeEnv {
		switch msg.String() {
		case "up", "k":
			if m.envScroll > 0 {
				m.envScroll--
			}
		case "down", "j":
			m.envScroll++
		default:
			m.mode = modeList
			m.envScroll = 0
			m.setStatus("", false)
		}
		return m, nil
	}

	// 의존성 트리 모드: ↑↓ 스크롤, 그 외 아무 키나 누르면 복귀
	if m.mode == modeDeps {
		switch msg.String() {
		case "up", "k":
			if m.depsScroll > 0 {
				m.depsScroll--
			}
		case "down", "j":
			m.depsScroll++
		default:
			m.mode = modeList
			m.depsScroll = 0
			m.setStatus("", false)
		}
		return m, nil
	}

	// 확인 모드
	if m.mode == modeConfirm {
		switch msg.String() {
		case "y", "Y", "enter":
			m.mode = modeList
			m.loading = true
			p := m.pending
			if p.fnSingle != nil {
				return m, runSingleCmd(p.fnSingle)
			}
			return m, runMultiCmd(p.ids, p.fn)
		case "n", "N", "esc", "q":
			m.mode = modeList
			m.setStatus("취소됨", false)
		}
		return m, nil
	}

	// 검색 입력 모드
	if m.mode == modeInput {
		switch msg.String() {
		case "esc":
			m.mode = modeList
			m.filter = ""
			m.input.SetValue("")
			m.input.Blur()
		case "enter":
			m.mode = modeList
			m.input.Blur()
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.filter = m.input.Value()
			m.clampCursor(len(m.filtered()))
			return m, cmd
		}
		return m, nil
	}

	// 목록 모드
	filtered := m.filtered()
	n := len(filtered)
	vs := m.viewportSize()

	// PgUp / PgDn: Type 비교로 처리 (터미널별 문자열 차이 무관)
	if msg.Type == tea.KeyPgUp {
		m.cursor -= vs
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.adjustScroll()
		return m, nil
	}
	if msg.Type == tea.KeyPgDown {
		m.cursor += vs
		if m.cursor >= n {
			if n > 0 {
				m.cursor = n - 1
			} else {
				m.cursor = 0
			}
		}
		m.adjustScroll()
		return m, nil
	}

	switch msg.String() {

	case "q", "ctrl+c":
		return m, tea.Quit

	// ── 커서 이동 (scrollOffset 자동 조정) ───────────────────────────────
	case "up":
		if m.cursor > 0 {
			m.cursor--
			m.adjustScroll()
		}

	case "down":
		if m.cursor < n-1 {
			m.cursor++
			m.adjustScroll()
		}

	case "ctrl+b":
		m.cursor -= vs
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.adjustScroll()

	case "ctrl+f":
		m.cursor += vs
		if m.cursor >= n {
			if n > 0 {
				m.cursor = n - 1
			} else {
				m.cursor = 0
			}
		}
		m.adjustScroll()

	case "home":
		m.cursor = 0
		m.scrollOffset = 0

	case "end":
		if n > 0 {
			m.cursor = n - 1
			m.adjustScroll()
		}

	// ── 검색 ─────────────────────────────────────────────────────────────
	case "/":
		m.mode = modeInput
		m.input.Focus()

	// ── g: 그룹 트리 화면 ────────────────────────────────────────────────
	case "g":
		m.treeScroll = 0
		m.mode = modeTree

	// ── d: 전체 의존성 트리 화면 ─────────────────────────────────────────
	case "d":
		m.depsScroll = 0
		m.mode = modeDeps

	// ── e: 전역 환경변수 조회 (goTbnova 접속 → 수신 → 단절) ──────────────
	case "e":
		m.loading = true
		m.setStatus("전역 환경변수 조회 중...", false)
		return m, doEnvFetch(m.client)

	// ── 도움말 ───────────────────────────────────────────────────────────
	case "?":
		m.mode = modeHelp

	// ── t: goTbnova에 Process Table Reload 명령 전송 (확인 후 실행) ───────
	case "t":
		m.pending = pendingAction{
			label:    "goTbnova Process Table Reload",
			fnSingle: func() (ControlResponse, error) { return m.client.Reload() },
		}
		m.mode = modeConfirm

	// ── f: 프로세스 목록 재요청 ───────────────────────────────────────────
	case "f":
		m.loading = true
		m.setStatus("목록 갱신 중...", false)
		return m, doFetch(m.client)

	// ── 선택 ─────────────────────────────────────────────────────────────
	case " ": // 개별 선택 토글 + 커서 아래 이동
		if p := m.cursorProc(filtered); p != nil {
			if m.sel[p.ID] {
				delete(m.sel, p.ID)
			} else {
				m.sel[p.ID] = true
			}
		}
		if m.cursor < n-1 {
			m.cursor++
			m.adjustScroll()
		}

	case "a": // 전체 선택
		for _, p := range filtered {
			m.sel[p.ID] = true
		}

	case "u": // 모든 선택 해제
		m.sel = make(map[string]bool)

	// ── s: 선택 프로세스 Start ────────────────────────────────────────────
	case "s":
		ids := m.targets(filtered)
		if len(ids) == 0 {
			m.setStatus("선택된 프로세스가 없습니다 (Space 키로 선택)", true)
			break
		}
		m.confirmMulti(m.actionLabel("Start", ids), ids, func(id string) (ControlResponse, error) {
			return m.client.Start(id)
		})

	// ── k: 선택 프로세스 Kill ─────────────────────────────────────────────
	case "k":
		ids := m.targets(filtered)
		if len(ids) == 0 {
			m.setStatus("선택된 프로세스가 없습니다 (Space 키로 선택)", true)
			break
		}
		m.confirmMulti(m.actionLabel("Kill", ids), ids, func(id string) (ControlResponse, error) {
			return m.client.Stop(id)
		})

	// ── h: 커서 위치 프로세스 상세 정보 ──────────────────────────────────
	case "h":
		if p := m.cursorProc(filtered); p != nil {
			cp := *p
			m.detailProc = &cp
			m.mode = modeDetail
		}
	}

	return m, nil
}

// ── 헬퍼 ─────────────────────────────────────────────────────────────────────

// targets: 현재 필터 목록에서 선택(sel)된 프로세스 ID 슬라이스를 반환한다.
func (m *Model) targets(filtered []ProcessSnapshot) []string {
	var ids []string
	for _, p := range filtered {
		if m.sel[p.ID] {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// actionLabel: 확인 다이얼로그에 표시할 동작 설명 문자열을 만든다.
func (m *Model) actionLabel(action string, ids []string) string {
	if len(ids) == 1 {
		return fmt.Sprintf("'%s' %s", ids[0], action)
	}
	preview := strings.Join(ids, ", ")
	if len([]rune(preview)) > 50 {
		preview = string([]rune(preview)[:48]) + ".."
	}
	return fmt.Sprintf("%s (%d건): %s", action, len(ids), preview)
}

// confirmMulti: 다수 프로세스 대상 명령을 modeConfirm 상태로 등록한다.
func (m *Model) confirmMulti(label string, ids []string, fn func(string) (ControlResponse, error)) {
	m.pending = pendingAction{label: label, ids: ids, fn: fn}
	m.mode = modeConfirm
}

func (m *Model) setStatus(text string, isErr bool) {
	m.statusLine = text
	m.statusErr = isErr
}

// cursorProc: 현재 커서 위치의 ProcessSnapshot 포인터를 반환한다. 범위 초과 시 nil.
func (m Model) cursorProc(filtered []ProcessSnapshot) *ProcessSnapshot {
	if len(filtered) == 0 || m.cursor >= len(filtered) {
		return nil
	}
	p := filtered[m.cursor]
	return &p
}

// filtered: m.filter 문자열로 ID·Group·Desc를 대소문자 무시 부분 검색한 결과를 반환한다.
func (m Model) filtered() []ProcessSnapshot {
	if m.filter == "" {
		return m.procs
	}
	f := strings.ToUpper(m.filter)
	var out []ProcessSnapshot
	for _, p := range m.procs {
		if strings.Contains(strings.ToUpper(p.ID), f) ||
			strings.Contains(strings.ToUpper(p.Group), f) ||
			strings.Contains(strings.ToUpper(p.Desc), f) {
			out = append(out, p)
		}
	}
	return out
}

// ── 명령 실행 ────────────────────────────────────────────────────────────────

// runSingleCmd: 인수 없는 단일 제어 함수를 비동기로 실행하고 결과를 cmdResultMsg로 반환한다.
func runSingleCmd(fn func() (ControlResponse, error)) tea.Cmd {
	return func() tea.Msg {
		resp, err := fn()
		if err != nil {
			return cmdResultMsg{text: "오류: " + err.Error(), isErr: true}
		}
		if resp.Result == "error" {
			return cmdResultMsg{text: "실패: " + resp.Message, isErr: true}
		}
		return cmdResultMsg{text: "완료", isErr: false}
	}
}

// runMultiCmd: ids 목록의 각 프로세스에 fn을 순차 실행하고 전체 결과를 multiResultMsg로 반환한다.
func runMultiCmd(ids []string, fn func(string) (ControlResponse, error)) tea.Cmd {
	return func() tea.Msg {
		results := make([]cmdResult, 0, len(ids))
		for _, id := range ids {
			resp, err := fn(id)
			if err != nil {
				results = append(results, cmdResult{id: id, ok: false, msg: err.Error()})
				continue
			}
			if resp.Result == "error" {
				results = append(results, cmdResult{id: id, ok: false, msg: resp.Message})
			} else {
				results = append(results, cmdResult{id: id, ok: true})
			}
		}
		return multiResultMsg{results: results}
	}
}

// ── View ─────────────────────────────────────────────────────────────────────

// View: 현재 모드에 맞는 화면 문자열을 반환한다. bubbletea가 매 업데이트마다 호출한다.
func (m Model) View() string {
	w := m.width
	if w < defaultWidth {
		w = defaultWidth
	}

	if m.mode == modeHelp {
		return m.renderHelpScreen(w)
	}
	if m.mode == modeDetail && m.detailProc != nil {
		return m.renderDetailScreen(w)
	}
	if m.mode == modeTree {
		return m.renderTreeScreen(w)
	}
	if m.mode == modeEnv {
		return m.renderEnvScreen(w)
	}
	if m.mode == modeDeps {
		return m.renderDepsScreen(w)
	}

	filtered := m.filtered()
	total := len(filtered)
	vs := m.viewportSize()

	// 페이지 정보: scrollOffset 기준
	totalPages := 1
	curPage := 1
	if vs > 0 && total > 0 {
		totalPages = (total + vs - 1) / vs
		curPage = m.scrollOffset/vs + 1
	}

	var b strings.Builder

	// ── 고정 상단 3줄 ──────────────────────────────────────────────────────

	// 1) 타이틀 바
	b.WriteString(m.renderTitle(w))

	// 2) 통계 바 (총 수, 상태별, 페이지)
	b.WriteString(m.renderStatBar(filtered, curPage, totalPages, vs, w))

	// 3) 검색바 (활성 시만 — 이 줄은 고정 3줄에 추가)
	if m.mode == modeInput || m.filter != "" {
		if m.mode == modeInput {
			b.WriteString(fg256(226) + " 검색: " + reset + m.input.View() + "\n")
		} else {
			b.WriteString(fg256(226) + " 검색: " + reset + fg256(15) + m.filter + reset + "\n")
		}
	}

	// 4) 컬럼 헤더 + 구분선
	b.WriteString(m.renderHeader(w))
	b.WriteString(fg256(240) + hline(w) + reset + "\n")

	// ── 스크롤 영역: 뷰포트 크기만큼 프로세스 행 ──────────────────────────
	b.WriteString(m.renderRows(filtered, vs, w))

	// ── 고정 하단 ──────────────────────────────────────────────────────────
	b.WriteString(fg256(240) + hline(w) + reset + "\n")
	b.WriteString(m.renderStatusLine(w))
	if m.mode == modeConfirm {
		b.WriteString(m.renderConfirm(w))
	}
	b.WriteString(m.renderHint(w))

	return b.String()
}

// ── 타이틀 바 ────────────────────────────────────────────────────────────────

// renderTitle: 타이틀 바(1줄) — 모니터명, 갱신시각, 선택수.
func (m Model) renderTitle(w int) string {
	upd := "미접속"
	if !m.lastUpdate.IsZero() {
		upd = m.lastUpdate.Format("2006-01-02 15:04:05")
	}
	selPart := ""
	if n := len(m.sel); n > 0 {
		selPart = fmt.Sprintf("  [선택:%d]", n)
	}
	text := fmt.Sprintf(" goTbnova Monitor  최종: %s%s", upd, selPart)
	return bold + bg256(20) + fgBrightWhite + padRight(text, w) + reset + "\n"
}

// ── 통계 바 ──────────────────────────────────────────────────────────────────

// renderStatBar: 통계 바(1줄) — 총 수, 상태별 카운트, 페이지 정보를 색상으로 표시한다.
// 비활성(DIS) 상태는 목록 행 색상으로는 처리되지만 카운터는 표시하지 않는다.
func (m Model) renderStatBar(filtered []ProcessSnapshot, curPage, totalPages, vs, w int) string {
	total := len(m.procs)
	shown := len(filtered)

	run, stop, back, wait := 0, 0, 0, 0
	for _, p := range filtered {
		switch p.State {
		case 2:
			run++
		case 3:
			stop++
		case 4:
			back++
		default:
			wait++
		}
	}

	cntPart := fmt.Sprintf(" [%s] 총 %d개", m.host, total)
	if m.filter != "" {
		cntPart = fmt.Sprintf(" [%s] 총 %d개(필터:%d)", m.host, total, shown)
	}

	statePart := fmt.Sprintf(
		"  %s실행:%d%s %s대기:%d%s %s종료중:%d%s %s백오프:%d%s",
		fg256(46), run, fg256(250),
		fg256(244), wait, fg256(250),
		fg256(196), stop, fg256(250),
		fg256(226), back, fg256(250),
	)

	pagePart := ""
	if totalPages > 1 {
		pagePart = fmt.Sprintf("  %sPG%d/%d%s",
			fg256(39), curPage, totalPages, reset)
	}

	return fg256(250) + bold + cntPart + reset +
		fg256(250) + statePart + reset + pagePart + "\n"
}

// ── 컬럼 헤더 ────────────────────────────────────────────────────────────────

const markerWidth = 2

type colLayout struct {
	id    int
	group int
	state int // 단축 STATE + 이유
	pid   int
	cnt   int
	start int // mm/dd-hh:mm:ss
	desc  int
}

// calcCols: 터미널 너비 w에서 각 컬럼의 문자 폭을 계산한다. desc 컬럼이 나머지를 가져간다.
func calcCols(w int) colLayout {
	// state: "BK:Abnormal" 최대 11자 → 12
	// start: "04/28-15:30:45" = 14자
	id, group, state, pid, cnt, start := 13, 10, 12, 6, 4, 14
	// spaces: id~group~state~pid~cnt~start~desc = 6개
	fixed := markerWidth + id + group + state + pid + cnt + start + 6
	desc := w - fixed
	if desc < 8 {
		desc = 8
	}
	return colLayout{id, group, state, pid, cnt, start, desc}
}

// renderHeader: 컬럼 헤더 행을 반환한다.
func (m Model) renderHeader(w int) string {
	c := calcCols(w)
	line := fmt.Sprintf("%-*s%-*s %-*s %-*s %*s %*s %-*s %-*s",
		markerWidth, "",
		c.id, "ID",
		c.group, "GROUP",
		c.state, "STATE",
		c.pid, "PID",
		c.cnt, "CNT",
		c.start, "START",
		c.desc, "DESC",
	)
	return bold + fg256(214) + line + reset + "\n"
}

// STATE 열 표시: 정지 이유가 있으면 "STATE:이유" 형태로 합친다
func stateField(state, stopReason int) string {
	base := stateName(state)
	if stopReason == 0 {
		return base
	}
	return base + ":" + stopReasonShort(stopReason)
}

// 정지 이유 단축 표기
func stopReasonShort(r int) string {
	switch r {
	case -1:
		return "Op"
	case -2:
		return "Holyday"
	case -3:
		return "Restart"
	case -4:
		return "Timeup"
	case -5:
		return "Reload"
	case -6:
		return "Abnormal"
	case -7:
		return "MaxCnt"
	default:
		return "?"
	}
}

// ── 프로세스 행 렌더링 ────────────────────────────────────────────────────────

// renderRows: 스크롤 뷰포트 범위의 프로세스 행을 렌더링한다. 빈 행으로 고정 높이를 유지한다.
func (m Model) renderRows(filtered []ProcessSnapshot, vs, w int) string {
	c := calcCols(w)
	var b strings.Builder
	n := len(filtered)

	if n == 0 {
		b.WriteString(fg256(244) + "  (프로세스 없음 — f키로 목록 로드)" + reset + "\n")
		return b.String()
	}

	end := m.scrollOffset + vs
	if end > n {
		end = n
	}

	for i := m.scrollOffset; i < end; i++ {
		p := filtered[i]
		isCursor := (i == m.cursor)
		isSel := m.sel[p.ID]

		pid := "-"
		if p.Pid > 0 {
			pid = fmt.Sprintf("%d", p.Pid)
		}

		// 마커: 선택 여부만 표시, 커서는 배경색으로 구분
		var marker string
		if isSel {
			marker = fg256(46) + "✓" + reset
		} else {
			marker = " "
		}

		// 컬럼별 문자열
		idStr := fmt.Sprintf("%-*s", c.id, truncate(p.ID, c.id))
		groupStr := fmt.Sprintf("%-*s", c.group, truncate(p.Group, c.group))
		stateStr := fmt.Sprintf("%-*s", c.state, truncate(stateField(p.State, p.StopReason), c.state))
		pidStr := fmt.Sprintf("%*s", c.pid, pid)
		cntStr := fmt.Sprintf("%*d", c.cnt, p.StartCount)
		startStr := fmt.Sprintf("%-*s", c.start, fmtStart(p.LastStart))
		descStr := fmt.Sprintf("%-*s", c.desc, truncate(p.Desc, c.desc))

		rowColor := stateColor(p.State)

		if isCursor {
			// 커서 행: 전체 반전 배경
			line := idStr + " " + groupStr + " " + stateStr + " " + pidStr + " " + cntStr + " " + startStr + " " + descStr
			b.WriteString(marker + " " + bg256(237) + fgBrightWhite + line + reset + "\n")
		} else {
			// CNT 컬럼 별도 색상 (1이 아니면 주황)
			cntColor := rowColor
			if p.StartCount != 1 {
				cntColor = fg256(208)
			}

			var line string
			if isSel {
				line = bg256(22) + rowColor +
					idStr + " " + groupStr + " " + stateStr + " " + pidStr + " " +
					reset + bg256(22) + cntColor + cntStr + reset + bg256(22) + rowColor +
					" " + startStr + " " + descStr + reset
			} else {
				line = rowColor +
					idStr + " " + groupStr + " " + stateStr + " " + pidStr + " " +
					reset + cntColor + cntStr + reset + rowColor +
					" " + startStr + " " + descStr + reset
			}
			b.WriteString(marker + " " + line + "\n")
		}
	}

	// 뷰포트 남은 공간을 빈 줄로 채워 하단 구분선 위치 고정
	rendered := end - m.scrollOffset
	for i := rendered; i < vs; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

// 상태별 기본 색상
func stateColor(state int) string {
	switch state {
	case 2: // RUNNING
		return fg256(46)
	case 3: // STOPPING → 빨간색
		return fg256(196)
	case 4: // BACKOFF
		return fg256(226)
	case 5: // DISABLED
		return fg256(160)
	default: // WAIT_*
		return fg256(244)
	}
}

// ── 상태 줄 ──────────────────────────────────────────────────────────────────

// renderStatusLine: 상태 줄(1줄) — 로딩 중이면 스피너, 오류면 빨간색, 정상이면 초록색.
func (m Model) renderStatusLine(w int) string {
	if m.loading {
		return fg256(51) + " ⟳ goTbnova 접속 중..." + reset + "\n"
	}
	if m.statusLine == "" {
		return "\n"
	}
	if m.statusErr {
		return fg256(196) + " ✗ " + m.statusLine + reset + "\n"
	}
	return fg256(46) + " ✓ " + m.statusLine + reset + "\n"
}

// ── 확인 다이얼로그 ───────────────────────────────────────────────────────────

// renderConfirm: y/N 확인 다이얼로그 줄을 반환한다.
func (m Model) renderConfirm(w int) string {
	maxLabel := w - 32
	if maxLabel < 10 {
		maxLabel = 10
	}
	msg := fmt.Sprintf(" [ %s ]  실행하시겠습니까? (y/N) ", truncate(m.pending.label, maxLabel))
	return bg256(52) + fgBrightWhite + bold + padRight(msg, w) + reset + "\n"
}

// ── 단축키 힌트 ──────────────────────────────────────────────────────────────

// renderHint: 화면 하단 단축키 힌트 줄을 반환한다.
func (m Model) renderHint(w int) string {
	line := " f:갱신 a/u:전체선택/해제 Space:개별선택/해제 s/k:시작/종료 q:종료  ?:도움말"
	return fg256(244) + truncate(line, w) + reset + "\n"
}

// ── 프로세스 상세 화면 ────────────────────────────────────────────────────────

// renderDetailScreen: h 키로 진입하는 프로세스 상세 화면을 렌더링한다.
// 콘텐츠를 lines 슬라이스로 구성한 뒤 detailScroll 오프셋으로 뷰포트를 잘라 출력한다.
func (m Model) renderDetailScreen(w int) string {
	p := m.detailProc

	// ── 콘텐츠 줄 목록 구성 ──────────────────────────────────────────────
	var lines []string
	add := func(s string) { lines = append(lines, s) }

	// 필드 렌더 헬퍼 (줄 단위로 add)
	field := func(name, val string) {
		lbl := padRight("  "+name, 18)
		add(fg256(214) + bold + lbl + reset + " " + fg256(15) + val + reset)
	}
	fieldColor := func(name, val, color string) {
		lbl := padRight("  "+name, 18)
		add(fg256(214) + bold + lbl + reset + " " + color + val + reset)
	}
	sec := func(title string) {
		add("")
		add(" " + bold + fg256(39) + title + reset)
	}

	// ── 기본 정보 ──────────────────────────────────────────────────────────
	sec("■ 기본 정보")
	field("ID", p.ID)
	field("Group", p.Group)
	desc := p.Desc
	if desc == "" {
		desc = "(없음)"
	}
	field("Description", desc)
	if p.Mode != "" {
		field("Mode", p.Mode)
	}
	if p.Path != "" {
		field("Path", p.Path)
	}
	if len(p.Args) > 0 {
		field("Args", strings.Join(p.Args, " "))
	}

	// ── 스케줄 정보 ────────────────────────────────────────────────────────
	if p.StMode != "" || p.St != "" || p.Et != "" {
		sec("■ 스케줄")
		if p.StMode != "" {
			field("StMode", p.StMode)
		}
		if p.St != "" || p.Et != "" {
			field("Time", fmt.Sprintf("%s ~ %s", p.St, p.Et))
		}
		weekendStr := "아니오"
		if p.Weekend {
			weekendStr = "예"
		}
		field("Weekend", weekendStr)
	}

	// ── 재시작 정책 ────────────────────────────────────────────────────────
	if p.Restart.Policy != "" {
		sec("■ 재시작 정책")
		field("Policy", p.Restart.Policy)
		backoffStr := "아니오"
		if p.Restart.Backoff {
			backoffStr = "예"
		}
		field("Backoff", backoffStr)
		if p.Restart.MaxCount > 0 {
			field("Max Count", fmt.Sprintf("%d", p.Restart.MaxCount))
		}
		if p.Restart.GracePeriod > 0 {
			field("Grace Period", fmt.Sprintf("%ds", p.Restart.GracePeriod))
		}
	}

	// ── 의존성 ─────────────────────────────────────────────────────────────
	if len(p.DependsOn) > 0 {
		sec("■ 의존성")
		healthyStr := "아니오"
		healthyColor := fg256(15)
		if p.DependsHealthy {
			healthyStr = "예"
			healthyColor = fg256(46)
		}
		fieldColor("Healthy Required", healthyStr, healthyColor)
		for i, dep := range p.DependsOn {
			field(fmt.Sprintf("DependsOn[%d]", i), dep)
		}
	}
	if p.Cascade.OnDeath != "" {
		cascadeColor := fg256(226)
		if p.Cascade.OnDeath == "stop_dependents" {
			cascadeColor = fg256(196)
		}
		fieldColor("Cascade", p.Cascade.OnDeath, cascadeColor+bold)
	}

	// ── 실행 상태 ──────────────────────────────────────────────────────────
	sec("■ 실행 상태")
	fieldColor("State", stateNameFull(p.State), stateColor(p.State)+bold)

	pid := "-"
	if p.Pid > 0 {
		pid = fmt.Sprintf("%d", p.Pid)
	}
	field("PID", pid)

	cntColor := fg256(15)
	cntNote := ""
	if p.StartCount != 1 {
		cntColor = fg256(208) + bold
		cntNote = fmt.Sprintf("  ← 재시작 %d회", p.StartCount-1)
	}
	fieldColor("Start Count", fmt.Sprintf("%d%s", p.StartCount, cntNote), cntColor)

	stopStr := stopReasonName(p.StopReason)
	stopColor := fg256(15)
	if p.StopReason != 0 {
		stopColor = fg256(196)
	}
	fieldColor("Stop Reason", stopStr, stopColor)

	// ── 시간 정보 ──────────────────────────────────────────────────────────
	sec("■ 시간 정보")

	lastStart := "-"
	if !p.LastStart.IsZero() {
		elapsed := fmtDuration(time.Since(p.LastStart))
		lastStart = fmt.Sprintf("%s  (경과: %s)", p.LastStart.Format("2006-01-02 15:04:05"), elapsed)
	}
	field("Last Start", lastStart)

	if p.NextRetry != nil {
		until := time.Until(*p.NextRetry)
		retryStr := p.NextRetry.Format("2006-01-02 15:04:05")
		if until > 0 {
			retryStr += fmt.Sprintf("  (후: %s)", fmtDuration(until))
		} else {
			retryStr += "  (대기 중)"
		}
		fieldColor("Next Retry", retryStr, fg256(226))
	} else {
		field("Next Retry", "-")
	}
	field("조회 시각", m.lastUpdate.Format("2006-01-02 15:04:05"))

	// ── 프로세스 환경변수 ──────────────────────────────────────────────────
	if len(p.Env) > 0 {
		sec("■ 프로세스 환경변수")
		keys := make([]string, 0, len(p.Env))
		for k := range p.Env {
			keys = append(keys, k)
		}
		for i := 0; i < len(keys)-1; i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		maxK := 0
		for _, k := range keys {
			if len([]rune(k)) > maxK {
				maxK = len([]rune(k))
			}
		}
		if maxK > 30 {
			maxK = 30
		}
		valW := w - 22 - maxK
		if valW < 10 {
			valW = 10
		}
		for _, k := range keys {
			kStr := fmt.Sprintf("%-*s", maxK, truncate(k, maxK))
			vStr := truncate(p.Env[k], valW)
			add("  " + fg256(39) + bold + kStr + reset + " = " + fg256(15) + vStr + reset)
		}
	}

	// ── 선택 상태 ──────────────────────────────────────────────────────────
	add("")
	selStr := "아니오"
	selColor := fg256(244)
	if m.sel[p.ID] {
		selStr = "예 (✓ 선택됨)"
		selColor = fg256(46) + bold
	}
	fieldColor("선택 상태", selStr, selColor)

	// ── 스크롤 렌더링 ─────────────────────────────────────────────────────
	var b strings.Builder

	// bodyH = 터미널 높이 - 타이틀(1) - 구분선(1) - 하단구분선(1) - 푸터(1)
	bodyH := m.height - 4
	if bodyH < 1 {
		bodyH = 1
	}

	total := len(lines)
	start := m.detailScroll
	if start > total-bodyH {
		start = total - bodyH
	}
	if start < 0 {
		start = 0
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	// 스크롤 위치 표시 (스크롤 가능할 때만)
	scrollInfo := ""
	if total > bodyH {
		pct := 0
		if total > bodyH {
			pct = start * 100 / (total - bodyH)
		}
		scrollInfo = fmt.Sprintf("  %s↑↓ 스크롤  %d/%d줄  %d%%%s",
			fg256(226), start+1, total, pct, reset+fgBrightWhite)
		if start > 0 {
			scrollInfo = fmt.Sprintf("  %s▲ %s↑↓ 스크롤  %d/%d줄  %d%%%s",
				fg256(46), fg256(226), start+1, total, pct, reset+fgBrightWhite)
		}
		if end < total {
			scrollInfo += fg256(46) + " ▼" + reset + fgBrightWhite
		}
	}

	titleText := fmt.Sprintf(" 상세 정보  —  %s%s", p.ID, scrollInfo)
	b.WriteString(bold + bg256(20) + fgBrightWhite + padRight(titleText, w) + reset + "\n")
	b.WriteString(fg256(240) + hline(w) + reset + "\n")

	for i := start; i < end; i++ {
		b.WriteString(lines[i] + "\n")
	}
	for i := end - start; i < bodyH; i++ {
		b.WriteString("\n")
	}

	b.WriteString(fg256(240) + hline(w) + reset + "\n")
	b.WriteString(fg256(244) + " 아무 키나 누르면 목록으로 돌아갑니다  (↑↓ 스크롤)" + reset + "\n")

	return b.String()
}

// 시작 시각을 mm/dd-hh:mm:ss 형태로 변환
func fmtStart(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("01/02-15:04:05")
}

// 경과 시간을 사람이 읽기 좋은 형태로 변환
func fmtDuration(d time.Duration) string {
	d = d.Truncate(time.Second)
	if d < 0 {
		d = -d
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// ── 도움말 화면 ──────────────────────────────────────────────────────────────

// ── 전체 의존성 트리 화면 ─────────────────────────────────────────────────────

// renderDepsScreen: d 키로 진입하는 전체 의존성 트리 화면을 렌더링한다.
// DFS 재귀로 트리를 구성하고 순환 참조를 방문 맵으로 탐지해 표시한다.
func (m Model) renderDepsScreen(w int) string {
	procs := m.procs

	// ── 자료구조 구성 ─────────────────────────────────────────────────────
	procMap := make(map[string]ProcessSnapshot, len(procs))
	for _, p := range procs {
		procMap[p.ID] = p
	}

	// childrenMap[parentID] = 이 부모에 depend하는 자식 ID 목록 (등장 순서 유지)
	childrenMap := make(map[string][]string)
	childSet := make(map[string]map[string]bool) // 중복 방지
	hasParent := make(map[string]bool)

	for _, p := range procs {
		for _, dep := range p.DependsOn {
			if childSet[dep] == nil {
				childSet[dep] = make(map[string]bool)
			}
			if !childSet[dep][p.ID] {
				childSet[dep][p.ID] = true
				childrenMap[dep] = append(childrenMap[dep], p.ID)
			}
			hasParent[p.ID] = true
		}
	}

	// 루트 = depends_on이 없는 프로세스 (m.procs 순서 유지)
	var roots []ProcessSnapshot
	var standalone []ProcessSnapshot // 루트이면서 자식도 없는 것
	for _, p := range procs {
		if !hasParent[p.ID] {
			if len(childrenMap[p.ID]) > 0 {
				roots = append(roots, p)
			} else {
				standalone = append(standalone, p)
			}
		}
	}

	// depends_on이 있지만 부모가 현재 목록에 없는 프로세스도 루트로 처리
	for _, p := range procs {
		if hasParent[p.ID] {
			allMissing := true
			for _, dep := range p.DependsOn {
				if _, ok := procMap[dep]; ok {
					allMissing = false
					break
				}
			}
			if allMissing {
				roots = append(roots, p)
			}
		}
	}

	// ── 줄 목록 구성 ─────────────────────────────────────────────────────
	var lines []string
	add := func(s string) { lines = append(lines, s) }

	// 방문 추적: id → 처음 등장한 부모 ID ("" = 루트)
	visited := make(map[string]string)

	// 재귀 DFS 렌더러
	// conts: 각 depth에서 부모가 아직 형제를 더 가지는지 여부 (│ 표시용)
	var renderNode func(id, parentID string, conts []bool, isLast bool)
	renderNode = func(id, parentID string, conts []bool, isLast bool) {
		// 트리 접두어 구성
		prefix := ""
		for _, cont := range conts {
			if cont {
				prefix += fg256(240) + "│  " + reset
			} else {
				prefix += "   "
			}
		}
		var connector string
		if isLast {
			connector = fg256(240) + "└─ " + reset
		} else {
			connector = fg256(240) + "├─ " + reset
		}
		if len(conts) == 0 {
			connector = "" // 루트는 접두어 없음
			prefix = ""
		}

		p, exists := procMap[id]

		// 이미 방문한 노드: 참조 표시만 하고 재귀 중단
		if firstParent, seen := visited[id]; seen {
			ref := id
			if firstParent != "" {
				ref = id + " ↑  (→ " + firstParent + " 하위 참조)"
			} else {
				ref = id + " ↑  (위 참조)"
			}
			add(prefix + connector + fg256(244) + ref + reset)
			return
		}
		visited[id] = parentID

		// 노드 행 렌더링
		var row string
		if !exists {
			// 목록에 없는 외부 프로세스 (의존 대상이지만 미등록)
			row = prefix + connector + fg256(244) + id + "  (미등록)" + reset
		} else {
			sf := stateField(p.State, p.StopReason)
			stateColor := stateColor(p.State)

			// 캐스케이드 표시
			cascade := ""
			switch p.Cascade.OnDeath {
			case "stop_dependents":
				cascade = "  " + fg256(196) + "[cascade:stop]" + reset
			case "notify":
				cascade = "  " + fg256(226) + "[cascade:notify]" + reset
			}

			// depends_on 의존 충족 여부 (자식 입장)
			depStatus := ""
			if len(p.DependsOn) > 0 && len(conts) > 0 {
				allOk := true
				for _, dep := range p.DependsOn {
					if dp, ok := procMap[dep]; !ok || dp.State != 2 {
						allOk = false
						break
					}
				}
				if allOk {
					depStatus = "  " + fg256(46) + "✓선행충족" + reset
				} else {
					depStatus = "  " + fg256(196) + "✗선행미충족" + reset
				}
			}

			row = prefix + connector +
				stateColor + bold + fmt.Sprintf("%-14s", truncate(id, 14)) + reset +
				"  " + stateColor + fmt.Sprintf("%-12s", truncate(sf, 12)) + reset +
				cascade + depStatus
		}
		add(row)

		// 자식 재귀
		children := childrenMap[id]
		for i, childID := range children {
			childIsLast := i == len(children)-1
			var childConts []bool
			if len(conts) == 0 {
				// 루트 자식: 부모(루트)의 형제 여부는 호출자가 결정
				childConts = []bool{!isLast}
			} else {
				childConts = append(append([]bool{}, conts...), !isLast)
			}
			renderNode(childID, id, childConts, childIsLast)
		}
	}

	// 의존성 체인 섹션
	hasDeps := len(roots) > 0
	if hasDeps {
		add(" " + bold + fg256(214) + "■ 의존성 체인" + reset)
		add("")
		for ri, root := range roots {
			rootIsLast := ri == len(roots)-1
			_ = rootIsLast
			renderNode(root.ID, "", []bool{}, true)
			add("")
		}
	}

	// 의존성은 있지만 부모가 없어서 분류 못된 나머지 (직접 depends_on 있는 고아)
	orphans := []ProcessSnapshot{}
	for _, p := range procs {
		if _, seen := visited[p.ID]; !seen && hasParent[p.ID] {
			orphans = append(orphans, p)
		}
	}
	if len(orphans) > 0 {
		add(" " + bold + fg256(226) + "■ 미연결 의존 프로세스 (선행 프로세스 미등록)" + reset)
		add("")
		for oi, p := range orphans {
			renderNode(p.ID, "", []bool{}, oi == len(orphans)-1)
			for _, dep := range p.DependsOn {
				isLast := dep == p.DependsOn[len(p.DependsOn)-1]
				connector := "├─ "
				if isLast {
					connector = "└─ "
				}
				if dp, ok := procMap[dep]; ok {
					sf := stateField(dp.State, dp.StopReason)
					add("   " + fg256(240) + connector + reset +
						fg256(244) + "↑선행: " + dep + "  [" + sf + "]" + reset)
				} else {
					add("   " + fg256(240) + connector + reset +
						fg256(244) + "↑선행: " + dep + "  (미등록)" + reset)
				}
			}
		}
		add("")
	}

	// 독립 프로세스 섹션
	if len(standalone) > 0 {
		add(" " + bold + fg256(244) + "■ 독립 프로세스 (의존성 없음)" + reset)
		add("")
		for _, p := range standalone {
			sf := stateField(p.State, p.StopReason)
			sc := stateColor(p.State)
			add("   " + sc + bold + fmt.Sprintf("%-14s", truncate(p.ID, 14)) + reset +
				"  " + sc + fmt.Sprintf("%-12s", sf) + reset +
				"  " + fg256(244) + truncate(p.Desc, w-35) + reset)
		}
	}

	// 의존성 데이터 없을 때 안내
	noDepsData := true
	for _, p := range procs {
		if len(p.DependsOn) > 0 || p.Cascade.OnDeath != "" {
			noDepsData = false
			break
		}
	}
	if noDepsData && len(procs) > 0 {
		lines = nil
		add("")
		add("  " + fg256(226) + "⚠ 의존성 정보가 없습니다." + reset)
		add("  " + fg256(244) + "goTbnova가 status 응답에 depends_on / cascade 필드를 포함해야 합니다." + reset)
		add("")
		add("  현재 등록된 프로세스:")
		for _, p := range procs {
			sf := stateField(p.State, p.StopReason)
			sc := stateColor(p.State)
			add("    " + sc + fmt.Sprintf("%-16s", truncate(p.ID, 16)) +
				"  " + sf + reset)
		}
	}

	// ── 스크롤 렌더링 ─────────────────────────────────────────────────────
	var b strings.Builder

	chainCnt := len(roots)
	aloneCnt := len(standalone)

	bodyH := m.height - 4
	if bodyH < 1 {
		bodyH = 1
	}

	total := len(lines)
	start := m.depsScroll
	if start > total-bodyH {
		start = total - bodyH
	}
	if start < 0 {
		start = 0
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	scrollInfo := ""
	if total > bodyH {
		pct := 0
		if total > bodyH {
			pct = start * 100 / (total - bodyH)
		}
		upMark := "  "
		if start > 0 {
			upMark = fg256(46) + "▲ " + reset + fgBrightWhite
		}
		downMark := ""
		if end < total {
			downMark = " " + fg256(46) + "▼" + reset + fgBrightWhite
		}
		scrollInfo = fmt.Sprintf("  %s%s%d/%d줄 %d%%%s",
			upMark, fg256(226), start+1, total, pct, reset+fgBrightWhite+downMark)
	}

	titleText := fmt.Sprintf(" 의존성 트리  (체인: %d  독립: %d)%s",
		chainCnt, aloneCnt, scrollInfo)
	b.WriteString(bold + bg256(20) + fgBrightWhite + padRight(titleText, w) + reset + "\n")
	b.WriteString(fg256(240) + hline(w) + reset + "\n")

	for i := start; i < end; i++ {
		b.WriteString(lines[i] + "\n")
	}
	for i := end - start; i < bodyH; i++ {
		b.WriteString("\n")
	}

	b.WriteString(fg256(240) + hline(w) + reset + "\n")
	b.WriteString(fg256(244) + " 아무 키나 누르면 목록으로 돌아갑니다  (↑↓ 스크롤)" + reset + "\n")

	return b.String()
}

// ── 환경변수 조회 명령 ────────────────────────────────────────────────────────

// doEnvFetch: goTbnova에 env 명령을 보내고 결과를 envMsg로 반환하는 Cmd를 만든다.
func doEnvFetch(c *Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.Env()
		if err != nil {
			return envMsg{err: err}
		}
		if resp.Result == "error" {
			return envMsg{err: fmt.Errorf("%s", resp.Message)}
		}
		return envMsg{vars: resp.Env}
	}
}

// ── 환경변수 화면 ─────────────────────────────────────────────────────────────

// renderEnvScreen: e 키로 진입하는 전역 환경변수 목록 화면을 렌더링한다.
// 키를 알파벳 순으로 정렬하고 envScroll 오프셋으로 스크롤한다.
func (m Model) renderEnvScreen(w int) string {
	var lines []string
	add := func(s string) { lines = append(lines, s) }

	vars := m.envVars

	if len(vars) == 0 {
		add("  " + fg256(244) + "(환경변수 없음)" + reset)
	} else {
		// 키 정렬
		keys := make([]string, 0, len(vars))
		for k := range vars {
			keys = append(keys, k)
		}
		// 알파벳 순 정렬
		for i := 0; i < len(keys)-1; i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		// 가장 긴 키 길이 계산 (정렬용)
		maxKey := 0
		for _, k := range keys {
			if len([]rune(k)) > maxKey {
				maxKey = len([]rune(k))
			}
		}
		if maxKey > 40 {
			maxKey = 40
		}

		valW := w - maxKey - 5 // "  KEY = VALUE"
		if valW < 10 {
			valW = 10
		}

		for _, k := range keys {
			v := vars[k]
			keyStr := fmt.Sprintf("%-*s", maxKey, truncate(k, maxKey))
			valStr := truncate(v, valW)
			line := "  " + fg256(39) + bold + keyStr + reset +
				" " + fg256(240) + "=" + reset +
				" " + fg256(15) + valStr + reset
			add(line)
		}
	}

	// 스크롤 렌더링 (help와 동일 구조)
	var b strings.Builder

	titleText := fmt.Sprintf(" goTbnova Monitor  전역 환경변수  (%d개)  ↑↓:스크롤", len(vars))
	b.WriteString(bold + bg256(20) + fgBrightWhite + padRight(titleText, w) + reset + "\n")
	b.WriteString(fg256(240) + hline(w) + reset + "\n")

	bodyH := m.height - 4
	if bodyH < 1 {
		bodyH = 1
	}

	total := len(lines)
	start := m.envScroll
	if start > total-bodyH {
		start = total - bodyH
	}
	if start < 0 {
		start = 0
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	for i := start; i < end; i++ {
		b.WriteString(lines[i] + "\n")
	}
	for i := end - start; i < bodyH; i++ {
		b.WriteString("\n")
	}

	b.WriteString(fg256(240) + hline(w) + reset + "\n")
	b.WriteString(fg256(244) + " 아무 키나 누르면 목록으로 돌아갑니다" + reset + "\n")

	return b.String()
}

// ── 그룹 트리 화면 ───────────────────────────────────────────────────────────

// renderTreeScreen: g 키로 진입하는 그룹 트리 화면을 렌더링한다.
// 프로세스를 Group 별로 묶어 트리 접두어(├─/└─)로 표시한다.
func (m Model) renderTreeScreen(w int) string {
	// ── 그룹별 프로세스 수집 (m.procs 순서 유지) ─────────────────────────
	type groupEntry struct {
		name  string
		procs []ProcessSnapshot
	}
	groupMap := make(map[string]*groupEntry)
	var groupOrder []string

	for _, p := range m.procs {
		g := p.Group
		if g == "" {
			g = "(그룹없음)"
		}
		if _, ok := groupMap[g]; !ok {
			groupMap[g] = &groupEntry{name: g}
			groupOrder = append(groupOrder, g)
		}
		cp := p
		groupMap[g].procs = append(groupMap[g].procs, cp)
	}

	// ── 컬럼 너비 (트리 접두어 3칸 + 선택마커 1칸 + 공백 1칸 = 5칸 오프셋) ──
	const treePrefix = 5 // "├─ ✓ " or "└─   "
	idW, stateW, pidW, cntW, startW := 13, 12, 6, 4, 14
	fixed := treePrefix + idW + 1 + stateW + 1 + pidW + 1 + cntW + 1 + startW + 1
	descW := w - fixed
	if descW < 6 {
		descW = 6
	}

	// ── 줄 목록 구성 ─────────────────────────────────────────────────────
	var lines []string
	add := func(s string) { lines = append(lines, s) }

	for gi, gname := range groupOrder {
		ge := groupMap[gname]

		// 그룹별 상태 집계
		run, stop, back, dis, wait := 0, 0, 0, 0, 0
		for _, p := range ge.procs {
			switch p.State {
			case 2:
				run++
			case 3:
				stop++
			case 4:
				back++
			case 5:
				dis++
			default:
				wait++
			}
		}

		// 그룹 헤더 빈줄 (첫 그룹 제외)
		if gi > 0 {
			add("")
		}

		summary := fmt.Sprintf(
			"  %s%s%s  총%d  %s실행%d%s %s대기%d%s %s종료중%d%s %s백오프%d%s %s비활성%d%s",
			bold+fg256(39), gname, reset,
			len(ge.procs),
			fg256(46), run, fg256(250),
			fg256(244), wait, fg256(250),
			fg256(196), stop, fg256(250),
			fg256(226), back, fg256(250),
			fg256(160), dis, reset,
		)
		add(summary)

		// 컬럼 헤더 (트리 들여쓰기)
		hdr := fmt.Sprintf("  %s   %-*s %-*s %*s %*s %-*s %-*s%s",
			fg256(240)+"─────"+reset,
			idW, "ID",
			stateW, "STATE",
			pidW, "PID",
			cntW, "CNT",
			startW, "START",
			descW, "DESC",
			reset,
		)
		add(bold + fg256(214) + hdr + reset)

		// 프로세스 행
		for pi, p := range ge.procs {
			isLast := pi == len(ge.procs)-1

			var branch string
			if isLast {
				branch = fg256(240) + "  └─" + reset
			} else {
				branch = fg256(240) + "  ├─" + reset
			}

			selMark := " "
			if m.sel[p.ID] {
				selMark = fg256(46) + "✓" + reset
			}

			pidS := "-"
			if p.Pid > 0 {
				pidS = fmt.Sprintf("%d", p.Pid)
			}

			idStr    := fmt.Sprintf("%-*s", idW,    truncate(p.ID, idW))
			stateStr := fmt.Sprintf("%-*s", stateW, truncate(stateField(p.State, p.StopReason), stateW))
			pidStr   := fmt.Sprintf("%*s",  pidW,   pidS)
			cntStr   := fmt.Sprintf("%*d",  cntW,   p.StartCount)
			startStr := fmt.Sprintf("%-*s", startW, fmtStart(p.LastStart))
			descStr  := fmt.Sprintf("%-*s", descW,  truncate(p.Desc, descW))

			rowColor := stateColor(p.State)
			cntColor := rowColor
			if p.StartCount != 1 {
				cntColor = fg256(208)
			}

			row := branch + " " + selMark + " " +
				rowColor + idStr + " " + stateStr + " " + pidStr + " " +
				reset + cntColor + cntStr + reset + rowColor +
				" " + startStr + " " + descStr + reset

			add(row)
		}
	}

	// 프로세스 없을 때
	if len(m.procs) == 0 {
		lines = []string{"  " + fg256(244) + "(프로세스 없음 — f키로 목록 로드)" + reset}
	}

	// ── 스크롤 렌더링 ─────────────────────────────────────────────────────
	var b strings.Builder

	bodyH := m.height - 4 // 타이틀(1)+구분선(1)+하단구분선(1)+푸터(1)
	if bodyH < 1 {
		bodyH = 1
	}

	total := len(lines)
	start := m.treeScroll
	if start > total-bodyH {
		start = total - bodyH
	}
	if start < 0 {
		start = 0
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	scrollInfo := ""
	if total > bodyH {
		pct := 0
		if total > bodyH {
			pct = start * 100 / (total - bodyH)
		}
		upMark := "  "
		if start > 0 {
			upMark = fg256(46) + "▲ " + reset + fgBrightWhite
		}
		downMark := ""
		if end < total {
			downMark = " " + fg256(46) + "▼" + reset + fgBrightWhite
		}
		scrollInfo = fmt.Sprintf("  %s%s%d/%d줄 %d%%%s",
			upMark, fg256(226), start+1, total, pct, reset+fgBrightWhite+downMark)
	}

	titleText := fmt.Sprintf(" 그룹 트리  (%d 그룹 / %d 프로세스)%s",
		len(groupOrder), len(m.procs), scrollInfo)
	b.WriteString(bold + bg256(20) + fgBrightWhite + padRight(titleText, w) + reset + "\n")
	b.WriteString(fg256(240) + hline(w) + reset + "\n")

	for i := start; i < end; i++ {
		b.WriteString(lines[i] + "\n")
	}
	for i := end - start; i < bodyH; i++ {
		b.WriteString("\n")
	}

	b.WriteString(fg256(240) + hline(w) + reset + "\n")
	b.WriteString(fg256(244) + " 아무 키나 누르면 목록으로 돌아갑니다  (↑↓ 스크롤)" + reset + "\n")

	return b.String()
}

// renderHelpScreen: ? 키로 진입하는 도움말 화면을 렌더링한다. helpScroll로 스크롤된다.
func (m Model) renderHelpScreen(w int) string {
	// 도움말 줄 목록을 먼저 구성한 뒤 터미널 높이에 맞게 출력
	type line struct{ s string }
	var lines []line

	add := func(s string) { lines = append(lines, line{s}) }

	type entry struct {
		key    string // 화면 표시용 (ANSI 포함 가능)
		keyVis int    // 키의 실제 시각 너비 (ANSI 제외)
		desc   string
	}

	const keyCol = 18 // 키 컬럼 시각 너비

	row := func(e entry) string {
		pad := keyCol - e.keyVis
		if pad < 1 {
			pad = 1
		}
		return "  " + fg256(39) + e.key + reset +
			strings.Repeat(" ", pad) + e.desc
	}

	sec := func(title string) {
		add("")
		add(" " + bold + fg256(214) + title + reset)
	}

	// ── 접속 / 명령 ──────────────────────────────────────────────────────
	sec("🔵 접속 / 명령")
	add(row(entry{"f", 1, "프로세스 목록 재요청  (goTbnova 접속 → 상태 수신 → 단절)"}))
	add(row(entry{"t", 1, "goTbnova에 Process Table Reload 명령 전송"}))
	add(row(entry{"s", 1, "선택 프로세스 Start  (미선택 시 동작 안 함)"}))
	add(row(entry{"k", 1, "선택 프로세스 Kill   (미선택 시 동작 안 함)"}))
	add(row(entry{"h", 1, "커서 프로세스 상세 정보 화면  (아무 키나 누르면 복귀)"}))
	add(row(entry{"g", 1, "그룹 트리 화면  (아무 키나 누르면 복귀)"}))
	add(row(entry{"e", 1, "전역 환경변수 조회  (goTbnova 접속 → 수신 → 단절)"}))
	add(row(entry{"d", 1, "전체 의존성 트리 화면  (아무 키나 누르면 복귀)"}))

	// ── 선택 ─────────────────────────────────────────────────────────────
	sec("🔵 선택")
	add(row(entry{"Space", 5, "개별 선택 토글  (커서 자동 아래 이동)"}))
	add(row(entry{"a", 1, "전체 선택  (현재 필터 목록 기준)"}))
	add(row(entry{"u", 1, "모든 선택 해제"}))
	// ── 커서 이동 ─────────────────────────────────────────────────────────
	sec("🔵 커서 이동")
	add(row(entry{"↑ / ↓", 5, "한 줄 위 / 아래"}))
	add(row(entry{"PgUp / Ctrl+B", 13, "한 페이지 위"}))
	add(row(entry{"PgDn / Ctrl+F", 13, "한 페이지 아래"}))
	add(row(entry{"Home", 4, "전체 처음"}))
	add(row(entry{"End",  3, "전체 마지막"}))
	add(row(entry{"/",    1, "ID·그룹·설명 검색  (Esc 해제)"}))

	// ── 색상 범례 ─────────────────────────────────────────────────────────
	sec("🔵 색상 범례 / STATE 약어")
	add(row(entry{fg256(46) + "초록" + reset,  3, "RUN  = RUNNING"}))
	add(row(entry{fg256(196) + "빨강" + reset, 3, "STOP = STOPPING  (종료중)"}))
	add(row(entry{fg256(226) + "노랑" + reset, 3, "BK   = BACKOFF  (재시도 대기)"}))
	add(row(entry{fg256(244) + "회색" + reset, 3, "WT   = WAIT_TIME  /  WD = WAIT_DEPEND"}))
	add(row(entry{fg256(208) + "주황" + reset, 3, "CNT ≠ 1  (재시작 이력 있음)  — CNT 열에만 적용"}))

	// ── 기타 ─────────────────────────────────────────────────────────────
	sec("🔵 기타")
	add(row(entry{"?", 0, "이 도움말"}))
	add(row(entry{"q / Ctrl+C", 9, "종료"}))

	// ── 렌더링: 터미널 높이에 맞게 출력 ─────────────────────────────────
	var b strings.Builder

	title := bold + bg256(20) + fgBrightWhite +
		padRight(" goTbnova Monitor  도움말  (↑↓ 스크롤)", w) + reset + "\n"
	footer := fg256(240) + hline(w) + reset + "\n" +
		fg256(244) + " 아무 키나 누르면 목록으로 돌아갑니다" + reset + "\n"

	// 타이틀(1) + 구분선(1) + 푸터(2) = 4줄 고정
	bodyH := m.height - 4
	if bodyH < 1 {
		bodyH = 1
	}

	// 스크롤 오프셋 적용
	total := len(lines)
	start := m.helpScroll
	if start > total-bodyH {
		start = total - bodyH
	}
	if start < 0 {
		start = 0
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	b.WriteString(title)
	b.WriteString(fg256(240) + hline(w) + reset + "\n")
	for i := start; i < end; i++ {
		b.WriteString(lines[i].s + "\n")
	}
	// 빈 줄로 채워 푸터 위치 고정
	for i := end - start; i < bodyH; i++ {
		b.WriteString("\n")
	}
	b.WriteString(footer)

	return b.String()
}

// ── 유틸 ─────────────────────────────────────────────────────────────────────

// visLen: ANSI 코드를 포함하지 않는 문자열의 터미널 표시 너비(칸 수)를 반환한다.
// 한글·CJK 문자는 2칸을 차지하므로 rune 수와 다를 수 있다.
func visLen(s string) int {
	n := 0
	for _, r := range s {
		if (r >= 0x1100 && r <= 0x11FF) || // 한글 자모
			(r >= 0xAC00 && r <= 0xD7A3) || // 한글 음절
			(r >= 0x3040 && r <= 0x9FFF) || // 일본어·CJK 통합 한자
			(r >= 0xF900 && r <= 0xFAFF) { // CJK 호환 한자
			n += 2
		} else {
			n++
		}
	}
	return n
}

// truncate: 문자열 s를 최대 n 룬(rune)으로 잘라 반환한다. 초과 시 ".."을 붙인다.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 2 {
		return string(runes[:n])
	}
	return string(runes[:n-2]) + ".."
}

// padRight: 문자열 s를 터미널 표시 폭 w가 되도록 오른쪽에 공백을 채운다.
// 한글·CJK 문자의 2칸 표시 폭을 visLen으로 올바르게 계산한다.
func padRight(s string, w int) string {
	vl := visLen(s)
	if vl >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vl)
}
