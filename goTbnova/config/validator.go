///
/// @file    config/validator.go
/// @brief   TOML 설정 파일 검증 및 덤프
/// @details Load() 에서 TOML 파싱 직후 호출된다.
///          오류 발견 시 즉시 반환하여 잘못된 설정으로 기동되는 것을 방지한다.
///
///          검증 항목:
///            1. 필수 필드 존재 여부 (ID, path, mode)
///            2. 허용값 범위 (mode, stmode, policy, cascade)
///            3. 시각 형식 (HH:MM:SS)
///            4. ID 중복 여부
///            5. 실행 파일 존재 여부
///            6. depends_on 순환 참조 여부
///
///          옵션:
///            -t <table_path>  nginx -t 상당, 설정 검증만 수행 후 종료
///                             오류 시 파일명, 라인 번호, 오류 원인을 출력한다.
///            -d <table_path>  설정 파일을 파싱하여 표준 출력으로 덤프 후 종료
///
/// @date    2026-04-27
/// @author  Cento
///

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"goTbnova/model"
)

////////////////////////////////////////////////////////////////////////////////
// 타입 정의
////////////////////////////////////////////////////////////////////////////////

/// ValidationError 는 설정 파일 검증 오류 정보를 담는다.
/// 파일명, 라인 번호, 필드명, 오류 원인을 포함한다.
type ValidationError struct {
	File   string ///< 설정 파일 경로
	Line   int    ///< 오류 발생 라인 번호 (0: 라인 특정 불가)
	Field  string ///< 오류 발생 필드명
	Reason string ///< 오류 원인 설명
}

/// Error 는 ValidationError 를 에러 인터페이스로 구현한다.
func (e *ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d [%s] %s", e.File, e.Line, e.Field, e.Reason)
	}
	return fmt.Sprintf("%s [%s] %s", e.File, e.Field, e.Reason)
}

/// lineIndex 는 설정 파일의 라인 번호 인덱스를 담는다.
/// 프로세스 ID → [[process]] 섹션 시작 라인 번호 맵
type lineIndex map[string]int

////////////////////////////////////////////////////////////////////////////////
// 공개 함수 — 설정 파일 테스트 (nginx -t 상당)
////////////////////////////////////////////////////////////////////////////////

