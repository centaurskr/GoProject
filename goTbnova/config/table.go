///
/// @file    config/table.go
/// @brief   goTbnova TOML 설정 파일 로드 및 파싱
/// @details TOML 형식의 프로세스 테이블을 읽어 model.TableConfig 로 변환한다.
///
///          처리 순서:
///            1. TOML 파일 파싱 → TableConfig
///            2. 필수 항목 검증 (validator.go)
///            3. 전역 env 변수 치환 ($VAR 형식)
///            4. 프로세스별 env 변수 치환 (전역 env → OS env 순서로 탐색)
///            5. path 절대경로 변환 (def_dir 기준)
///            6. 런타임 구조체(PInfo, GInfo) 생성
///            7. 종속관계(Dependents) 연결
///
///          환경변수 치환 규칙:
///            - $VAR 형식만 지원 (${VAR} 미지원)
///            - 전역 [env] → OS 환경변수 순서로 탐색
///            - 치환 불가 시 빈 문자열 (쉘 동작과 동일)
///            - 치환은 Load() 시점에 1회만 수행
///
/// @date    2026-04-27
/// @author  Cento
///

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/toml"

	"goTbnova/model"
)

////////////////////////////////////////////////////////////////////////////////
// 상수 정의
////////////////////////////////////////////////////////////////////////////////

/// Back-off 대기 시간 수열 (초)
/// 1 → 2 → 4 → 8 → 16 → 32 → 300 (이후 상한 유지)
var backoffTable = []int{1, 2, 4, 8, 16, 32, 300}

/// 기본값 상수
const (
	defaultGracePeriod = 10 ///< SIGTERM 후 SIGKILL 까지 기본 대기 시간 (초)
	defaultMaxCount    = 5  ///< 기본 최대 재기동 횟수
	defaultInterval    = 10 ///< 기본 헬스체크 주기 (초) — 현재 미사용, 확장 예비
)

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// Load 는 TOML 설정 파일을 읽어 런타임 구조체 슬라이스를 반환한다.
///
/// @param  path  TOML 설정 파일 전체 경로
/// @return processes  PInfo 슬라이스 (manager 에서 사용)
/// @return groups     GInfo 슬라이스 (manager 에서 사용)
/// @return globalEnv  전역 환경변수 맵 (runner 에서 exec.Cmd.Env 에 사용)
/// @return err        파싱/검증 실패 시 에러
func Load(path string) (
	processes []*model.PInfo,
	groups []*model.GInfo,
	globalEnv map[string]string,
	err error,
) {
	/// Step 1: TOML 파싱
	var cfg model.TableConfig
	if _, err = toml.DecodeFile(path, &cfg); err != nil {
		err = fmt.Errorf("TOML 파싱 실패 [%s]: %w", path, err)
		return
	}

	/// Step 2: 필수 항목 검증
	if err = Validate(&cfg); err != nil {
		err = fmt.Errorf("설정 검증 실패: %w", err)
		return
	}

	/// Step 3: 전역 env 내부 변수 치환
	/// 예) LOG_DIR = "$LOG_HOME/log" → 전역 env 내 다른 키 참조
	globalEnv = expandGlobalEnv(cfg.Env)

	/// Step 4: 기본값 보완
	applyDefaults(&cfg)

	/// Step 5: GInfo 생성
	groups = buildGroups(&cfg)

	/// Step 6: PInfo 생성
	processes, err = buildProcesses(&cfg, groups, globalEnv)
	if err != nil {
		return
	}

	/// Step 7: 종속관계(Dependents) 연결
	if err = linkDependents(processes); err != nil {
		return
	}

	return
}

/// BackoffDuration 은 실패 횟수에 따른 Back-off 대기 시간을 반환한다.
///
/// @param  failCount  연속 실패 횟수 (1부터 시작)
/// @return            대기 시간 (time.Duration)
func BackoffDuration(failCount int) time.Duration {
	idx := failCount - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(backoffTable) {
		idx = len(backoffTable) - 1
	}
	return time.Duration(backoffTable[idx]) * time.Second
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 환경변수 치환
////////////////////////////////////////////////////////////////////////////////

/// expandGlobalEnv 는 전역 env 맵 내부의 $VAR 를 치환한다.
/// 전역 env 내에서 다른 키를 참조할 수 있다.
/// 순환 참조는 탐지하지 않으며 탐색 불가 시 OS 환경변수를 사용한다.
///
/// @param  raw  원본 전역 env 맵
/// @return      치환 완료된 env 맵
func expandGlobalEnv(raw map[string]string) map[string]string {
	result := make(map[string]string, len(raw))

	/// 전역 env 를 먼저 복사 (미치환 원본 참조용)
	for k, v := range raw {
		result[k] = v
	}

	/// 치환: 전역 env 내부 참조 → OS 환경변수 순서
	for k, v := range result {
		result[k] = expandVar(v, result)
	}
	return result
}

/// expandProcessEnv 는 프로세스별 env 맵의 $VAR 를 치환한다.
/// 전역 env → OS 환경변수 순서로 탐색한다.
///
/// @param  raw        원본 프로세스별 env 맵
/// @param  globalEnv  치환 완료된 전역 env 맵
/// @return            치환 완료된 env 맵
func expandProcessEnv(raw map[string]string, globalEnv map[string]string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = expandVar(v, globalEnv)
	}
	return result
}

