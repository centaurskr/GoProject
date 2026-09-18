///
/// @file    goClsTbnova/main.go
/// @brief   goTbnova 관리 프로세스 정리 유틸리티
/// @details goTbnova 종료 후 독립 실행 중인 관리 대상 프로세스를 탐지하고 종료한다.
///
///          배경:
///            goTbnova 는 Setsid=true 로 자식을 독립 세션으로 기동한다.
///            goTbnova 종료 후 자식들은 ppid=1(init) 하위로 재배치되어 계속 실행된다.
///            이 유틸리티는 그 프로세스들을 process.toml 기준으로 탐지하여 정리한다.
///
///          탐지 기준 (AND 조건):
///            1. /proc/{pid}/status PPid == 1
///            2. /proc/{pid}/status Uid  == 대상 유저 UID (-u 옵션)
///            3. /proc/{pid}/exe 절대경로가 process.toml 의 path 와 일치
///               (절대경로 미일치 시 basename 보조 비교)
///
///          버그 수정 이력:
///            - basename 중복 ID 문제: managedByBase 를 []string 으로 관리
///            - killAll 실패 시 exit code 1 반환
///            - exe 빈 문자열 처리 명시
///            - ppid/uid 미파싱 시 명시적 오류 반환
///
///          사용법:
///            goClsTbnova [options] <table_path>
///
///          옵션:
///            -u <user>   대상 유저명 (기본값: 현재 유저)
///            -n          dry-run (목록만 출력, 종료 안함)
///            -f          강제 종료 (확인 없이 즉시 SIGKILL)
///            -s          SIGTERM 사용 (기본값: SIGKILL)
///            -v          상세 출력
///
///          주의:
///            goTbnova 가 실행 중일 때는 사용하지 말 것.
///            반드시 goTbnova 종료 후에 실행한다.
///
/// @date    2026-04-27
/// @author  Cento
///

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
)

////////////////////////////////////////////////////////////////////////////////
// 타입 정의
////////////////////////////////////////////////////////////////////////////////

/// procInfo 는 /proc 에서 읽은 프로세스 정보를 담는다.
type procInfo struct {
	pid     int      ///< 프로세스 ID
	ppid    int      ///< 부모 프로세스 ID
	uid     int      ///< 실행 유저 UID (real uid)
	ids     []string ///< 매칭된 process.toml ID 목록 (동일 실행파일 여러 인스턴스 대비)
	cmdline string   ///< 전체 실행 명령 (/proc/{pid}/cmdline, null 구분 → 공백 변환)
	exe     string   ///< 실행 파일 절대경로 (/proc/{pid}/exe 심볼릭링크)
}

/// tableProcess 는 process.toml 에서 읽은 프로세스 정의를 담는다.
/// goClsTbnova 는 최소한의 필드만 파싱한다.
type tableProcess struct {
	ID   string `toml:"id"`
	Path string `toml:"path"`
}

/// tableConfig 는 process.toml 최소 구조를 담는다.
type tableConfig struct {
	Default struct {
		DefDir string `toml:"def_dir"`
	} `toml:"default"`
	Process []tableProcess `toml:"process"`
}

/// managedPaths 는 process.toml 파싱 결과를 담는다.
/// 동일 실행파일을 여러 프로세스가 공유하는 경우(NAFTMAT-0~9 등)를 올바르게 처리한다.
type managedPaths struct {
	byPath map[string][]string ///< 절대경로  → ID 목록
	byBase map[string][]string ///< basename → ID 목록 (보조)
}

////////////////////////////////////////////////////////////////////////////////
// main
////////////////////////////////////////////////////////////////////////////////

