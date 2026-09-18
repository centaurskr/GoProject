///
/// @file    config/envfile.go
/// @brief   goTbnova 환경 설정 파일 (gotbnova.toml) 로드
/// @details $ENV_HOME/gotbnova.toml 을 읽어 goTbnova 실행 환경을 구성한다.
///
///          파일 위치 결정:
///            1. $ENV_HOME/gotbnova.toml   (os.Getenv("ENV_HOME"))
///            2. 실행 파일 디렉토리/gotbnova.toml  (폴백)
///            둘 다 없으면 기본값으로 동작한다.
///
///          섹션 구성:
///            [server]  모니터 TCP 바인딩 설정
///            [log]     로그 백엔드 설정 (net / file)
///            [work]    작업 디렉토리 설정
///
/// @date    2026-04-27
/// @author  Cento
///

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

////////////////////////////////////////////////////////////////////////////////
// 상수
////////////////////////////////////////////////////////////////////////////////

const (
	envFileName         = "gotbnova.toml"           ///< 환경 파일명
	defaultBind         = "127.0.0.1"               ///< 기본 바인딩 IP
	defaultLogType      = "net"                     ///< 기본 로그 백엔드
	defaultVectorSocket = "/tmp/VECTOR.UNIX.DGRAM"  ///< 기본 Vector UDS 소켓 경로
)

////////////////////////////////////////////////////////////////////////////////
// 타입 정의
////////////////////////////////////////////////////////////////////////////////

/// EnvFile 은 gotbnova.toml 전체 구조를 담는다.
type EnvFile struct {
	Server EnvServer `toml:"server"` ///< 모니터 서버 설정
	Log    EnvLog    `toml:"log"`    ///< 로그 백엔드 설정
	Work   EnvWork   `toml:"work"`   ///< 작업 디렉토리 설정
}

/// EnvServer 는 [server] 섹션을 담는다.
type EnvServer struct {
	/// 모니터 TCP 바인딩 IP 주소
	///   "127.0.0.1"   : 로컬 전용 (기본값, 권장)
	///   "0.0.0.0"     : 모든 인터페이스 허용
	///   "192.168.x.x" : 특정 IP 만 허용
	Bind string `toml:"bind"` ///< 기본값: "127.0.0.1"
}

/// EnvLog 는 [log] 섹션을 담는다.
type EnvLog struct {
	/// 로그 백엔드 종류
	///   "net"  : centolib/netlog → Vector UDS → VictoriaLogs
	///   "file" : 로컬 파일 직접 기록
	Type string `toml:"type"` ///< 기본값: "net"

	/// Vector UNIX Domain Socket 경로 (type = "net" 일 때만 사용)
	/// centolib/netlog 의 SocketPath 와 동일한 값이어야 한다.
	VectorSocket string `toml:"vector_socket"` ///< 기본값: "/tmp/VECTOR.UNIX.DGRAM"

	/// 로컬 파일 로그 디렉토리 (type = "file" 일 때만 사용)
	/// 비어있으면 실행파일경로/log 로 자동 탐색
	/// $LOG_HOME 환경변수는 참조하지 않는다.
	Dir string `toml:"dir"` ///< 기본값: ""
}

/// EnvWork 는 [work] 섹션을 담는다.
type EnvWork struct {
	/// goTbnova 작업 디렉토리
	/// 비어있으면 현재 디렉토리 유지
	Dir string `toml:"dir"` ///< 기본값: ""
}

////////////////////////////////////////////////////////////////////////////////
// 공개 함수
////////////////////////////////////////////////////////////////////////////////