/// TestConfig 는 설정 파일을 검증하고 결과를 표준 출력으로 출력한 뒤 종료한다.
/// 성공 시 os.Exit(0), 실패 시 os.Exit(1).
/// goTbnova 를 실제로 기동하지 않는다.
///
/// 출력 예시 (성공):
///   goTbnova: TOML 문법 검사 OK [/etc/goTbnova/process.toml]
///   goTbnova: 버전: 1.0.0 / 전역env: 2개 / 그룹: 1개 / 프로세스: 3개
///   goTbnova:   [ OK ] L.081  AGENT       P  00:00:00~23:59:59
///   goTbnova:   [ OK ] L.095  DB_AGENT    R  02:30:00~23:55:55  (depends: AGENT)
///   goTbnova:   [ OK ] L.112  APP_SERVER  R  02:35:00~23:55:55  (depends: DB_AGENT)
///   goTbnova: depends_on 순환 참조 없음
///   goTbnova: 설정 검증 OK
///
/// 출력 예시 (실패):
///   goTbnova: TOML 문법 검사 OK [/etc/goTbnova/process.toml]
///   goTbnova: 버전: 1.0.0 / 전역env: 2개 / 그룹: 1개 / 프로세스: 3개
///   goTbnova:   [ OK ] L.081  AGENT       P  00:00:00~23:59:59
///   goTbnova:   [ERR] L.095  DB_AGENT    R  02:30:00~23:55:55
///              └─ /etc/goTbnova/process.toml:98 [path] 실행 파일이 없습니다 [/home/cento/bin/DB_Agent]
///   goTbnova:   [ OK ] L.112  APP_SERVER  R  02:35:00~23:55:55
///   goTbnova: depends_on 순환 참조 없음
///   goTbnova: 설정 검증 FAILED (1개 오류)
///
/// @param  path  TOML 설정 파일 전체 경로
func TestConfig(path string) {
	const prefix = "goTbnova: "
	errCount := 0

	/// Step 1: TOML 문법 파싱
	var cfg model.TableConfig
	md, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		/// BurntSushi/toml 은 문법 오류 시 라인/컬럼 정보를 포함한 메시지를 반환한다
		fmt.Fprintf(os.Stderr, "%s[ERR] TOML 문법 오류:\n", prefix)
		fmt.Fprintf(os.Stderr, "       └─ %s: %v\n", path, err)
		fmt.Fprintf(os.Stderr, "%s설정 검증 FAILED\n", prefix)
		os.Exit(1)
	}
	_ = md
	fmt.Printf("%sTOML 문법 검사 OK [%s]\n", prefix, path)

	/// Step 2: 프로세스 섹션 라인 번호 사전 스캔
	lidx, scanErr := buildLineIndex(path)
	if scanErr != nil {
		fmt.Fprintf(os.Stderr, "%s[WARN] 라인 번호 스캔 실패: %v\n", prefix, scanErr)
	}

	/// Step 3: 요약 출력
	fmt.Printf("%s버전: %s / 전역env: %d개 / 그룹: %d개 / 프로세스: %d개\n",
		prefix,
		mapGet(cfg.Version, "version", "(없음)"),
		len(cfg.Env),
		len(cfg.Groups),
		len(cfg.Process),
	)

	/// Step 4: 프로세스별 검증 및 출력
	seen := make(map[string]bool, len(cfg.Process))
	for _, p := range cfg.Process {
		lineNo := lidx[p.ID]
		verrs  := validateProcessFull(path, &p, cfg.Default.DefDir, seen, lineNo)

		/// 상태 태그
		tag := " OK"
		if len(verrs) > 0 {
			tag = "ERR"
			errCount += len(verrs)
		}

		/// depends_on 표시
		depStr := ""
		if len(p.DependsOn) > 0 {
			depStr = "  (depends: " + strings.Join(p.DependsOn, ", ") + ")"
		}

		/// 라인 번호 표시
		lineStr := "L.???"
		if lineNo > 0 {
			lineStr = fmt.Sprintf("L.%03d", lineNo)
		}

		fmt.Printf("%s  [%s] %s  %-13s %-2s  %s~%s%s\n",
			prefix, tag, lineStr,
			p.ID, string(p.Mode),
			p.St, p.Et,
			depStr,
		)

		/// 오류 상세 출력
		for _, ve := range verrs {
			if ve.Line > 0 {
				fmt.Fprintf(os.Stderr, "             └─ %s:%d [%s] %s\n",
					ve.File, ve.Line, ve.Field, ve.Reason)
			} else {
				fmt.Fprintf(os.Stderr, "             └─ %s [%s] %s\n",
					ve.File, ve.Field, ve.Reason)
			}
		}
	}

	/// Step 5: depends_on 순환 참조 검사
	if err := checkCircularDepends(cfg.Process); err != nil {
		fmt.Fprintf(os.Stderr, "%s[ERR] %v\n", prefix, err)
		errCount++
	} else {
		fmt.Printf("%sdepends_on 순환 참조 없음\n", prefix)
	}

	/// Step 6: 최종 결과
	if errCount > 0 {
		fmt.Fprintf(os.Stderr, "%s설정 검증 FAILED (%d개 오류)\n", prefix, errCount)
		os.Exit(1)
	}
	fmt.Printf("%s설정 검증 OK\n", prefix)
	os.Exit(0)
}

////////////////////////////////////////////////////////////////////////////////
// 공개 함수 — 설정 파일 덤프 (-d 옵션)
////////////////////////////////////////////////////////////////////////////////

