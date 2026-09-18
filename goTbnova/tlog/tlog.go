///
/// @file    tlog/tlog.go
/// @brief   goTbnova 전용 로거 — 이중 백엔드 지원
/// @details 두 가지 로그 백엔드를 지원한다.
///
///          백엔드 선택:
///            goTbnova -log=net  <table>  — centolib/netlog → Vector → VictoriaLogs (기본값)
///            goTbnova -log=file <table>  — 로컬 파일 직접 기록 (스탠드얼론 환경)
///
///          공통 로그 포맷:
///            [net]  JSON → Vector UDS 소켓 → VictoriaLogs
///            [file] TEXT → $LOG_HOME/goTbnova_YYYYMMDD.log
///
///          로컬 파일 포맷 예시:
///            2025/01/04 13:32:58.123 [INFO ] [process_start    ] id=AGENT pid=1234 count=1 mode=P
///            2025/01/04 13:32:59.456 [ERROR] [process_abnormal ] id=DB_AGENT pid=5678 fail_count=3
///            2025/01/04 00:00:00.001 [INFO ] [date_changed     ] weekday=1 date=2025-01-05
///
///          설계 원칙:
///            1. Backend 인터페이스로 두 구현체를 추상화한다.
///            2. 모든 로그는 event= 필드를 포함한다.
///            3. 초기화 실패는 치명적이지 않다 (stderr 폴백).
///            4. 로컬 파일은 일자 변경 시 자동으로 새 파일로 교체된다.
///            5. 로컬 파일 쓰기는 mutex 로 보호한다.
///
/// @date    2026-04-27
/// @author  Cento
///

package tlog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"centolib/netlog"
)

////////////////////////////////////////////////////////////////////////////////
// 상수
////////////////////////////////////////////////////////////////////////////////

/// vectorSocketPath 는 Vector UNIX Datagram 소켓 경로이다.
/// 기본값이며 InitWithSocket() 호출 시 gotbnova.toml 의 값으로 덮어쓴다.
var vectorSocketPath = "/tmp/VECTOR.UNIX.DGRAM"

////////////////////////////////////////////////////////////////////////////////
// 백엔드 인터페이스
////////////////////////////////////////////////////////////////////////////////

/// Backend 는 로그 출력 백엔드 인터페이스이다.
/// netBackend 와 fileBackend 가 구현한다.
type Backend interface {
	/// log 는 레벨과 메시지를 기록한다.
	/// caller depth 는 항상 3 으로 고정한다 (tlog.logPush → Backend.log → runtime.Caller)
	log(level, msg string)

	/// close 는 백엔드를 종료한다.
	close()

	/// rotatIfNeeded 는 일자 변경 시 로그 파일을 교체한다.
	/// netBackend 는 no-op 이다.
	rotateIfNeeded()
}

////////////////////////////////////////////////////////////////////////////////
// 전역 상태
////////////////////////////////////////////////////////////////////////////////

var (
	backend Backend    ///< 현재 활성화된 백엔드 (Init() 후 설정)
	once    sync.Once  ///< Init() 중복 호출 방지
)

/// LogMode 는 로그 백엔드 선택 값을 나타낸다.
type LogMode string

const (
	ModeNet  LogMode = "net"  ///< centolib/netlog 사용 (기본값)
	ModeFile LogMode = "file" ///< 로컬 파일 직접 기록
)

////////////////////////////////////////////////////////////////////////////////
// 초기화 / 종료
////////////////////////////////////////////////////////////////////////////////

