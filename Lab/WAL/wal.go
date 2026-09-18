package main

import (
    "bufio"
    "compress/gzip"
    "encoding/json"
    "fmt"
    "io"
    "os"
    //"path/filepath"
    "sync"
    "time"
)

type WAL struct {
    filename     string
    rolloverSize int64
    file         *os.File
    writer       *bufio.Writer
    mu           sync.Mutex
}

func (w *WAL) Init(filename string, rolloverSize int64) error {
    w.filename = filename
    w.rolloverSize = rolloverSize

    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return fmt.Errorf("failed to open WAL file: %w", err)
    }
    if _, err := file.Seek(0, io.SeekEnd); err != nil {
        file.Close()
        return fmt.Errorf("failed to seek to end: %w", err)
    }

    w.file = file
    w.writer = bufio.NewWriter(file)

    go w.run()
    return nil
}

func (w *WAL) WriteJSON(data interface{}) error {
    w.mu.Lock()
    defer w.mu.Unlock()

    enc := json.NewEncoder(w.writer)
    if err := enc.Encode(data); err != nil {
        return err
    }
    return w.writer.Flush()
}

func (w *WAL) run() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        w.mu.Lock()
        fi, err := w.file.Stat()
        if err == nil && fi.Size() >= w.rolloverSize {
            w.rollover()
        }
        w.mu.Unlock()
    }
}

func (w *WAL) rollover() error {
    oldName := w.filename
    ts := time.Now().Format("20060102_150405")
    newName := fmt.Sprintf("%s.%s", oldName, ts)

    // Close current file
    w.writer.Flush()
    w.file.Close()

    // Rename
    if err := os.Rename(oldName, newName); err != nil {
        return err
    }

    // Compress old file
    go compressFile(newName)

    // Reopen new WAL file
    file, err := os.OpenFile(oldName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
    if err != nil {
        return err
    }
    w.file = file
    w.writer = bufio.NewWriter(file)
    return nil
}

func compressFile(filename string) {
    in, err := os.Open(filename)
    if err != nil {
        return
    }
    defer in.Close()

    out, err := os.Create(filename + ".gz")
    if err != nil {
        return
    }
    defer out.Close()

    gw := gzip.NewWriter(out)
    defer gw.Close()
    io.Copy(gw, in)

    os.Remove(filename)
}

func (w *WAL) ReadAll() ([]map[string]interface{}, error) {
    w.mu.Lock()
    defer w.mu.Unlock()

    w.writer.Flush()

    file, err := os.Open(w.filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var logs []map[string]interface{}
    dec := json.NewDecoder(file)
    for {
        var obj map[string]interface{}
        if err := dec.Decode(&obj); err != nil {
            if err == io.EOF {
                break
            }
            return nil, err
        }
        logs = append(logs, obj)
    }
    return logs, nil
}

func main() {
    var logger WAL
    if err := logger.Init("sample.wal", 1024*10); err != nil { // 10KB rollover
        panic(err)
    }

    // JSON 쓰기
    for i := 0; i < 5; i++ {
        record := map[string]interface{}{
            "TR":     "ORDER",
            "COIN":   "BTC.USDT",
            "CHANNEL":"CN02",
            "MKIDX":  fmt.Sprintf("%d", 4441+i),
            "MKCIDX": fmt.Sprintf("%d", i),
            "SEQ":    i,
            "TIME":   time.Now().Format(time.RFC3339),
        }
        if err := logger.WriteJSON(record); err != nil {
            fmt.Println("Write error:", err)
        }
        time.Sleep(time.Second)
    }

    // WAL 내용 읽기
    logs, err := logger.ReadAll()
    if err != nil {
        panic(err)
    }
    fmt.Println("=== WAL ReadAll ===")
    for _, log := range logs {
        fmt.Println(log)
    }

    // 잠시 대기 (롤오버/압축 확인 가능)
    fmt.Println("프로그램이 15초 동안 대기합니다. 로그 파일을 확인해보세요.")
    time.Sleep(15 * time.Second)
}