/// DumpConfig 는 TOML 설정 파일을 파싱하여 표준 출력으로 덤프한 뒤 종료한다.
/// 환경변수 치환($VAR) 결과까지 반영하여 실제 적용값을 확인할 수 있다.
/// 성공 시 os.Exit(0), 파싱 실패 시 os.Exit(1).
///
/// @param  path  TOML 설정 파일 전체 경로
func DumpConfig(path string) {
	const sep  = "============================================================"
	const line = "────────────────────────────────────────────────────────────"

	var cfg model.TableConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "TOML 파싱 오류: %v\n", err)
		os.Exit(1)
	}

	globalEnv := expandGlobalEnv(cfg.Env)
	applyDefaults(&cfg)

	/// 라인 인덱스 (덤프에서도 표시)
	lidx, _ := buildLineIndex(path)

	fmt.Println(sep)
	fmt.Printf("goTbnova 설정 덤프 [%s]\n", path)
	fmt.Println(sep)

	fmt.Println("\n[버전]")
	if v, ok := cfg.Version["version"]; ok {
		fmt.Printf("  version : %s\n", v)
	} else {
		fmt.Println("  (없음)")
	}

	fmt.Println("\n[기본값]")
	fmt.Printf("  def_dir : %s\n", cfg.Default.DefDir)
	fmt.Printf("  def_cnt : %d\n", cfg.Default.DefCnt)

	fmt.Printf("\n[전역 환경변수] (%d개) — $VAR 치환 후 값\n", len(globalEnv))
	if len(globalEnv) == 0 {
		fmt.Println("  (없음)")
	} else {
		for k, v := range globalEnv {
			fmt.Printf("  %-16s = %s\n", k, v)
		}
	}

	fmt.Printf("\n[그룹] (%d개)\n", len(cfg.Groups))
	if len(cfg.Groups) == 0 {
		fmt.Println("  (없음)")
	} else {
		for name, g := range cfg.Groups {
			fmt.Printf("  %-14s : %s\n", name, g.Desc)
		}
	}

	fmt.Printf("\n[프로세스] (%d개)\n", len(cfg.Process))
	for i, p := range cfg.Process {
		fmt.Println(line)

		absPath := p.Path
		if !filepath.IsAbs(absPath) && cfg.Default.DefDir != "" {
			absPath = filepath.Join(cfg.Default.DefDir, absPath)
		}
		procEnv  := expandProcessEnv(p.Env, globalEnv)
		weekendStr := "가동"
		if p.Weekend { weekendStr = "비가동 (토/일)" }
		stmodeDesc := "Q (즉시)"
		if p.StMode == model.StartSlow { stmodeDesc = "S (지연 2초)" }
		dependsStr := "(없음)"
		if len(p.DependsOn) > 0 {
			dependsStr = strings.Join(p.DependsOn, ", ")
			if p.DependsHealthy { dependsStr += " [안정 확인 후 기동]" }
		}
		argsStr := "(없음)"
		if len(p.Args) > 0 { argsStr = strings.Join(p.Args, " ") }
		backoffStr := "false"
		if p.Restart.Backoff { backoffStr = "true" }
		restartStr := fmt.Sprintf("%s / backoff=%s / max=%d / grace=%ds",
			p.Restart.Policy, backoffStr, p.Restart.MaxCount, p.Restart.GracePeriod)

		lineStr := "???"
		if ln := lidx[p.ID]; ln > 0 { lineStr = fmt.Sprintf("%d", ln) }

		fmt.Printf("No.%-3d  Line    : %s\n",    i+1, lineStr)
		fmt.Printf("        ID      : %s\n",     p.ID)
		fmt.Printf("        Group   : %s\n",     groupOrDefault(p.Group))
		fmt.Printf("        Mode    : %s\n",     modeDescription(p.Mode))
		fmt.Printf("        Weekend : %s\n",     weekendStr)
		fmt.Printf("        StMode  : %s\n",     stmodeDesc)
		fmt.Printf("        Time    : %s ~ %s\n", p.St, p.Et)
		fmt.Printf("        Path    : %s\n",     absPath)
		fmt.Printf("        Args    : %s\n",     argsStr)
		fmt.Printf("        Depends : %s\n",     dependsStr)
		fmt.Printf("        Restart : %s\n",     restartStr)
		fmt.Printf("        Cascade : %s\n",     p.Cascade.OnDeath)
		if len(procEnv) == 0 {
			fmt.Println("        Env     : (없음)")
		} else {
			first := true
			for k, v := range procEnv {
				if first {
					fmt.Printf("        Env     : %s=%s\n", k, v)
					first = false
				} else {
					fmt.Printf("                : %s=%s\n", k, v)
				}
			}
		}
		fmt.Printf("        Desc    : %s\n", p.Desc)
	}
	fmt.Println(sep)
	os.Exit(0)
}

////////////////////////////////////////////////////////////////////////////////
// 공개 함수 — 설정 검증 (Load 내부 호출용)
////////////////////////////////////////////////////////////////////////////////