/// Init 은 goTbnova 로거를 초기화한다.
///
/// ModeNet 동작:
///   Vector UNIX socket(/tmp/VECTOR.UNIX.DGRAM) 연결 실패 시 에러를 반환한다.
///   main() 에서 에러 수신 시 os.Exit(1) 로 즉시 종료해야 한다.
///   소켓 파일 존재 여부를 사전 확인하여 명확한 오류 메시지를 제공한다.
///
/// ModeFile 동작:
///   파일/디렉토리 생성 실패 시 에러를 반환한다.
///
/// @param  mode    로그 백엔드 선택 (ModeNet / ModeFile)
/// @param  table   설정 파일 경로 (netlog pextra / 파일명 prefix 에 사용)
/// @param  logDir  로컬 파일 모드 시 로그 디렉토리 ($LOG_HOME). ModeNet 에서는 무시됨.
/// @return         초기화 실패 시 에러 (ModeNet: 소켓 연결 실패, ModeFile: 파일 생성 실패)
func Init(mode LogMode, table, logDir string) error {
	var initErr error
	once.Do(func() {
		switch mode {

		case ModeFile:
			b, err := newFileBackend(logDir)
			if err != nil {
				initErr = fmt.Errorf("로컬 파일 백엔드 초기화 실패: %w", err)
				return
			}
			backend = b
			fmt.Fprintf(os.Stderr, "[INFO] tlog: 로컬 파일 로그 → %s\n", logDir)

		default: /// ModeNet (기본값)
			/// 소켓 파일 존재 여부 사전 확인 — 명확한 오류 메시지 제공
			if _, err := os.Stat(vectorSocketPath); os.IsNotExist(err) {
				initErr = fmt.Errorf(
					"Vector UNIX socket 없음 [%s]\n"+
					"  Vector 에이전트가 실행 중인지 확인하거나\n"+
					"  로컬 파일 로그를 사용하려면 -log=file 옵션을 지정하세요",
					vectorSocketPath)
				return
			}

			/// centolib/netlog 초기화 (UDS Datagram 연결)
			if err := netlog.InitNetLogout("goTbnova", table); err != nil {
				initErr = fmt.Errorf(
					"Vector UNIX socket 연결 실패 [%s]: %w\n"+
					"  Vector 에이전트가 실행 중인지 확인하거나\n"+
					"  로컬 파일 로그를 사용하려면 -log=file 옵션을 지정하세요",
					vectorSocketPath, err)
				return
			}
			backend = &netBackend{}
		}
	})
	return initErr
}

/// InitWithSocket 은 Vector UDS 소켓 경로를 지정하여 로거를 초기화한다.
/// gotbnova.toml [log] vector_socket 값을 전달받아 사용한다.
/// Init() 과 동일하게 동작하되 vectorSocketPath 를 먼저 설정한다.
///
/// @param  mode         로그 백엔드 선택 (ModeNet / ModeFile)
/// @param  table        설정 파일 경로
/// @param  logDir       로컬 파일 모드 시 로그 디렉토리
/// @param  socketPath   Vector UDS 소켓 경로 (빈 문자열이면 기본값 유지)
/// @return              초기화 실패 시 에러
func InitWithSocket(mode LogMode, table, logDir, socketPath string) error {
	if socketPath != "" {
		vectorSocketPath = socketPath
	}
	return Init(mode, table, logDir)
}

/// Close 는 로거를 종료한다.
/// 미전송 로그를 flush 하고 파일/소켓을 닫는다.
/// main() 의 defer 로 호출한다.
func Close() {
	if backend != nil {
		backend.close()
	}
}

/// RotateIfNeeded 는 일자 변경 시 로그 파일을 교체한다.
/// process/manager.go 의 일자 변경 감지 루프에서 호출한다.
func RotateIfNeeded() {
	if backend != nil {
		backend.rotateIfNeeded()
	}
}

////////////////////////////////////////////////////////////////////////////////
// netBackend — centolib/netlog 래퍼
////////////////////////////////////////////////////////////////////////////////

/// netBackend 는 centolib/netlog 를 Backend 인터페이스로 래핑한다.
type netBackend struct{}

func (n *netBackend) log(level, msg string) {
	switch level {
	case "INFO":  netlog.Infor("%s", msg)
	case "WARN":  netlog.Warn("%s", msg)
	case "ERROR": netlog.Error("%s", msg)
	case "DEBUG": netlog.Debug("%s", msg)
	}
}

func (n *netBackend) close()           { netlog.Close() }
func (n *netBackend) rotateIfNeeded()  {}  ///< netlog 는 일자 교체 불필요

////////////////////////////////////////////////////////////////////////////////
// fileBackend — 로컬 파일 직접 기록
////////////////////////////////////////////////////////////////////////////////