func main() {
	var (
		optUser    string ///< -u: 대상 유저명
		optDryRun  bool   ///< -n: dry-run
		optForce   bool   ///< -f: 확인 없이 즉시 종료
		optSIGTERM bool   ///< -s: SIGKILL 대신 SIGTERM 사용
		optVerbose bool   ///< -v: 상세 출력
	)

	flag.StringVar(&optUser,  "u", "",    "대상 유저명 (기본값: 현재 유저)")
	flag.BoolVar(&optDryRun,  "n", false, "dry-run: 목록만 출력, 종료 안함")
	flag.BoolVar(&optForce,   "f", false, "확인 없이 즉시 종료")
	flag.BoolVar(&optSIGTERM, "s", false, "SIGKILL 대신 SIGTERM 발송")
	flag.BoolVar(&optVerbose, "v", false, "상세 출력 (cmdline 전체 표시)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
사용법:
  goClsTbnova [options] <table_path>

설명:
  goTbnova 종료 후 독립 실행 중인 관리 대상 프로세스를 탐지하고 종료한다.
  ppid=1 인 프로세스 중 process.toml 에 등록된 것만 대상으로 한다.

  ⚠️  goTbnova 가 실행 중일 때는 사용하지 말 것.

옵션:
  -u <user>   대상 유저명 (기본값: 현재 유저)
  -n          dry-run: 목록만 출력, 종료 안함
  -f          확인 없이 즉시 종료 (SIGKILL)
  -s          SIGKILL 대신 SIGTERM 발송
  -v          상세 출력 (cmdline 전체 표시)

예시:
  goClsTbnova    /home/cento/conf/process.toml          목록 확인 + 확인 후 종료
  goClsTbnova -n /home/cento/conf/process.toml          dry-run (목록만)
  goClsTbnova -f /home/cento/conf/process.toml          확인 없이 즉시 종료
  goClsTbnova -u cento -s /home/cento/conf/process.toml SIGTERM 발송

종료 코드:
  0  정상 (탐지 없음 포함)
  1  오류 (인자 오류, 파일 읽기 실패, 시그널 발송 실패)

`)
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	tablePath := flag.Arg(0)

	/// ── 1. 대상 유저 UID 확인 ────────────────────────────────────────────
	targetUID, err := resolveUID(optUser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 유저 확인 실패: %v\n", err)
		os.Exit(1)
	}

	/// ── 2. process.toml 파싱 → 관리 대상 경로 집합 ──────────────────────
	managed, err := loadManagedPaths(tablePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 설정 파일 읽기 실패: %v\n", err)
		os.Exit(1)
	}
	if len(managed.byPath) == 0 {
		fmt.Println("process.toml 에 프로세스가 없습니다.")
		os.Exit(0)
	}

	/// ── 3. /proc 스캔 → 대상 프로세스 탐지 ─────────────────────────────
	found, err := scanProc(targetUID, managed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] /proc 스캔 실패: %v\n", err)
		os.Exit(1)
	}

	/// ── 4. 결과 출력 ──────────────────────────────────────────────────────
	if len(found) == 0 {
		fmt.Println("정리할 프로세스가 없습니다.")
		os.Exit(0)
	}

	printFound(found, optVerbose)

	/// ── 5. dry-run 이면 종료 ──────────────────────────────────────────────
	if optDryRun {
		fmt.Printf("\n[dry-run] 실제 종료하지 않았습니다. -f 옵션 없이 실행하거나 -f 로 즉시 종료하세요.\n")
		os.Exit(0)
	}

	/// ── 6. 확인 프롬프트 ────────────────────────────────────────────────
	sig := syscall.SIGKILL
	sigName := "SIGKILL"
	if optSIGTERM {
		sig = syscall.SIGTERM
		sigName = "SIGTERM"
	}

	if !optForce {
		fmt.Printf("\n위 %d개 프로세스에 %s 를 발송합니다. 계속하시겠습니까? [y/N] ",
			len(found), sigName)
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.TrimSpace(strings.ToLower(ans))
		if ans != "y" && ans != "yes" {
			fmt.Println("취소되었습니다.")
			os.Exit(0)
		}
	}

	/// ── 7. 시그널 발송 ────────────────────────────────────────────────────
	if failed := killAll(found, sig, sigName, optVerbose); failed > 0 {
		os.Exit(1)
	}
}

////////////////////////////////////////////////////////////////////////////////
// 핵심 함수
////////////////////////////////////////////////////////////////////////////////

/// loadManagedPaths 는 process.toml 을 파싱하여 관리 대상 경로 집합을 반환한다.
///
/// 동일 실행파일을 여러 프로세스가 공유하는 경우(NAFTMAT-0~9 등)
/// byPath / byBase 가 ID 목록([]string) 을 담아 ID 덮어쓰기를 방지한다.
///
/// @param  tablePath  process.toml 전체 경로
/// @return            managedPaths (byPath: 절대경로→IDs, byBase: basename→IDs)
/// @return            에러
func loadManagedPaths(tablePath string) (managedPaths, error) {
	var cfg tableConfig
	if _, err := toml.DecodeFile(tablePath, &cfg); err != nil {
		return managedPaths{}, fmt.Errorf("TOML 파싱 실패: %w", err)
	}

	mp := managedPaths{
		byPath: make(map[string][]string),
		byBase: make(map[string][]string),
	}

	for _, proc := range cfg.Process {
		absPath := proc.Path
		if !filepath.IsAbs(absPath) {
			if cfg.Default.DefDir == "" {
				fmt.Fprintf(os.Stderr,
					"[WARN] [%s] 상대경로이나 def_dir 미설정 — 건너뜀\n", proc.ID)
				continue
			}
			absPath = filepath.Join(cfg.Default.DefDir, absPath)
		}

		/// 절대경로 → ID 등록 (중복 허용: append)
		mp.byPath[absPath] = append(mp.byPath[absPath], proc.ID)

		/// basename → ID 등록 (보조: 절대경로 미일치 시 사용)
		base := filepath.Base(absPath)
		mp.byBase[base] = append(mp.byBase[base], proc.ID)
	}
	return mp, nil
}

/// scanProc 는 /proc 를 스캔하여 탐지 기준에 맞는 프로세스 목록을 반환한다.
///
/// 탐지 기준 (모두 충족):
///   1. ppid == 1
///   2. uid  == targetUID
///   3. exe 절대경로가 managed.byPath 에 포함, 또는 basename 이 managed.byBase 에 포함
///
/// @param  targetUID  대상 유저 UID
/// @param  managed    loadManagedPaths() 반환값
/// @return            탐지된 procInfo 슬라이스 (PID 오름차순)
/// @return            에러
func scanProc(targetUID int, managed managedPaths) ([]procInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("/proc ReadDir 실패: %w", err)
	}

	var found []procInfo

	for _, entry := range entries {
		/// 숫자 디렉토리만 처리 (PID 디렉토리)
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		info, err := readProcInfo(pid)
		if err != nil {
			continue ///< 이미 종료된 프로세스 등 — 무시
		}

		/// 조건 1: ppid == 1
		if info.ppid != 1 {
			continue
		}

		/// 조건 2: uid 일치
		if info.uid != targetUID {
			continue
		}

		/// 조건 3: exe 빈 문자열 방어
		if info.exe == "" {
			continue
		}

		/// 조건 3-A: 절대경로 일치
		if ids, ok := managed.byPath[info.exe]; ok {
			info.ids = ids
			found = append(found, info)
			continue
		}

		/// 조건 3-B: basename 보조 비교
		/// (실행파일이 심볼릭링크 등으로 경로가 다를 경우 대비)
		if ids, ok := managed.byBase[filepath.Base(info.exe)]; ok {
			info.ids = ids
			found = append(found, info)
		}
	}

	/// PID 오름차순 정렬
	sort.Slice(found, func(i, j int) bool {
		return found[i].pid < found[j].pid
	})
	return found, nil
}

/// readProcInfo 는 /proc/{pid}/ 에서 프로세스 정보를 읽는다.
///
/// ppid 와 uid 가 파싱되지 않으면 에러를 반환한다.
///
/// @param  pid  프로세스 ID
/// @return      procInfo
/// @return      읽기 실패 또는 파싱 실패 시 에러
func readProcInfo(pid int) (procInfo, error) {
	info := procInfo{pid: pid}

	/// /proc/{pid}/status 파싱 → ppid, uid
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	f, err := os.Open(statusPath)
	if err != nil {
		return info, err
	}
	defer f.Close()

	parsedPPid, parsedUID := false, false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line   := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "PPid:":
			if v, err := strconv.Atoi(fields[1]); err == nil {
				info.ppid  = v
				parsedPPid = true
			}
		case "Uid:":
			/// Uid: real effective saved fs — real uid 사용
			if v, err := strconv.Atoi(fields[1]); err == nil {
				info.uid  = v
				parsedUID = true
			}
		}
		if parsedPPid && parsedUID {
			break ///< 두 필드 모두 파싱 완료 — 조기 종료
		}
	}

	if !parsedPPid || !parsedUID {
		return info, errors.New("ppid/uid 파싱 실패")
	}

	/// /proc/{pid}/exe 심볼릭링크 → 실행 파일 절대경로
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	if exe, err := os.Readlink(exePath); err == nil {
		info.exe = exe
	}

	/// /proc/{pid}/cmdline → 실행 명령 전체 (null 구분 → 공백 변환)
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	if data, err := os.ReadFile(cmdlinePath); err == nil {
		info.cmdline = strings.ReplaceAll(
			strings.TrimRight(string(data), "\x00"),
			"\x00", " ")
	}

	return info, nil
}

/// killAll 은 탐지된 프로세스에 시그널을 발송한다.
///
/// @param  procs    종료할 프로세스 목록
/// @param  sig      발송할 시그널
/// @param  sigName  시그널 이름 (로그용)
/// @param  verbose  상세 출력 여부
/// @return          실패 건수 (0 이면 전체 성공)
func killAll(procs []procInfo, sig syscall.Signal, sigName string, verbose bool) int {
	success, failed := 0, 0

	for _, p := range procs {
		idStr := strings.Join(p.ids, ",")
		err   := syscall.Kill(p.pid, sig)
		if err != nil {
			fmt.Printf("  [FAIL] PID %-8d  ID=%-20s  %s: %v\n",
				p.pid, idStr, sigName, err)
			failed++
		} else {
			fmt.Printf("  [ OK ] PID %-8d  ID=%-20s  %s 발송\n",
				p.pid, idStr, sigName)
			if verbose {
				fmt.Printf("         exe=%s\n", p.exe)
			}
			success++
		}
	}

	fmt.Printf("\n완료: 성공 %d / 실패 %d\n", success, failed)
	return failed
}

////////////////////////////////////////////////////////////////////////////////
// 출력 함수
////////////////////////////////////////////////////////////////////////////////

/// printFound 는 탐지된 프로세스 목록을 출력한다.
func printFound(procs []procInfo, verbose bool) {
	const sep = "────────────────────────────────────────────────────────────────────"
	fmt.Printf("\n탐지된 프로세스 %d개 (ppid=1, goTbnova 관리 대상):\n", len(procs))
	fmt.Println(sep)
	fmt.Printf("  %-8s  %-24s  %s\n", "PID", "ID(s)", "EXE")
	fmt.Println(sep)

	for _, p := range procs {
		idStr := strings.Join(p.ids, ", ")
		fmt.Printf("  %-8d  %-24s  %s\n", p.pid, idStr, p.exe)
		if verbose {
			fmt.Printf("  %8s  %-24s  cmdline: %s\n", "", "", p.cmdline)
		}
	}
	fmt.Println(sep)
}

////////////////////////////////////////////////////////////////////////////////
// 유틸리티
////////////////////////////////////////////////////////////////////////////////

/// resolveUID 는 유저명을 UID 로 변환한다.
/// 빈 문자열이면 현재 프로세스의 real UID 를 반환한다.
///
/// @param  username  유저명 (빈 문자열: 현재 유저)
/// @return           UID
/// @return           에러
func resolveUID(username string) (int, error) {
	if username == "" {
		return os.Getuid(), nil
	}
	u, err := user.Lookup(username)
	if err != nil {
		return 0, fmt.Errorf("유저 [%s] 를 찾을 수 없습니다: %w", username, err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, fmt.Errorf("UID 변환 실패: %w", err)
	}
	return uid, nil
}
