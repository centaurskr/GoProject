///
/// @file    config/diff.go
/// @brief   테이블 리로드 시 변경분 감지
/// @details 현재 운영 중인 설정과 새 설정을 비교하여
///          추가/삭제/변경된 프로세스를 구분한다.
///          실행 중인 프로세스는 설정이 변경되지 않았으면 그대로 유지한다.
///
/// @date    2026-04-27
/// @author  Cento
///

package config

import (
	"goTbnova/model"
)

////////////////////////////////////////////////////////////////////////////////
// 타입 정의
////////////////////////////////////////////////////////////////////////////////

/// DiffResult 는 테이블 리로드 시 변경분 탐지 결과를 담는다.
type DiffResult struct {
	Added    []model.ProcessConfig ///< 새로 추가된 프로세스 설정
	Removed  []model.ProcessConfig ///< 삭제된 프로세스 설정
	Modified []model.ProcessConfig ///< 변경된 프로세스 설정 (새 설정값)
	Unchanged []model.ProcessConfig ///< 변경 없는 프로세스 설정
}

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// Diff 는 현재 설정과 새 설정을 비교하여 변경분을 반환한다.
/// ID 를 기준으로 비교하며, 내용 변경은 핵심 필드만 확인한다.
///
/// @param  current  현재 운영 중인 PInfo 슬라이스
/// @param  next     새로 로드된 TableConfig
/// @return          DiffResult
func Diff(current []*model.PInfo, next *model.TableConfig) DiffResult {
	var result DiffResult

	/// 현재 프로세스 ID 맵 구성
	currMap := make(map[string]*model.PInfo, len(current))
	for _, p := range current {
		currMap[p.Config.ID] = p
	}

	/// 새 설정 ID 맵 구성
	nextMap := make(map[string]model.ProcessConfig, len(next.Process))
	for _, pc := range next.Process {
		nextMap[pc.ID] = pc
	}

	/// 새 설정 기준으로 추가/변경/유지 분류
	for _, pc := range next.Process {
		curr, exists := currMap[pc.ID]
		if !exists {
			result.Added = append(result.Added, pc)
			continue
		}
		if isModified(curr.Config, pc) {
			result.Modified = append(result.Modified, pc)
		} else {
			result.Unchanged = append(result.Unchanged, pc)
		}
	}

	/// 현재 설정 기준으로 삭제된 프로세스 분류
	for _, p := range current {
		if _, exists := nextMap[p.Config.ID]; !exists {
			result.Removed = append(result.Removed, p.Config)
		}
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수
////////////////////////////////////////////////////////////////////////////////

/// isModified 는 프로세스 설정의 핵심 필드가 변경되었는지 확인한다.
/// 재기동이 필요한 변경인지를 판단하는 기준이다.
///
/// 비교 대상:
///   - path, args       실행 파일 및 인자 변경
///   - mode, st, et     운영 모드 및 시간 변경
///   - weekend          주말 운영 여부 변경
///   - restart 정책     재기동 정책 변경
///   - depends_on       종속관계 변경
///
/// 비교 제외:
///   - desc             설명은 운영에 영향 없음
///   - env              환경변수는 다음 기동 시 적용
///
/// @param  curr  현재 설정
/// @param  next  새 설정
/// @return       변경 여부
func isModified(curr, next model.ProcessConfig) bool {
	if curr.Path    != next.Path    { return true }
	if curr.Mode    != next.Mode    { return true }
	if curr.St      != next.St      { return true }
	if curr.Et      != next.Et      { return true }
	if curr.Weekend != next.Weekend { return true }
	if curr.StMode  != next.StMode  { return true }

	/// 재기동 정책 비교
	if curr.Restart.Policy      != next.Restart.Policy      { return true }
	if curr.Restart.Backoff     != next.Restart.Backoff     { return true }
	if curr.Restart.MaxCount    != next.Restart.MaxCount    { return true }
	if curr.Restart.GracePeriod != next.Restart.GracePeriod { return true }

	/// Cascade 비교
	if curr.Cascade.OnDeath != next.Cascade.OnDeath { return true }

	/// Args 비교
	if !equalStringSlice(curr.Args, next.Args) { return true }

	/// DependsOn 비교
	if !equalStringSlice(curr.DependsOn, next.DependsOn) { return true }

	return false
}

/// equalStringSlice 는 두 문자열 슬라이스가 동일한지 확인한다.
/// 순서까지 동일해야 true 를 반환한다.
///
/// @param  a, b  비교할 슬라이스
/// @return       동일 여부
func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