/// LoadEnvFile 은 $ENV_HOME/gotbnova.toml 을 읽어 EnvFile 구조체를 반환한다.
///
/// 파일 위치 결정 순서:
///   1. $ENV_HOME/gotbnova.toml
///   2. 실행 파일 디렉토리/gotbnova.toml  (폴백)
///
/// 파일이 없으면 기본값으로 초기화된 EnvFile 을 반환한다. (오류 아님)
/// 파일이 있으나 파싱 실패하면 오류를 반환한다.
///
/// @return  EnvFile  로드된 환경 설정 (없으면 기본값)
/// @return  string   실제 로드된 파일 경로 (없으면 빈 문자열)
/// @return  error    파싱 실패 시 오류
func LoadEnvFile() (EnvFile, string, error) {
	ef := defaultEnvFile()

	path := resolveEnvFilePath()
	if path == "" {
		/// 파일 없음 — 기본값으로 동작
		return ef, "", nil
	}

	if _, err := toml.DecodeFile(path, &ef); err != nil {
		return defaultEnvFile(), path,
			fmt.Errorf("gotbnova.toml 파싱 실패 [%s]: %w", path, err)
	}

	/// 기본값 보완
	applyEnvDefaults(&ef)

	return ef, path, nil
}

/// DumpEnvFile 은 EnvFile 내용을 표준 출력으로 덤프한다.
/// main.go 의 -d 옵션에서 호출한다.
///
/// @param  ef    덤프할 EnvFile
/// @param  path  로드된 파일 경로
func DumpEnvFile(ef EnvFile, path string) {
	const sep = "──────────────────────────────────────"

	if path == "" {
		fmt.Println("[gotbnova.toml] 파일 없음 — 기본값 사용")
	} else {
		fmt.Printf("[gotbnova.toml] %s\n", path)
	}
	fmt.Println(sep)

	fmt.Println("\n[server]")
	fmt.Printf("  bind          = %q\n", ef.Server.Bind)

	fmt.Println("\n[log]")
	fmt.Printf("  type          = %q\n", ef.Log.Type)
	if ef.Log.Type == "net" {
		fmt.Printf("  vector_socket = %q\n", ef.Log.VectorSocket)
	} else {
		fmt.Printf("  dir           = %q\n", ef.Log.Dir)
	}

	fmt.Println("\n[work]")
	fmt.Printf("  dir           = %q\n", ef.Work.Dir)
	fmt.Println(sep)
}

////////////////////////////////////////////////////////////////////////////////
// 내부 함수
////////////////////////////////////////////////////////////////////////////////

/// resolveEnvFilePath 는 gotbnova.toml 의 실제 경로를 결정한다.
///
/// 탐색 순서:
///   1. $ENV_HOME/gotbnova.toml
///   2. 실행 파일 디렉토리/gotbnova.toml
///
/// @return  존재하는 파일 경로 (없으면 빈 문자열)
func resolveEnvFilePath() string {
	/// 1순위: $ENV_HOME/gotbnova.toml
	if envHome := os.Getenv("ENV_HOME"); envHome != "" {
		path := filepath.Join(envHome, envFileName)
		if fileExists(path) {
			return path
		}
	}

	/// 2순위: 실행 파일 디렉토리/gotbnova.toml
	if exe, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exe), envFileName)
		if fileExists(path) {
			return path
		}
	}

	return ""
}

/// defaultEnvFile 은 모든 항목이 기본값으로 설정된 EnvFile 을 반환한다.
func defaultEnvFile() EnvFile {
	return EnvFile{
		Server: EnvServer{
			Bind: defaultBind,
		},
		Log: EnvLog{
			Type:         defaultLogType,
			VectorSocket: defaultVectorSocket,
			Dir:          "",
		},
		Work: EnvWork{
			Dir: "",
		},
	}
}

/// applyEnvDefaults 는 EnvFile 에서 빈 값에 기본값을 적용한다.
///
/// @param  ef  보완할 EnvFile 포인터
func applyEnvDefaults(ef *EnvFile) {
	if ef.Server.Bind == "" {
		ef.Server.Bind = defaultBind
	}
	if ef.Log.Type == "" {
		ef.Log.Type = defaultLogType
	}
	if ef.Log.VectorSocket == "" {
		ef.Log.VectorSocket = defaultVectorSocket
	}
}

/// fileExists 는 파일이 존재하는지 확인한다.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
