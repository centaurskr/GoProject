package main
import (
	"github.com/google/uuid"
	"fmt"
)

func main() {
	uniqueKey := uuid.New().String()
	fmt.Println(uniqueKey)
}