/// Validate 는 TableConfig 의 내용을 검증한다.
/// 첫 번째 오류 발견 시 즉시 반환한다.
///
/// @param  cfg  파싱 완료된 TableConfig
/// @return      검증 실패 시 에러 (성공 시 nil)
func Validate(cfg *model.TableConfig) error {
	if len(cfg.Process) == 0 {
		return fmt.Errorf("[process] 섹션이 비어 있습니다")
	}
	if err := checkDuplicateID(cfg.Process); err != nil {
		return err
	}
	for _, p := range cfg.Process {
		if err := validateProcess(&p, cfg.Default.DefDir); err != nil {
			return err
		}
	}
	if err := checkCircularDepends(cfg.Process); err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 라인 번호 인덱스 구성
////////////////////////////////////////////////////////////////////////////////

/// buildLineIndex 는 TOML 파일을 라인 단위로 스캔하여
/// 프로세스 ID → [[process]] 섹션 내 id 필드 라인 번호 맵을 구성한다.
///
/// 탐지 방식:
///   [[process]] 마커를 만나면 섹션 시작으로 간주한다.
///   이후 'id = "XXXX"' 패턴을 만나면 해당 ID 의 라인 번호로 등록한다.
///
/// @param  path  TOML 파일 경로
/// @return       ID → 라인 번호 맵
/// @return       파일 읽기 실패 시 에러
func buildLineIndex(path string) (lineIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	idx     := make(lineIndex)
	lineNo  := 0
	inProc  := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNo++
		raw  := scanner.Text()
		text := strings.TrimSpace(raw)

		/// 주석 및 빈 줄 스킵
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		/// [[process]] 섹션 시작 탐지
		if text == "[[process]]" {
			inProc = true
			continue
		}

		/// 다른 섹션 시작 시 프로세스 모드 해제
		if strings.HasPrefix(text, "[") && !strings.HasPrefix(text, "[[process]]") {
			inProc = false
			continue
		}

		/// 프로세스 섹션 내에서 id 필드 탐지
		/// 형식: id = "AGENT" 또는 id = 'AGENT'
		if inProc && strings.HasPrefix(text, "id") {
			parts := strings.SplitN(text, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				if key == "id" {
					val := strings.TrimSpace(parts[1])
					val  = strings.Trim(val, `"'`)
					val  = strings.TrimSpace(val)
					if val != "" {
						idx[val] = lineNo
					}
				}
			}
		}
	}
	return idx, scanner.Err()
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 상세 검증 (TestConfig 용)
////////////////////////////////////////////////////////////////////////////////

/// validateProcessFull 은 프로세스 하나를 전체 검증하고
/// 발견된 모든 ValidationError 를 반환한다.
/// 첫 번째 오류에서 중단하지 않고 모든 오류를 수집한다.
///
/// @param  file    설정 파일 경로 (오류 메시지용)
/// @param  p       검증할 ProcessConfig
/// @param  defDir  기본 실행 경로
/// @param  seen    ID 중복 검사용 맵
/// @param  lineNo  [[process]] id 필드 라인 번호
/// @return         ValidationError 슬라이스 (없으면 nil)
func validateProcessFull(
	file   string,
	p      *model.ProcessConfig,
	defDir string,
	seen   map[string]bool,
	lineNo int,
) []*ValidationError {

	var errs []*ValidationError

	newErr := func(field, reason string) *ValidationError {
		return &ValidationError{File: file, Line: lineNo, Field: field, Reason: reason}
	}

	/// ID 길이
	if len(p.ID) > 13 {
		errs = append(errs, newErr("id",
			fmt.Sprintf("ID 는 13자 이내여야 합니다 (현재 %d자)", len(p.ID))))
	}

	/// ID 중복
	if seen[p.ID] {
		errs = append(errs, newErr("id",
			fmt.Sprintf("ID 중복: [%s]", p.ID)))
	}
	seen[p.ID] = true

	/// path 비어있음
	if strings.TrimSpace(p.Path) == "" {
		errs = append(errs, newErr("path", "path 가 비어 있습니다"))
	} else {
		/// 실행 파일 존재 확인
		absPath := p.Path
		if !filepath.IsAbs(absPath) {
			if defDir == "" {
				errs = append(errs, newErr("path",
					"상대경로 사용 시 [default] def_dir 이 필요합니다"))
			} else {
				absPath = filepath.Join(defDir, absPath)
			}
		}
		if absPath != "" {
			if info, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
				errs = append(errs, newErr("path",
					fmt.Sprintf("실행 파일이 없습니다 → %s", absPath)))
			} else if statErr == nil {
				/// 실행 권한 확인
				if info.Mode()&0111 == 0 {
					errs = append(errs, newErr("path",
						fmt.Sprintf("실행 권한이 없습니다 → %s (mode: %s)",
							absPath, info.Mode())))
				}
			}
		}
	}

	/// mode 허용값
	switch p.Mode {
	case model.ModePermanent, model.ModeRestricted, model.ModeOnce:
	default:
		errs = append(errs, newErr("mode",
			fmt.Sprintf("P/R/O 중 하나여야 합니다 (현재: %q)", p.Mode)))
	}

	/// R, O 모드 시각 검증
	if p.Mode == model.ModeRestricted || p.Mode == model.ModeOnce {
		if err := validateHMS(p.St); err != nil {
			errs = append(errs, newErr("st", err.Error()))
		}
		if err := validateHMS(p.Et); err != nil {
			errs = append(errs, newErr("et", err.Error()))
		}
	}

	/// stmode 허용값
	if p.StMode != "" {
		switch p.StMode {
		case model.StartQuick, model.StartSlow:
		default:
			errs = append(errs, newErr("stmode",
				fmt.Sprintf("Q/S 중 하나여야 합니다 (현재: %q)", p.StMode)))
		}
	}

	/// restart.policy 허용값
	if p.Restart.Policy != "" {
		switch p.Restart.Policy {
		case model.PolicyAlways, model.PolicyOnFailure, model.PolicyNever:
		default:
			errs = append(errs, newErr("restart.policy",
				fmt.Sprintf("always/on-failure/never 중 하나여야 합니다 (현재: %q)", p.Restart.Policy)))
		}
	}

	/// restart.grace_period 범위
	if p.Restart.GracePeriod < 0 {
		errs = append(errs, newErr("restart.grace_period",
			fmt.Sprintf("0 이상이어야 합니다 (현재: %d)", p.Restart.GracePeriod)))
	}

	/// restart.max_count 범위
	if p.Restart.MaxCount < 0 {
		errs = append(errs, newErr("restart.max_count",
			fmt.Sprintf("0 이상이어야 합니다 (현재: %d)", p.Restart.MaxCount)))
	}

	/// cascade.on_death 허용값
	if p.Cascade.OnDeath != "" {
		switch p.Cascade.OnDeath {
		case model.CascadeNone, model.CascadeNotify, model.CascadeStopDependents:
		default:
			errs = append(errs, newErr("cascade.on_death",
				fmt.Sprintf("none/notify/stop_dependents 중 하나여야 합니다 (현재: %q)",
					p.Cascade.OnDeath)))
		}
	}

	return errs
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 공통 검증 (Validate 용)
////////////////////////////////////////////////////////////////////////////////

/// checkDuplicateID 는 프로세스 ID 중복을 검사한다.
func checkDuplicateID(processes []model.ProcessConfig) error {
	seen := make(map[string]bool, len(processes))
	for _, p := range processes {
		id := strings.TrimSpace(p.ID)
		if id == "" {
			return fmt.Errorf("ID 가 비어 있는 프로세스가 있습니다")
		}
		if seen[id] {
			return fmt.Errorf("프로세스 ID 중복: [%s]", id)
		}
		seen[id] = true
	}
	return nil
}

/// validateProcess 는 프로세스 하나의 설정을 검증한다. (Load 내부 호출용)
func validateProcess(p *model.ProcessConfig, defDir string) error {
	id := p.ID
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("ID 가 비어 있습니다")
	}
	if len(id) > 13 {
		return fmt.Errorf("프로세스 [%s]: ID 는 13자 이내여야 합니다", id)
	}
	if strings.TrimSpace(p.Path) == "" {
		return fmt.Errorf("프로세스 [%s]: path 가 비어 있습니다", id)
	}
	absPath := p.Path
	if !filepath.IsAbs(absPath) {
		if defDir == "" {
			return fmt.Errorf("프로세스 [%s]: 상대경로 사용 시 def_dir 이 필요합니다", id)
		}
		absPath = filepath.Join(defDir, absPath)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("프로세스 [%s]: 실행 파일이 없습니다 [%s]", id, absPath)
	}
	switch p.Mode {
	case model.ModePermanent, model.ModeRestricted, model.ModeOnce:
	default:
		return fmt.Errorf("프로세스 [%s]: mode 는 P/R/O 중 하나여야 합니다 (현재: %s)", id, p.Mode)
	}
	if p.Mode == model.ModeRestricted || p.Mode == model.ModeOnce {
		if err := validateHMS(p.St); err != nil {
			return fmt.Errorf("프로세스 [%s]: st %w", id, err)
		}
		if err := validateHMS(p.Et); err != nil {
			return fmt.Errorf("프로세스 [%s]: et %w", id, err)
		}
	}
	if p.StMode != "" {
		switch p.StMode {
		case model.StartQuick, model.StartSlow:
		default:
			return fmt.Errorf("프로세스 [%s]: stmode 는 Q/S 중 하나여야 합니다 (현재: %s)", id, p.StMode)
		}
	}
	if p.Restart.Policy != "" {
		switch p.Restart.Policy {
		case model.PolicyAlways, model.PolicyOnFailure, model.PolicyNever:
		default:
			return fmt.Errorf("프로세스 [%s]: restart.policy 오류 (현재: %s)", id, p.Restart.Policy)
		}
	}
	if p.Cascade.OnDeath != "" {
		switch p.Cascade.OnDeath {
		case model.CascadeNone, model.CascadeNotify, model.CascadeStopDependents:
		default:
			return fmt.Errorf("프로세스 [%s]: cascade.on_death 오류 (현재: %s)", id, p.Cascade.OnDeath)
		}
	}
	if p.Restart.GracePeriod < 0 {
		return fmt.Errorf("프로세스 [%s]: grace_period 는 0 이상이어야 합니다", id)
	}
	if p.Restart.MaxCount < 0 {
		return fmt.Errorf("프로세스 [%s]: max_count 는 0 이상이어야 합니다", id)
	}
	return nil
}