/// expandVar 는 문자열 내의 $VAR 를 치환한다.
///
/// 탐색 순서:
///   1. lookup 맵 (전역 env)
///   2. OS 환경변수
///   3. 빈 문자열 (탐색 실패)
///
/// 지원 형식: $VAR  (영문자/숫자/밑줄로 구성된 변수명)
/// 미지원   : ${VAR}
///
/// @param  s       원본 문자열
/// @param  lookup  우선 탐색할 맵 (전역 env)
/// @return         치환 완료된 문자열
func expandVar(s string, lookup map[string]string) string {
	if !strings.Contains(s, "$") {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] != '$' {
			b.WriteByte(s[i])
			i++
			continue
		}

		/// $ 발견 — 변수명 추출
		i++ /// $ 건너뜀
		j := i
		for j < len(s) && isVarChar(rune(s[j])) {
			j++
		}

		if j == i {
			/// $ 뒤에 변수명 없음 → $ 그대로 출력
			b.WriteByte('$')
			continue
		}

		varName := s[i:j]
		i = j

		/// 탐색 순서: lookup 맵 → OS 환경변수
		if v, ok := lookup[varName]; ok {
			b.WriteString(v)
		} else if v := os.Getenv(varName); v != "" {
			b.WriteString(v)
		}
		/// 탐색 실패 시 빈 문자열 (아무것도 쓰지 않음)
	}
	return b.String()
}

/// isVarChar 는 환경변수명에 허용되는 문자인지 확인한다.
/// 영문자, 숫자, 밑줄(_) 만 허용한다.
func isVarChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 기본값 보완
////////////////////////////////////////////////////////////////////////////////