/// fileBackend 는 로컬 파일에 직접 로그를 기록하는 백엔드이다.
///
/// 파일명 규칙:
///   {logDir}/goTbnova_YYYYMMDD.log
///
/// Flush 정책:
///   - ERROR / WARN : 즉시 Flush (중요 이벤트 유실 방지)
///   - INFO / DEBUG : flushTicker 2초 주기 자동 Flush
///   - 일자 교체    : 교체 전 Flush
///   - close()      : 종료 전 Flush
///
/// 일자 변경 시 새 파일로 자동 교체된다.
/// 파일 쓰기는 bufio.Writer 로 버퍼링하며 mutex 로 보호한다.
type fileBackend struct {
	mu      sync.Mutex    ///< 파일 쓰기 보호
	logDir  string        ///< 로그 디렉토리
	curDate string        ///< 현재 로그 파일 날짜 "YYYYMMDD"
	file    *os.File      ///< 현재 로그 파일
	writer  *bufio.Writer ///< 버퍼 쓰기 (성능 최적화)
	stopCh  chan struct{}  ///< flushTicker 종료 신호
}

/// newFileBackend 는 fileBackend 를 생성하고 오늘 날짜 로그 파일을 연다.
///
/// @param  logDir  로그 디렉토리 경로
/// @return         fileBackend 포인터
/// @return         파일 생성 실패 시 에러
func newFileBackend(logDir string) (*fileBackend, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("로그 디렉토리 생성 실패 [%s]: %w", logDir, err)
	}
	b := &fileBackend{
		logDir: logDir,
		stopCh: make(chan struct{}),
	}
	if err := b.openFile(time.Now()); err != nil {
		return nil, err
	}
	/// 2초 주기 자동 Flush goroutine 기동
	/// INFO/DEBUG 로그가 버퍼에 쌓이지 않도록 주기적으로 flush 한다.
	go b.flushTicker()
	return b, nil
}

/// openFile 은 날짜 기반 로그 파일을 열거나 새로 생성한다.
///
/// @param  t  파일명 기준 시각
/// @return    에러
func (b *fileBackend) openFile(t time.Time) error {
	date := t.Format("20060102")
	path := filepath.Join(b.logDir, fmt.Sprintf("goTbnova_%s.log", date))

	/// 기존 파일이 열려있으면 flush 후 닫기
	if b.file != nil {
		b.writer.Flush()
		b.file.Close()
	}

	/// O_APPEND: 재기동 시에도 기존 로그에 이어서 기록
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("로그 파일 열기 실패 [%s]: %w", path, err)
	}

	b.file    = f
	b.writer  = bufio.NewWriterSize(f, 64*1024) ///< 64KB 버퍼
	b.curDate = date
	return nil
}

/// log 는 로그 한 줄을 파일에 기록한다.
///
/// 포맷:
///   2025/01/04 13:32:58.123 [INFO ] [process_start    ] id=AGENT pid=1234
///
/// @param  level  로그 레벨 문자열
/// @param  msg    로그 메시지
func (b *fileBackend) log(level, msg string) {
	now := time.Now()

	/// 호출자 정보 획득 (depth 3: logPush → log → 여기)
	_, file, line, ok := runtime.Caller(3)
	src := "unknown:0"
	if ok {
		src = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	/// 레벨 5자 고정폭
	lvl := fmt.Sprintf("%-5s", level)

	entry := fmt.Sprintf("%s [%s] [%-16s] %s  (%s)\n",
		now.Format("2006/01/02 15:04:05.000"),
		lvl,
		extractEvent(msg),
		msg,
		src,
	)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.writer.WriteString(entry)
	/// ERROR / WARN 은 즉시 Flush (중요 이벤트 유실 방지)
	/// INFO / DEBUG 는 flushTicker(2초) 에서 일괄 Flush
	if level == "ERROR" || level == "WARN" {
		b.writer.Flush()
	}
}

/// rotateIfNeeded 는 일자가 변경되었으면 새 로그 파일로 교체한다.
/// process/manager.go 의 일자 변경 루프에서 호출된다.
func (b *fileBackend) rotateIfNeeded() {
	now  := time.Now()
	date := now.Format("20060102")

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.curDate == date {
		return
	}
	if err := b.openFile(now); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] tlog: 로그 파일 교체 실패: %v\n", err)
	}
}

/// close 는 버퍼를 flush 하고 파일을 닫는다.
/// flushTicker goroutine 을 먼저 종료한 뒤 최종 Flush 를 수행한다.
func (b *fileBackend) close() {
	/// flushTicker 종료
	close(b.stopCh)

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.writer != nil { b.writer.Flush() }
	if b.file   != nil { b.file.Close() }
}

