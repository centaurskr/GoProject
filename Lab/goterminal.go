package main

import (
    "fmt"
    "os"
    "unicode"
    "golang.org/x/term"
)

const (
    ESC   = 27
    ENTER = 13
    BS    = 8
    CURUP = 'A'
    CURDN = 'B'
    CURRT = 'C'
    CURLF = 'D'
    F10   = '0'
    F3    = '3'
    F4    = '4'
)

func SetRawMode() (func(), error) {
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        return nil, err
    }
    return func() { term.Restore(int(os.Stdin.Fd()), oldState) }, nil
}

func GetChar() []byte {
    var ch [1]byte
    _, err := os.Stdin.Read(ch[:])
    if err != nil {
        return []byte{0} // Return 0 to indicate an error
    }

    if ch[0] == ESC {
        var seq [2]byte
        os.Stdin.Read(seq[:]) // Read next 2 bytes
        if seq[0] == '[' {
            return []byte{seq[1]} // Return detected escape sequence
        }
        return []byte{ESC} // Unrecognized, return ESC
    }
    return ch[:]
}

func XYGetString(x, y, maxLen int, msg string) (string, int) {
    fmt.Printf("\033[%d;%dH%s", y, x, msg) // Move cursor & print msg

    var buff []rune
    pos := 0

    for {
        // Update input field
        fmt.Printf("\033[%d;%dH\033[K%s", y, x+len(msg), string(buff))
        fmt.Printf("\033[%d;%dH", y, x+len(msg)+pos)

        ch := GetChar()
        if len(ch) == 0 {
            continue
        }

        switch ch[0] {
        case ENTER:
            return string(buff), pos
        case ESC:
            return "", F3
        case BS:
            if pos > 0 {
                buff = append(buff[:pos-1], buff[pos:]...)
                pos--
            }
        case ' ':
            if pos < maxLen {
                buff = append(buff[:pos], append([]rune{' '}, buff[pos:]...)...)
                pos++
            }
        default:
            if unicode.IsPrint(rune(ch[0])) && pos < maxLen {
                buff = append(buff[:pos], append([]rune{rune(ch[0])}, buff[pos:]...)...)
                pos++
            }
        }
    }
}

func main() {
    reset, err := SetRawMode()
    if err != nil {
        fmt.Printf("%v\n", err)
        os.Exit(1)
    }
    defer reset()

    fmt.Print("\033[2J") // Clear screen
    input, length := XYGetString(10, 5, 30, "Enter your name: ")
    fmt.Printf("\nYou entered: %s (Length: %d)\n", input, length)
}
