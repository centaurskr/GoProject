///
/// @file netlog.go
/// @brief 전사 표준 스키마를 준수하는 VictoriaLogs용 네트워크 로거 패키지
/// @date   : 2025. 12. 22. (월) 10:33:44 KST
/// @author : Cento
/// @details C, Go, Java, Node.js 통합 로그 관리를 위해 설계되었으며, 
/// Unix Domain Socket을 통해 Vector로 로그를 비동기 전송합니다.
/// @file netlog.go
/// @brief Unix Domain Socket(Datagram)을 이용한 비동기 로그 전송 패키지
/// @details Vector 에이전트와 통신하며, 내부적으로 고루틴과 채널을 사용하여 
/// Non-blocking 로그 전송을 지원합니다.
//
// @sample
// package main
// import (
//	"time"
//	"your_project/netlog"
// )
// func main() {
//	netlog.InitNetLogout("PaymentService", 1)
//	// 프로그램 종료 시 로그 유실 방지
//	defer netlog.Close()
//	// 비동기 로그 전송 (매우 빠름)
//	netlog.Info("Service started")
//	for i := 0; i < 100; i++ {
//		go func(id int) {
//			netlog.Debug("Processing task %d", id)
//		}(i)
//	}
//
//	time.Sleep(1 * time.Second) // 로그가 전송될 시간 확보
//}

package netlog

import (
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"sync"
	"time"
	"strings"
)

// @brief 로그 소켓 경로 및 내부 버퍼 크기 설정
const (
	SocketPath = "/tmp/VECTOR.UNIX.DGRAM" //!< Vector 소켓 파일 위치
	BufferSize = 10000                     //!< 로그를 저장할 채널의 버퍼 크기
)

// @struct LogMessage
// @brief Vector 및 VictoriaLogs 전송용 JSON 구조체
type LogMessage struct {
	Ts       int64  `json:"ts"`       //!< 나노초 단위 타임스탬프
	Level    string `json:"level"`    //!< 로그 레벨 (DEBUG, INFOR, ERROR)
	Pname    string `json:"pname"`    //!< 프로세스 이름
	Extra    string `json:"extra"`    //!< 프로세스 추가 이름
	Source   string `json:"source"`   //!< 소스 파일명
	Function string `json:"function"` //!< 함수명
	Line     int    `json:"line"`     //!< 라인 번호
	Message  string `json:"message"`  //!< 로그 본문
}

// @struct Logger
// @brief 로거 상태 관리 구조체
type Logger struct {
	conn    *net.UnixConn    //!< 유닉스 소켓 연결 객체
	logChan chan LogMessage  //!< 로그 데이터 큐
	pname   string           //!< 기본 프로세스명
	pextra  string           //!< 기본 프로세스 추가이름
	wg      sync.WaitGroup   //!< 워커 종료 대기용
	done    chan struct{}    //!< 종료 신호 채널
}

var instance *Logger
var once sync.Once

// @brief 로거 초기화 함수
// @param pname 프로세스 이름
// @param pextra 프로세스 추가이름
// @return error 초기화 실패 시 에러 반환
func InitNetLogout(pname string, pextra string) error {
	var initErr error
	once.Do(func() {
		addr, err := net.ResolveUnixAddr("unixgram", SocketPath)
		if err != nil {
			initErr = err
			return
		}

		conn, err := net.DialUnix("unixgram", nil, addr)
		if err != nil {
			initErr = err
			return
		}

		instance = &Logger{
			conn:    conn,
			logChan: make(chan LogMessage, BufferSize),
			pname:   pname,
			pextra:  pextra,
			done:    make(chan struct{}),
		}

		// 백그라운드 워커 실행
		instance.wg.Add(1)
		go instance.worker()
	})
	return initErr
}

// @brief 백그라운드 로그 전송 워커
// @details 채널로부터 로그 메시지를 소비하여 소켓으로 실제 전송을 수행합니다.
func (l *Logger) worker() {
	defer l.wg.Done()
	defer l.conn.Close()

	for {
		select {
		case msg, ok := <-l.logChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(msg)
			l.conn.Write(data)
		case <-l.done:
			// 종료 처리: 채널에 남은 모든 로그 소진
			for len(l.logChan) > 0 {
				msg := <-l.logChan
				data, _ := json.Marshal(msg)
				l.conn.Write(data)
			}
			return
		}
	}
}

// @brief 로거 종료 함수
// @details 모든 로그를 전송한 뒤 소켓을 닫고 백그라운드 루틴을 종료합니다.
func Close() {
	if instance != nil {
		close(instance.done)
		instance.wg.Wait()
	}
}

// @brief 공통 로그 푸시 로직
// @param level 로그 레벨
// @param msg 로그 메시지 내용
func logPush(level, msg string) {
	if instance == nil {
		return
	}

	// 호출자 정보 획득 (depth 2)
	pc, file, line, _ := runtime.Caller(2)
	name := runtime.FuncForPC(pc).Name()
	funcName := "unknown"
	if i := strings.LastIndex(name, "."); i != -1 {
		funcName = name[i+1:]
	}else{
		funcName = name
	}
	payload := LogMessage{
		Ts:       time.Now().UnixNano(),
		Level:    level,
		Pname:    instance.pname,
		Extra:    instance.pextra,
		Source:   filepath.Base(file),
		Function: filepath.Base(funcName),
		Line:     line,
		Message:  msg,
	}

	// Non-blocking 채널 전송
	select {
	case instance.logChan <- payload:
	default:
		// 버퍼 가득 참 (Drop)
	}
}

// @brief 정보 레벨 로그 출력
// @param format 포맷 문자열
// @param v 가변 인자
func Infor(format string, v ...interface{}) {
	logPush("INFO", fmt.Sprintf(format, v...))
}

// @brief 에러 레벨 로그 출력
// @param format 포맷 문자열
// @param v 가변 인자
func Error(format string, v ...interface{}) {
	logPush("ERROR", fmt.Sprintf(format, v...))
}

// @brief 디버그 레벨 로그 출력
// @param format 포맷 문자열
// @param v 가변 인자
func Debug(format string, v ...interface{}) {
	logPush("DEBUG", fmt.Sprintf(format, v...))
}
// @brief 워닝 레벨 로그 출력
// @param format 포맷 문자열
// @param v 가변 인자
func Warn(format string, v ...interface{}) {
	logPush("WARN", fmt.Sprintf(format, v...))
}
