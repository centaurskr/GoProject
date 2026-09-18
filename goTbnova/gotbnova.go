///
/// @file    main.go
/// @brief   goTbnova 진입점
/// @details 명령행 인자를 파싱하고 각 컴포넌트를 조립하여 기동한다.
///
///          사용법:
///            goTbnova [options] <table_path>
///
///          옵션:
///            -t            설정 파일 검증만 수행 후 종료 (nginx -t 상당)
///            -d            설정 파일 덤프 후 종료
///
///          환경 설정:
///            $ENV_HOME/gotbnova.toml 에서 실행 환경을 로드한다.
///            파일 없으면 기본값으로 동작한다.
///
///          기동 순서:
///            1. 인자 파싱
///            2. -t/-d 옵션 처리 (해당 시 처리 후 종료)
///            3. tlog 초기화
///            4. Manager 생성 (설정 로드)
///            5. control/Server 생성 (TCP 포트 바인딩)
///               → 실패 시 C버전 또는 다른 goTbnova 가 실행 중
///            6. 시그널 핸들러 등록
///            7. Manager.Run() goroutine 기동
///            8. control/Server.Serve() goroutine 기동
///            9. 시그널 대기 (SIGTERM / SIGINT)
///           10. ctx 취소 → Graceful Shutdown
///           11. tlog.Close()
///
///          시그널 처리:
///            SIGTERM  → Graceful Shutdown (systemd stop)
///            SIGINT   → Graceful Shutdown (Ctrl+C)
///            SIGPIPE  → 무시 (TCP 연결 중단 시 발생)
///            SIGSEGV  → 로그 기록 후 종료
///
/// @date    2026-04-27
/// @author  Cento
///

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"goTbnova/config"
	"goTbnova/control"
	"goTbnova/process"
	"goTbnova/tlog"
)

////////////////////////////////////////////////////////////////////////////////
// 버전 정보
////////////////////////////////////////////////////////////////////////////////

const (
	appName    = "goTbnova"
	appVersion = "1.0.0"
)

////////////////////////////////////////////////////////////////////////////////
// main
////////////////////////////////////////////////////////////////////////////////

