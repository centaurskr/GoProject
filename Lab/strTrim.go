package main

import (
    "fmt"
    "strings"
)

func main() {
    s := "example/com/"
    trimmed := strings.TrimSuffix(s, "/")
    fmt.Printf("%s.TrimSuffix(/)-> %s\n", s, trimmed) // Output: example.com
}

