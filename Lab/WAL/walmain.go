package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	//"path/filepath"
	"time"
)

const (
	walFileName     = "wal.log"
	snapshotName    = "snapshot.json"
	recordHdrLength = 4 // uint32 length prefix
	crcLen          = 4 // uint32 crc
	commitFlagLen   = 1 // 1 byte commit flag
)

// WAL record payload structure
type WALRecord struct {
	Seq   uint64 `json:"seq"`
	Op    string `json:"op"` // "set" or "del"
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	Ts    int64  `json:"ts"`
}

type WAL struct {
	f   *os.File
	seq uint64
}

func OpenWAL(path string) (*WAL, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	// move offset to end for appends
	_, err = f.Seek(0, io.SeekEnd)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &WAL{f: f, seq: 0}, nil
}

func (w *WAL) Close() error {
	if w.f == nil {
		return nil
	}
	return w.f.Close()
}

// Append a record and fully commit it (two-step: write with commit=0, sync, set commit=1, sync)
// If simulateNoCommit == true: it writes the first part (commit=0 & synced) and returns WITHOUT setting commit=1.
// This simulates a crash between the two steps.
func (w *WAL) Append(rec WALRecord, simulateNoCommit bool) error {
	rec.Seq++
	rec.Ts = time.Now().UnixNano()

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	// record layout:
	// [4 bytes length L][L bytes data][4 bytes CRC32(data)][1 byte commitFlag]
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	if _, err := buf.Write(data); err != nil {
		return err
	}
	crc := crc32.ChecksumIEEE(data)
	if err := binary.Write(&buf, binary.BigEndian, crc); err != nil {
		return err
	}
	// commit flag: 0 initially
	if err := buf.WriteByte(0); err != nil {
		return err
	}

	// append to file
	// write full record with commit=0
	if _, err := w.f.Write(buf.Bytes()); err != nil {
		return err
	}
	if err := w.f.Sync(); err != nil {
		return err
	}

	if simulateNoCommit {
		// simulate crash before final commit - leave commit flag 0
		return nil
	}

	// set commit flag to 1: seek to last byte and write 1, then sync
	off, err := w.f.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	// last byte of file is at off-1 (current pos after write)
	if _, err := w.f.WriteAt([]byte{1}, off-1); err != nil {
		return err
	}
	if err := w.f.Sync(); err != nil {
		return err
	}
	return nil
}

// Recover scans WAL file and returns reconstructed key-value map based on committed records only.
func Recover(path string) (map[string]string, error) {
	res := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		// if file missing, return empty state
		if os.IsNotExist(err) {
			return res, nil
		}
		return nil, err
	}
	defer f.Close()

	for {
		// read length
		var l uint32
		if err := binary.Read(f, binary.BigEndian, &l); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// truncated header -> stop
				break
			}
			return nil, err
		}
		data := make([]byte, l)
		if _, err := io.ReadFull(f, data); err != nil {
			// truncated data -> stop
			break
		}
		var crc uint32
		if err := binary.Read(f, binary.BigEndian, &crc); err != nil {
			// truncated crc -> stop
			break
		}
		commitFlag := make([]byte, 1)
		if _, err := io.ReadFull(f, commitFlag); err != nil {
			// truncated commit flag -> stop
			break
		}
		// only apply if committed == 1
		if commitFlag[0] != 1 {
			// ignore uncommitted/partially written
			continue
		}
		// verify crc
		if crc != crc32.ChecksumIEEE(data) {
			// corrupted -> skip
			continue
		}
		var rec WALRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			// malformed -> skip
			continue
		}
		// apply op
		switch rec.Op {
		case "set":
			res[rec.Key] = rec.Value
		case "del":
			delete(res, rec.Key)
		default:
			// ignore unknown op
		}
	}
	return res, nil
}

func saveSnapshot(state map[string]string, path string) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(state); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func printUsage() {
	fmt.Println("WAL test program")
	fmt.Println("Usage examples:")
	fmt.Println("  go run main.go append -op set -k foo -v bar        # append and commit a set")
	fmt.Println("  go run main.go append -op set -k foo -v crash      # append but simulate crash before commit")
	fmt.Println("  go run main.go recover                             # recover and show state")
	fmt.Println("  go run main.go show                                # show snapshot if exists")
	fmt.Println("  go run main.go reset                               # remove wal and snapshot")
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	cmd := os.Args[1]

	// flags for append
	appendFlagSet := flag.NewFlagSet("append", flag.ExitOnError)
	op := appendFlagSet.String("op", "set", "operation: set/del")
	key := appendFlagSet.String("k", "", "key")
	val := appendFlagSet.String("v", "", "value")
	noCommit := appendFlagSet.Bool("no-commit", false, "simulate no commit (crash before commit)")

	switch cmd {
	case "append":
		appendFlagSet.Parse(os.Args[2:])
		if *key == "" {
			fmt.Println("key required")
			return
		}
		w, err := OpenWAL(walFileName)
		if err != nil {
			fmt.Println("open wal:", err)
			return
		}
		defer w.Close()

		rec := WALRecord{
			Op:    *op,
			Key:   *key,
			Value: *val,
		}
		simulate := *noCommit
		if *val == "crash" {
			// convenience: -v crash acts like no-commit
			simulate = true
		}
		if err := w.Append(rec, simulate); err != nil {
			fmt.Println("append:", err)
			return
		}
		if simulate {
			fmt.Println("Appended (UNCOMMITTED). Simulating crash: exiting without commit.")
			// exit immediately to emulate crash (we already returned from Append without final commit)
			os.Exit(1)
		}
		fmt.Println("Appended and committed.")
	case "recover":
		state, err := Recover(walFileName)
		if err != nil {
			fmt.Println("recover error:", err)
			return
		}
		fmt.Println("Recovered state (in-memory):")
		for k, v := range state {
			fmt.Printf("  %s = %s\n", k, v)
		}
		// save snapshot
		if err := saveSnapshot(state, snapshotName); err != nil {
			fmt.Println("failed save snapshot:", err)
		} else {
			fmt.Println("Snapshot saved to", snapshotName)
		}
	case "show":
		// show snapshot if exists, else recover
		b, err := os.ReadFile(snapshotName)
		if err == nil {
			fmt.Println("Snapshot content:")
			fmt.Println(string(b))
			return
		}
		fmt.Println("Snapshot not found; try recover to regenerate.")
	case "reset":
		os.Remove(walFileName)
		os.Remove(snapshotName)
		fmt.Println("Removed wal and snapshot")
	default:
		printUsage()
	}
}