/// main 은 goTbnova 의 진입점이다.
func main() {

	/// ── 1. 명령행 인자 파싱 ───────────────────────────────────────────────

	var (
		optTest    bool   ///< -t:  process.toml 검증 후 종료
		optDump    bool   ///< -d:  process.toml 덤프 후 종료
		optTestEnv bool   ///< -te: gotbnova.toml 검증 후 종료
		optVersion bool   ///< -v:  버전 출력
	)

	flag.BoolVar(&optTest,    "t",  false, "process.toml 검증만 수행 후 종료 (nginx -t 상당)")
	flag.BoolVar(&optDump,    "d",  false, "process.toml 파싱 결과를 표준 출력으로 덤프 후 종료")
	flag.BoolVar(&optTestEnv, "te", false, "gotbnova.toml 검증 및 내용 출력 후 종료")
	flag.BoolVar(&optVersion, "v",  false, "버전 정보 출력")

	flag.Usage = usage
	flag.Parse()

	/// ── 2. -v: 버전 출력 ──────────────────────────────────────────────────
	if optVersion {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	/// ── 3. -te: gotbnova.toml 검증 후 종료 ────────────────────────────
	/// process.toml 인자 없이 단독 실행 가능하므로 table_path 검증 전에 수행
	if optTestEnv {
		testEnvFile()
		/// testEnvFile 내부에서 os.Exit() 호출
	}

	/// table_path 인자 확인 (-te 이외 모든 명령에 필수)
	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}
	tablePath := flag.Arg(0)

	/// table_path 는 절대경로 필수
	if len(tablePath) == 0 || tablePath[0] != '/' {
		fmt.Fprintf(os.Stderr, "[ERROR] table_path 는 절대경로여야 합니다: %s\n", tablePath)
		os.Exit(1)
	}

	/// ── 4. -t: process.toml 검증 후 종료 ───────────────────────────────
	/// tlog 초기화 없이 수행 (검증은 독립적으로 동작)
	if optTest {
		config.TestConfig(tablePath)
		/// TestConfig 내부에서 os.Exit() 호출
	}

	/// ── 5. -d: process.toml 덤프 후 종료 ──────────────────────────────
	if optDump {
		config.DumpConfig(tablePath)
		/// DumpConfig 내부에서 os.Exit() 호출
	}

	/// ── 6. gotbnova.toml 환경 파일 로드 ──────────────────────────────────
	/// $ENV_HOME/gotbnova.toml → 실행파일경로/gotbnova.toml 순으로 탐색
	/// 파일 없으면 기본값으로 동작
	envCfg, envPath, err := config.LoadEnvFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] gotbnova.toml 로드 실패: %v\n", err)
		os.Exit(1)
	}
	if envPath != "" {
		fmt.Fprintf(os.Stderr, "[INFO] 환경 파일 로드: %s\n", envPath)
	} else {
		fmt.Fprintf(os.Stderr, "[INFO] gotbnova.toml 없음 — 기본값으로 동작\n")
	}

	/// ── 7. tlog 초기화 ─────────────────────────────────────────────────
	/// 로그 설정은 gotbnova.toml [log] 섹션에서 로드
	logMode := tlog.LogMode(envCfg.Log.Type)
	logDir  := resolveLogDir(logMode, envCfg.Log.Dir)

	if logMode == tlog.ModeFile && logDir == "" {
		fmt.Fprintf(os.Stderr,
			"[ERROR] log.type=file 이지만 로그 디렉토리를 결정할 수 없습니다.\n"+
			"        gotbnova.toml [log] dir 을 설정하세요.\n")
		os.Exit(1)
	}
	if err := tlog.InitWithSocket(logMode, tablePath, logDir, envCfg.Log.VectorSocket); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 로그 초기화 실패:\n  %v\n", err)
		os.Exit(1)
	}
	defer tlog.Close()

	/// ── 8. Manager 생성 (설정 로드 + 검증) ───────────────────────────────
	mgr, err := process.NewManager(tablePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 초기화 실패: %v\n", err)
		tlog.Error("event=startup_failed reason=%s", err.Error())
		os.Exit(1)
	}

	/// ── 9. control/Server 생성 (TCP 포트 바인딩) ─────────────────────────
	/// 바인딩 IP 는 gotbnova.toml [server] bind 에서 로드
	if envCfg.Server.Bind != "127.0.0.1" {
		fmt.Fprintf(os.Stderr,
			"[WARN] 모니터 서버를 외부 IP [%s] 에 바인딩합니다.\n"+
			"       신뢰할 수 있는 네트워크에서만 사용하세요.\n", envCfg.Server.Bind)
		tlog.Warn("event=bind_external addr=%s reason=외부IP바인딩_보안주의", envCfg.Server.Bind)
	}
	srv, err := control.NewServer(mgr.ControlCh(), envCfg.Server.Bind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 모니터 서버 시작 실패: %v\n", err)
		tlog.Error("event=startup_failed reason=%s", err.Error())
		os.Exit(1)
	}

	/// ── 9. 시그널 핸들러 등록 ─────────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGTERM, ///< systemd stop
		syscall.SIGINT,  ///< Ctrl+C
		syscall.SIGSEGV, ///< 세그멘테이션 폴트
	)
	/// SIGPIPE 는 무시 (TCP 연결 중단 시 발생)
	signal.Ignore(syscall.SIGPIPE)

	/// work 디렉토리 변경 (gotbnova.toml [work] dir)
	if envCfg.Work.Dir != "" {
		if err := os.Chdir(envCfg.Work.Dir); err != nil {
			tlog.Warn("event=chdir_failed dir=%s err=%s", envCfg.Work.Dir, err.Error())
		} else {
			tlog.Info("event=chdir dir=%s", envCfg.Work.Dir)
		}
	}

	tlog.Startup(tablePath, srv.Port())
	fmt.Printf("%s %s 기동 완료 — 테이블: %s\n", appName, appVersion, tablePath)

	/// ── 10. 컴포넌트 goroutine 기동 ──────────────────────────────────────
	/// Manager 메인 루프
	go mgr.Run(ctx)

	/// TCP 모니터 서버
	go srv.Serve(ctx)

	/// ── 11. 시그널 대기 ───────────────────────────────────────────────────
	sig := <-sigCh
	switch sig {
	case syscall.SIGSEGV:
		tlog.Error("event=sigsegv reason=세그멘테이션폴트")
		os.Exit(1)
	default:
		tlog.Shutdown(sig.String())
		fmt.Printf("\n%s 종료 중 (signal: %s) ...\n", appName, sig)
	}

	/// ── 12. Graceful Shutdown ─────────────────────────────────────────────
	/// ctx 취소 → Manager.Run() 과 Server.Serve() 가 종료됨
	cancel()

	/// tlog.Close() 는 defer 로 처리 (미전송 로그 flush)
	fmt.Printf("%s 종료 완료\n", appName)
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수
////////////////////////////////////////////////////////////////////////////////