/// applyDefaults 는 설정 파일에 누락된 항목에 기본값을 적용한다.
///
/// @param  cfg  파싱 완료된 TableConfig (수정됨)
func applyDefaults(cfg *model.TableConfig) {
	if cfg.Default.DefCnt == 0 {
		cfg.Default.DefCnt = defaultMaxCount
	}

	for i := range cfg.Process {
		p := &cfg.Process[i]

		/// 재기동 정책 기본값
		if p.Restart.Policy == "" {
			p.Restart.Policy = model.PolicyOnFailure
		}
		if p.Restart.MaxCount == 0 {
			p.Restart.MaxCount = cfg.Default.DefCnt
		}
		if p.Restart.GracePeriod == 0 {
			p.Restart.GracePeriod = defaultGracePeriod
		}

		/// 종속 연쇄 기본값
		if p.Cascade.OnDeath == "" {
			p.Cascade.OnDeath = model.CascadeNone
		}

		/// 재기동 방식 기본값
		if p.StMode == "" {
			p.StMode = model.StartQuick
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 런타임 구조체 생성
////////////////////////////////////////////////////////////////////////////////

/// buildGroups 는 TableConfig 의 Groups 맵을 GInfo 슬라이스로 변환한다.
/// "Default" 그룹을 항상 첫 번째로 추가한다.
///
/// @param  cfg  파싱 완료된 TableConfig
/// @return      GInfo 슬라이스
func buildGroups(cfg *model.TableConfig) []*model.GInfo {
	groups := make([]*model.GInfo, 0, len(cfg.Groups)+1)

	/// Default 그룹은 항상 존재
	groups = append(groups, &model.GInfo{
		Name: "Default",
		Desc: "기본 그룹",
	})

	for name, gc := range cfg.Groups {
		if name == "Default" {
			continue /// 중복 방지
		}
		groups = append(groups, &model.GInfo{
			Name: name,
			Desc: gc.Desc,
		})
	}
	return groups
}

/// buildProcesses 는 TableConfig 의 Process 슬라이스를 PInfo 슬라이스로 변환한다.
///
/// @param  cfg        파싱 완료된 TableConfig
/// @param  groups     buildGroups() 결과
/// @param  globalEnv  치환 완료된 전역 env 맵
/// @return            PInfo 슬라이스
/// @return            에러
func buildProcesses(
	cfg *model.TableConfig,
	groups []*model.GInfo,
	globalEnv map[string]string,
) ([]*model.PInfo, error) {

	processes := make([]*model.PInfo, 0, len(cfg.Process))

	for _, pc := range cfg.Process {

		/// path 절대경로 변환
		absPath, err := resolvePath(pc.Path, cfg.Default.DefDir)
		if err != nil {
			return nil, fmt.Errorf("프로세스 [%s] path 오류: %w", pc.ID, err)
		}
		pc.Path = absPath

		/// 프로세스별 env 치환
		pc.Env = expandProcessEnv(pc.Env, globalEnv)

		/// 기동/종료 시각 파싱
		start, end, err := parseTimeRange(pc.St, pc.Et)
		if err != nil {
			return nil, fmt.Errorf("프로세스 [%s] 시각 파싱 오류: %w", pc.ID, err)
		}

		p := &model.PInfo{
			Config: pc,
			State:  model.StateWaitTime,
			Start:  start,
			End:    end,
		}

		/// 그룹 연결 및 Members 등록
		g := findGroup(groups, pc.Group)
		if g == nil {
			g = groups[0] /// Default 그룹으로 fallback
		}
		g.Members = append(g.Members, p)

		processes = append(processes, p)
	}
	return processes, nil
}

/// linkDependents 는 depends_on 설정을 기반으로 PInfo.Dependents 를 연결한다.
/// A.depends_on = ["B"] 이면 B.Dependents 에 A 를 추가한다.
///
/// @param  processes  buildProcesses() 결과
/// @return            depends_on 에 존재하지 않는 ID 가 있으면 에러
func linkDependents(processes []*model.PInfo) error {
	/// ID → PInfo 인덱스 맵 구성
	idx := make(map[string]*model.PInfo, len(processes))
	for _, p := range processes {
		idx[p.Config.ID] = p
	}

	for _, p := range processes {
		for _, depID := range p.Config.DependsOn {
			dep, ok := idx[depID]
			if !ok {
				return fmt.Errorf(
					"프로세스 [%s] depends_on [%s]: 존재하지 않는 ID",
					p.Config.ID, depID,
				)
			}
			dep.Dependents = append(dep.Dependents, p)
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수 — 유틸리티
////////////////////////////////////////////////////////////////////////////////

/// resolvePath 는 실행 파일 경로를 절대경로로 변환한다.
/// 이미 절대경로이면 그대로 반환한다.
/// 상대경로이면 defDir 을 기준으로 절대경로를 구성한다.
///
/// @param  path    원본 path (절대 또는 상대)
/// @param  defDir  기준 디렉토리 (TOML [default].def_dir)
/// @return         절대경로
/// @return         에러
func resolvePath(path, defDir string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	if defDir == "" {
		return "", fmt.Errorf("상대경로 [%s] 사용 시 def_dir 이 필요합니다", path)
	}
	return filepath.Join(defDir, path), nil
}

/// parseTimeRange 는 "HH:MM:SS" 형식의 시작/종료 시각을 오늘 날짜 기준 time.Time 으로 변환한다.
/// 시작 시각이 종료 시각보다 크면 종료 시각을 익일로 처리한다.
/// (예: st=23:00:00, et=01:30:00 → 당일 23시 ~ 익일 01시 30분)
///
/// @param  st  시작 시각 문자열 "HH:MM:SS"
/// @param  et  종료 시각 문자열 "HH:MM:SS"
/// @return     start, end time.Time
/// @return     에러
func parseTimeRange(st, et string) (start, end time.Time, err error) {
	now := time.Now()
	loc := now.Location()

	start, err = parseHMS(st, now, loc)
	if err != nil {
		err = fmt.Errorf("시작 시각 [%s] 파싱 오류: %w", st, err)
		return
	}
	end, err = parseHMS(et, now, loc)
	if err != nil {
		err = fmt.Errorf("종료 시각 [%s] 파싱 오류: %w", et, err)
		return
	}

	/// 자정을 넘기는 구간 처리
	/// st=23:00:00, et=01:30:00 → et 를 익일로 조정
	if start.After(end) {
		end = end.Add(24 * time.Hour)
	}
	return
}

/// parseHMS 는 "HH:MM:SS" 문자열을 오늘 날짜 기준 time.Time 으로 변환한다.
///
/// @param  s    시각 문자열 "HH:MM:SS"
/// @param  now  기준 시각 (오늘 날짜 추출용)
/// @param  loc  타임존
/// @return      time.Time
/// @return      에러
func parseHMS(s string, now time.Time, loc *time.Location) (time.Time, error) {
	var hh, mm, ss int
	n, err := fmt.Sscanf(s, "%02d:%02d:%02d", &hh, &mm, &ss)
	if err != nil || n != 3 {
		return time.Time{}, fmt.Errorf("형식 오류 (HH:MM:SS 필요): %s", s)
	}
	if hh < 0 || hh > 23 || mm < 0 || mm > 59 || ss < 0 || ss > 59 {
		return time.Time{}, fmt.Errorf("시각 범위 오류: %s", s)
	}
	return time.Date(now.Year(), now.Month(), now.Day(), hh, mm, ss, 0, loc), nil
}

/// findGroup 은 groups 슬라이스에서 name 과 일치하는 GInfo 를 반환한다.
/// 없으면 nil 을 반환한다.
///
/// @param  groups  GInfo 슬라이스
/// @param  name    찾을 그룹 이름
/// @return         GInfo 포인터 (없으면 nil)
func findGroup(groups []*model.GInfo, name string) *model.GInfo {
	for _, g := range groups {
		if g.Name == name {
			return g
		}
	}
	return nil
}
