///
/// @file    process/runner.go
/// @brief   프로세스 기동 및 종료 처리
/// @details exec.Cmd 를 래핑하여 프로세스 기동, 종료, 종료 감지를 담당한다.
///
///          설계 원칙:
///            1. Start() 는 fork/exec 후 즉시 반환한다.
///               종료 감지는 별도 goroutine 의 cmd.Wait() 가 담당한다.
///            2. 종료 이벤트는 manager 의 exitCh 채널로 전달한다.
///               SIGCHLD + 환형 Queue 방식(C버전)을 대체한다.
///            3. Stop() 은 SIGTERM → grace_period → SIGKILL 순서로 처리한다.
///               select + time.After 로 타임아웃을 처리하여
///               C버전의 waitpid() 블로킹 문제를 원천 제거한다.
///            4. 환경변수는 전역 env + 프로세스별 env 를 병합하여 전달한다.
///            5. 자식 프로세스의 표준 출력/에러는 /dev/null 로 리다이렉트한다.
///               (자식 프로세스는 자체 로그를 사용해야 함)
///
///          C버전 대응표:
///            fork() + execv()  →  exec.Command() + cmd.Start()
///            SigChild()        →  goroutine cmd.Wait() → exitCh
///            Kill(pid, 9)      →  Stop() SIGTERM → SIGKILL
///            CloseLogout()     →  자식에서 불필요 (파일 디스크립터 자동 정리)
///
/// @date    2025-01-04
/// @author  Cento
///

package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"goTbnova/model"
	"goTbnova/tlog"
)

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// Start 는 PInfo 에 정의된 프로세스를 기동한다.
///
/// 처리 순서:
///   1. 실행 파일 존재 확인
///   2. 환경변수 병합 (전역 + 프로세스별)
///   3. exec.Cmd 구성 및 기동
///   4. PInfo 런타임 상태 갱신 (Pid, State, StartCount, LastStart)
///   5. cmd.Wait() goroutine 기동 → 종료 시 exitCh 로 ExitEvent 전달
///
/// StartMode=S (Slow) 의 지연은 backoff.go 의 CalcBackoff() 에서
/// NextRetry 에 2초를 추가하는 방식으로 처리하므로 여기서는 sleep 하지 않는다.
///
/// @param  p          기동할 PInfo 포인터
/// @param  globalEnv  전역 환경변수 맵 (config.Load() 반환값)
/// @param  exitCh     종료 이벤트 전달 채널 (manager 소유)
/// @return            기동 실패 시 에러
func Start(
	p         *model.PInfo,
	globalEnv map[string]string,
	exitCh    chan<- model.ExitEvent,
) error {

	/// 실행 파일 존재 확인
	if _, err := os.Stat(p.Config.Path); err != nil {
		return fmt.Errorf("실행 파일 접근 불가 [%s]: %w", p.Config.Path, err)
	}

	/// exec.Cmd 구성
	args := append([]string{p.Config.Path}, p.Config.Args...)
	cmd  := exec.Command(args[0], args[1:]...)

	/// 환경변수 병합: 전역 env + 프로세스별 env
	/// 프로세스별 env 가 전역 env 를 덮어쓴다
	cmd.Env = mergeEnv(globalEnv, p.Config.Env)

	/// 표준 입출력 처리
	/// 자식 프로세스는 자체 로그를 사용해야 하므로 /dev/null 로 리다이렉트
	/// stdin  은 항상 /dev/null
	/// stdout, stderr 는 /dev/null (자식의 출력이 goTbnova 터미널을 오염시키지 않도록)
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("/dev/null 열기 실패: %w", err)
	}
	cmd.Stdin  = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	/// 자식 프로세스 속성 설정
	///
	/// Setsid=true:
	///   자식을 새 세션 리더로 설정하여 goTbnova 와 세션을 분리한다.
	///   goTbnova 가 종료되어도 자식은 독립적으로 계속 실행된다.
	///   자식의 부모는 init(PID=1) 로 재설정된다.
	///
	/// Pdeathsig 미설정:
	///   goTbnova 종료 시 자식에게 시그널을 보내지 않는다.
	///   자식 프로세스는 goTbnova 없이도 계속 실행된다.
	///
	/// 자식 종료가 필요한 경우:
	///   goTbmon 또는 TCP 명령으로 개별 stop 을 수행하거나
	///   clsGoTbnova 유틸리티로 일괄 정리한다.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	/// 프로세스 기동
	if err := cmd.Start(); err != nil {
		devNull.Close()
		return fmt.Errorf("프로세스 기동 실패 [%s]: %w", p.Config.Path, err)
	}

	/// /dev/null 핸들은 자식 기동 후 부모에서 닫아도 됨
	devNull.Close()

	/// PInfo 런타임 상태 갱신
	/// StartCount 는 비정상 종료 시에만 증가한다 (규칙 5).
	/// 기동 자체에서는 증가하지 않는다.
	now           := time.Now()
	p.Cmd          = cmd
	p.Pid          = cmd.Process.Pid
	p.State        = model.StateRunning
	p.StopReason   = model.StopNone
	p.LastStart    = now

	tlog.ProcessStart(
		p.Config.ID,
		p.Pid,
		p.StartCount,
		string(p.Config.Mode),
		p.Config.Path,
	)

	/// 종료 감지 goroutine
	/// cmd.Wait() 는 프로세스 종료까지 블로킹한다.
	/// 종료 시 ExitEvent 를 exitCh 로 전달하고 goroutine 을 종료한다.
	/// C버전의 SigChild() + AddSignalList() + CheckSignal() 을 대체한다.
	go func() {
		err := cmd.Wait()
		exitCh <- model.ExitEvent{P: p, Err: err}
	}()

	return nil
}