/// flushTicker 는 2초 주기로 버퍼를 Flush 하는 백그라운드 goroutine 이다.
/// INFO/DEBUG 로그가 버퍼에 장시간 쌓이는 것을 방지한다.
/// close() 에서 stopCh 를 닫으면 종료된다.
func (b *fileBackend) flushTicker() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			if b.writer != nil {
				b.writer.Flush()
			}
			b.mu.Unlock()

		case <-b.stopCh:
			return
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// stderrBackend — 초기화 실패 시 폴백
////////////////////////////////////////////////////////////////////////////////

/// stderrBackend 는 netlog / 파일 초기화 실패 시 stderr 로 출력하는 폴백 백엔드이다.
type stderrBackend struct{}

func newStderrBackend() *stderrBackend { return &stderrBackend{} }

func (s *stderrBackend) log(level, msg string) {
	fmt.Fprintf(os.Stderr, "%s [%-5s] %s\n",
		time.Now().Format("2006/01/02 15:04:05.000"), level, msg)
}
func (s *stderrBackend) close()          {}
func (s *stderrBackend) rotateIfNeeded() {}

////////////////////////////////////////////////////////////////////////////////
// 내부 공통 로그 함수
////////////////////////////////////////////////////////////////////////////////

/// logPush 는 백엔드로 로그를 전달하는 내부 함수이다.
/// backend 가 nil 이면 (Init 전) 무시한다.
///
/// @param  level   로그 레벨
/// @param  msg     로그 메시지
func logPush(level, msg string) {
	if backend == nil {
		return
	}
	backend.log(level, msg)
}

/// extractEvent 는 로그 메시지에서 event=XXX 값을 추출한다.
/// 파일 백엔드의 컬럼 정렬에 사용한다.
/// event 필드가 없으면 빈 문자열을 반환한다.
///
/// @param  msg  로그 메시지 문자열
/// @return      event 값 (최대 16자)
func extractEvent(msg string) string {
	const prefix = "event="
	idx := strings.Index(msg, prefix)
	if idx < 0 {
		return ""
	}
	rest := msg[idx+len(prefix):]
	end  := strings.IndexByte(rest, ' ')
	if end < 0 {
		end = len(rest)
	}
	event := rest[:end]
	if len(event) > 16 {
		event = event[:16]
	}
	return event
}

////////////////////////////////////////////////////////////////////////////////
// 시스템 이벤트
////////////////////////////////////////////////////////////////////////////////

/// Startup 은 goTbnova 기동 이벤트를 기록한다.
func Startup(table string, port int) {
	logPush("INFO", fmt.Sprintf("event=startup table=%s port=%d", table, port))
}

/// Shutdown 은 goTbnova 종료 이벤트를 기록한다.
func Shutdown(reason string) {
	logPush("INFO", fmt.Sprintf("event=shutdown reason=%s", reason))
}

/// DateChanged 는 일자 변경 이벤트를 기록한다.
func DateChanged(weekday int, date string) {
	logPush("INFO", fmt.Sprintf("event=date_changed weekday=%d date=%s", weekday, date))
}

////////////////////////////////////////////////////////////////////////////////
// 테이블 이벤트
////////////////////////////////////////////////////////////////////////////////

/// TableLoad 는 설정 파일 최초 로드 완료를 기록한다.
func TableLoad(path string, procCnt, groupCnt int) {
	logPush("INFO", fmt.Sprintf(
		"event=table_load path=%s proc_count=%d group_count=%d",
		path, procCnt, groupCnt))
}

/// TableReload 는 테이블 리로드 완료를 기록한다.
func TableReload(path string, added, removed, modified int) {
	logPush("INFO", fmt.Sprintf(
		"event=table_reload path=%s added=%d removed=%d modified=%d",
		path, added, removed, modified))
}

/// TableReloadFail 은 테이블 리로드 실패를 기록한다.
func TableReloadFail(path string, err error) {
	logPush("ERROR", fmt.Sprintf(
		"event=table_reload_fail path=%s err=%s reason=기존설정유지",
		path, err.Error()))
}

////////////////////////////////////////////////////////////////////////////////
// 프로세스 기동/종료 이벤트
////////////////////////////////////////////////////////////////////////////////

/// ProcessStart 는 프로세스 기동 이벤트를 기록한다.
func ProcessStart(id string, pid, count int, mode, path string) {
	logPush("INFO", fmt.Sprintf(
		"event=process_start id=%s pid=%d count=%d mode=%s path=%s",
		id, pid, count, mode, path))
}

