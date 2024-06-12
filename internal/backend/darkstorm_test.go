package backend

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStuff(t *testing.T) {
	for i := 0; i < 50; i++ {
		go func() {
			id, err := uuid.NewV7()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(id.String())
		}()
	}
	time.Sleep(3 * time.Second)
	t.Fatal("end")
}
