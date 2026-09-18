package main

import (
    "fmt"
)

const (
    // ANSI escape codes for colors
    Red    = "\033[01;31m"
    Green  = "\033[01;32m"
    Yellow = "\033[01;33m"
    Blue   = "\033[01;34m"
    Reset  = "\033[0m"  // Resets the terminal color
)

func main() {
    fmt.Println(Red + "This text is red! 31" + Reset)
    fmt.Println(Green + "This text is green! 32" + Reset)
    fmt.Println(Yellow + "This text is yellow! 33" + Reset)
    fmt.Println(Blue + "This text is blue! 34" + Reset)
}