/// ProcessExit 는 프로세스 정상 종료 이벤트를 기록한다.
func ProcessExit(id string, pid int, reason string) {
	logPush("INFO", fmt.Sprintf(
		"event=process_exit id=%s pid=%d reason=%s",
		id, pid, reason))
}

/// ProcessAbnormal 은 프로세스 비정상 종료를 기록한다.
/// Grafana 알림 트리거 대상 이벤트이다.
func ProcessAbnormal(id string, pid, failCount int, exitErr string) {
	logPush("ERROR", fmt.Sprintf(
		"event=process_abnormal id=%s pid=%d fail_count=%d exit_err=%s",
		id, pid, failCount, exitErr))
}

/// ProcessStopReq 는 운영자의 프로세스 종료 요청을 기록한다.
func ProcessStopReq(id string, pid, graceSec int) {
	logPush("INFO", fmt.Sprintf(
		"event=process_stop_req id=%s pid=%d grace_sec=%d",
		id, pid, graceSec))
}

/// ProcessBackoff 는 Back-off 재기동 대기 상태를 기록한다.
func ProcessBackoff(id string, failCount, delaySec int, nextRetry string) {
	logPush("WARN", fmt.Sprintf(
		"event=process_backoff id=%s fail_count=%d delay_sec=%d next_retry=%s",
		id, failCount, delaySec, nextRetry))
}

/// ProcessDisabled 는 MaxCount 초과로 프로세스가 비활성화됨을 기록한다.
func ProcessDisabled(id string, maxCount int) {
	logPush("ERROR", fmt.Sprintf(
		"event=process_disabled id=%s max_count=%d reason=운영자개입필요",
		id, maxCount))
}

/// ProcessWaitDepend 는 종속 프로세스 대기 시작을 기록한다.
func ProcessWaitDepend(id, waitFor string) {
	logPush("INFO", fmt.Sprintf(
		"event=process_waitdepend id=%s wait_for=%s",
		id, waitFor))
}

/// ProcessCascade 는 종속 프로세스 연쇄 처리를 기록한다.
func ProcessCascade(sourceID, targetID, action string) {
	logPush("WARN", fmt.Sprintf(
		"event=process_cascade source=%s target=%s action=%s",
		sourceID, targetID, action))
}

////////////////////////////////////////////////////////////////////////////////
// 모니터 이벤트
////////////////////////////////////////////////////////////////////////////////

/// MonitorConnect 는 모니터 클라이언트 접속을 기록한다.
func MonitorConnect(addr string) {
	logPush("INFO", fmt.Sprintf("event=monitor_connect addr=%s", addr))
}

/// MonitorCmd 는 모니터에서 수신한 제어 명령을 기록한다.
func MonitorCmd(cmd, id, result string) {
	logPush("INFO", fmt.Sprintf(
		"event=monitor_cmd cmd=%s id=%s result=%s",
		cmd, id, result))
}

////////////////////////////////////////////////////////////////////////////////
// 범용 로그 함수
// 위의 이벤트 함수로 표현하기 어려운 경우에만 사용한다.
////////////////////////////////////////////////////////////////////////////////

/// Info 는 일반 정보 로그를 기록한다.
func Info(format string, v ...interface{}) {
	logPush("INFO", fmt.Sprintf(format, v...))
}

/// Error 는 오류 로그를 기록한다.
func Error(format string, v ...interface{}) {
	logPush("ERROR", fmt.Sprintf(format, v...))
}

/// Warn 는 경고 로그를 기록한다.
func Warn(format string, v ...interface{}) {
	logPush("WARN", fmt.Sprintf(format, v...))
}

/// Debug 는 디버그 로그를 기록한다.
func Debug(format string, v ...interface{}) {
	logPush("DEBUG", fmt.Sprintf(format, v...))
}

////////////////////////////////////////////////////////////////////////////////
// 유틸리티
////////////////////////////////////////////////////////////////////////////////

/// StopReasonStr 은 StopReason 정수값을 로그용 문자열로 변환한다.
///
/// @param  reason  StopReason 정수값
/// @return         로그용 문자열
func StopReasonStr(reason int) string {
	switch reason {
	case  0: return "none"
	case -1: return "operator"
	case -2: return "holiday"
	case -3: return "restart"
	case -4: return "time_up"
	case -5: return "reload"
	case -6: return "abnormal"
	case -7: return "max_count"
	default: return fmt.Sprintf("unknown(%d)", reason)
	}
}