/// validateHMS 는 "HH:MM:SS" 형식의 시각 문자열을 검증한다.
func validateHMS(s string) error {
	var hh, mm, ss int
	n, err := fmt.Sscanf(s, "%02d:%02d:%02d", &hh, &mm, &ss)
	if err != nil || n != 3 {
		return fmt.Errorf("형식 오류 — HH:MM:SS 형식 필요 (현재: %q)", s)
	}
	if hh < 0 || hh > 23 {
		return fmt.Errorf("시(HH) 범위 오류 — 0~23 이어야 합니다 (현재: %d)", hh)
	}
	if mm < 0 || mm > 59 {
		return fmt.Errorf("분(MM) 범위 오류 — 0~59 이어야 합니다 (현재: %d)", mm)
	}
	if ss < 0 || ss > 59 {
		return fmt.Errorf("초(SS) 범위 오류 — 0~59 이어야 합니다 (현재: %d)", ss)
	}
	return nil
}

/// checkCircularDepends 는 depends_on 설정의 순환 참조를 DFS 로 탐지한다.
func checkCircularDepends(processes []model.ProcessConfig) error {
	graph := make(map[string][]string, len(processes))
	idSet := make(map[string]bool, len(processes))
	for _, p := range processes {
		graph[p.ID] = p.DependsOn
		idSet[p.ID]  = true
	}
	for _, p := range processes {
		for _, depID := range p.DependsOn {
			if !idSet[depID] {
				return fmt.Errorf("프로세스 [%s]: depends_on [%s] 는 존재하지 않는 ID", p.ID, depID)
			}
		}
	}
	visited := make(map[string]bool, len(processes))
	inStack := make(map[string]bool, len(processes))
	var dfs func(id string) error
	dfs = func(id string) error {
		if inStack[id] {
			return fmt.Errorf("순환 참조 탐지: [%s]", id)
		}
		if visited[id] { return nil }
		inStack[id] = true
		for _, dep := range graph[id] {
			if err := dfs(dep); err != nil {
				return fmt.Errorf("%w ← [%s]", err, id)
			}
		}
		inStack[id] = false
		visited[id]  = true
		return nil
	}
	for _, p := range processes {
		if err := dfs(p.ID); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 덤프 출력 유틸리티
////////////////////////////////////////////////////////////////////////////////

/// modeDescription 은 ProcessMode 의 출력용 설명 문자열을 반환한다.
func modeDescription(mode model.ProcessMode) string {
	switch mode {
	case model.ModePermanent:  return "P (Permanent  — 상시 운영)"
	case model.ModeRestricted: return "R (Restricted — 시간 제한)"
	case model.ModeOnce:       return "O (Once       — 1회 실행)"
	default:                   return string(mode)
	}
}

/// groupOrDefault 는 그룹명이 비어 있으면 "Default" 를 반환한다.
func groupOrDefault(group string) string {
	if group == "" { return "Default" }
	return group
}

/// mapGet 은 map[string]string 에서 키를 조회하고 없으면 def 를 반환한다.
func mapGet(m map[string]string, key, def string) string {
	if v, ok := m[key]; ok { return v }
	return def
}