/// usage 는 사용법을 출력한다.
/// flag.Usage 에 등록된다.
func usage() {
	fmt.Fprintf(os.Stderr, `
사용법:
  %s [options] <table_path>

인자:
  table_path    TOML 설정 파일 전체 경로 (절대경로 필수)

옵션:
  -t            process.toml 검증만 수행 후 종료 (nginx -t 상당)
                오류 시 파일명, 라인 번호, 오류 원인을 출력합니다.
  -d            process.toml 파싱 결과를 표준 출력으로 덤프 후 종료
                $VAR 환경변수 치환 결과도 함께 출력합니다.
  -te           gotbnova.toml 검증 및 내용 출력 후 종료
                $ENV_HOME/gotbnova.toml 을 탐색하여 파싱 결과를 출력합니다.
                process.toml 인자 없이 단독 실행 가능합니다.
  -v            버전 정보 출력

환경 설정 ($ENV_HOME/gotbnova.toml):
  [server] bind          모니터 바인딩 IP      (기본: 127.0.0.1)
  [log]    type          로그 백엔드            (기본: net)
  [log]    vector_socket Vector UDS 소켓 경로   (기본: /tmp/VECTOR.UNIX.DGRAM)
  [log]    dir           파일 로그 디렉토리     (type=file 시)
  [work]   dir           작업 디렉토리

예시:
  export ENV_HOME=/home/cento/Project/env
  %s /home/cento/conf/process.toml        정상 기동
  %s -t  /home/cento/conf/process.toml    process.toml 검증
  %s -d  /home/cento/conf/process.toml    process.toml 덤프
  %s -te                                   gotbnova.toml 검증

`, appName, appName, appName, appName, appName)
}

/// testEnvFile 은 gotbnova.toml 을 탐색·파싱하고 결과를 출력한 뒤 종료한다.
/// -te 옵션에서 호출된다. process.toml 인자 없이 단독 실행 가능하다.
///
/// 출력 내용:
///   - 탐색 경로 ($ENV_HOME, 실행파일 디렉토리)
///   - 파일 발견 여부
///   - 파싱 성공/실패
///   - 각 섹션의 최종 적용값
func testEnvFile() {
	const sep = "────────────────────────────────────────────────────────"
	fmt.Printf("goTbnova: gotbnova.toml 검증\n")
	fmt.Println(sep)

	/// 탐색 경로 표시
	envHome := os.Getenv("ENV_HOME")
	if envHome != "" {
		fmt.Printf("  탐색 1: %s/gotbnova.toml\n", envHome)
	} else {
		fmt.Println("  탐색 1: $ENV_HOME 미설정 — 건너뜀")
	}
	if exe, err := os.Executable(); err == nil {
		fmt.Printf("  탐색 2: %s/gotbnova.toml\n", filepath.Dir(exe))
	}

	/// 로드 및 파싱
	ef, path, err := config.LoadEnvFile()
	if err != nil {
		fmt.Printf("\ngoTbnova: [FAIL] gotbnova.toml 파싱 오류\n  %v\n", err)
		os.Exit(1)
	}

	if path == "" {
		fmt.Println("\n  결과: 파일 없음 — 기본값으로 동작합니다.")
	} else {
		fmt.Printf("\n  결과: 로드 성공 [%s]\n", path)
	}

	/// 설정값 출력
	fmt.Println(sep)
	config.DumpEnvFile(ef, path)
	fmt.Println(sep)
	fmt.Println("goTbnova: gotbnova.toml 검증 OK")
	os.Exit(0)
}

/// resolveLogDir 은 log.type=file 모드 시 로그 파일 저장 디렉토리를 결정한다.
///
/// 탐색 순서:
///   1. gotbnova.toml [log] dir 설정값   (최우선)
///   2. 실행 파일 디렉토리/log            (폴백)
///   3. 빈 문자열 반환                    (호출자가 오류 처리)
///
/// Note:
///   $LOG_HOME 환경변수는 참조하지 않는다.
///   로그 디렉토리는 gotbnova.toml [log] dir 에서 관리한다.
///
/// @param  mode    로그 모드 (ModeFile 이 아니면 빈 문자열 반환)
/// @param  logdir  gotbnova.toml [log] dir 값 (빈 문자열이면 폴백 탐색)
/// @return         로그 디렉토리 절대경로 (결정 불가 시 빈 문자열)
func resolveLogDir(mode tlog.LogMode, logdir string) string {
	if mode != tlog.ModeFile {
		return ""
	}

	/// 1순위: gotbnova.toml [log] dir 설정값
	if logdir != "" {
		return logdir
	}

	/// 2순위: 실행 파일과 동일 디렉토리의 log/ 서브디렉토리
	exe, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(exe), "log")
	}

	/// 결정 불가 → 호출자가 오류 처리
	return ""
}