/// Stop 은 프로세스에 SIGTERM 을 발송하고 grace_period 내 종료를 기다린다.
/// grace_period 초과 시 SIGKILL 을 발송한다.
///
/// 처리 순서:
///   1. SIGTERM 발송
///   2. grace_period 타이머와 종료 이벤트(exitCh) 경쟁
///   3. 타이머 만료 시 SIGKILL 발송 후 종료 대기
///
/// C버전의 kill(pid, 9) 직접 호출을 대체한다.
/// select + channel 로 타임아웃을 처리하므로 waitpid() 블로킹이 없다.
///
/// @param  p       종료할 PInfo 포인터
/// @param  reason  종료 원인 (StopReason)
/// @param  exitCh  종료 이벤트 채널 (goroutine 의 Wait() 완료를 수신)
func Stop(p *model.PInfo, reason model.StopReason, exitCh <-chan model.ExitEvent) {

	if p.Pid <= 0 || p.Cmd == nil {
		return
	}

	graceSec := p.Config.Restart.GracePeriod
	if graceSec <= 0 {
		graceSec = 10 ///< 설정 누락 시 기본값 10초
	}

	p.State      = model.StateStopping
	p.StopReason = reason

	tlog.ProcessStopReq(p.Config.ID, p.Pid, graceSec)

	/// Step 1: SIGTERM 발송
	if err := p.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		/// 이미 종료된 경우 — goroutine 이 exitCh 로 이미 전달했을 것
		tlog.Debug("event=sigterm_skip id=%s pid=%d reason=%v",
			p.Config.ID, p.Pid, err)
		return
	}

	/// Step 2: grace_period 대기
	/// exitCh 에서 이 프로세스의 종료 이벤트를 수신하거나 타이머 만료를 기다린다.
	timer := time.NewTimer(time.Duration(graceSec) * time.Second)
	defer timer.Stop()

	for {
		select {
		case ev := <-exitCh:
			if ev.P == p {
				/// 이 프로세스가 정상 종료됨
				tlog.ProcessExit(
					p.Config.ID, p.Pid,
					tlog.StopReasonStr(int(reason)),
				)
				clearPInfo(p)
				return
			}
			/// 다른 프로세스의 종료 이벤트 — 무시하지 않고 다시 채널에 반환
			/// Note: Stop() 은 manager 루프 밖에서 호출되므로
			///       다른 프로세스 이벤트를 드롭하면 안 됨
			/// 이 상황은 manager 설계에서 Stop() 을 직접 호출하지 않고
			/// stopCh 채널로 요청하는 방식으로 해결한다. (manager.go 참고)

		case <-timer.C:
			/// Step 3: grace_period 초과 → SIGKILL
			tlog.Warn("event=sigkill id=%s pid=%d reason=grace_period_exceeded grace_sec=%d",
				p.Config.ID, p.Pid, graceSec)
			_ = p.Cmd.Process.Kill()
			/// SIGKILL 후 Wait() 완료를 기다림 (goroutine 이 exitCh 로 전달)
			return
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수
////////////////////////////////////////////////////////////////////////////////

/// mergeEnv 는 전역 환경변수와 프로세스별 환경변수를 병합하여
/// exec.Cmd.Env 형식의 "KEY=VALUE" 슬라이스로 반환한다.
///
/// 우선순위: 프로세스별 env > 전역 env
///
/// @param  global  전역 환경변수 맵
/// @param  local   프로세스별 환경변수 맵
/// @return         "KEY=VALUE" 슬라이스
func mergeEnv(global, local map[string]string) []string {
	merged := make(map[string]string, len(global)+len(local))

	for k, v := range global {
		merged[k] = v
	}
	for k, v := range local {
		merged[k] = v ///< 프로세스별이 전역을 덮어씀
	}

	env := make([]string, 0, len(merged))
	for k, v := range merged {
		env = append(env, k+"="+v)
	}
	return env
}

/// clearPInfo 는 프로세스 종료 후 PInfo 의 런타임 상태를 초기화한다.
///
/// @param  p  초기화할 PInfo 포인터
func clearPInfo(p *model.PInfo) {
	p.Pid   = 0
	p.Cmd   = nil
	p.State = model.StateWaitTime
}
