package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeWheel(t *testing.T) {
	tw, _ := New()

	fmt.Println(tw.AddTask(5*time.Second, func() error {
		fmt.Println(time.Now())
		return nil
	}))
	fmt.Println(tw.AddCycleTask(1*time.Second, func() error {
		fmt.Println(time.Now())
		return nil
	}))
	tw.Start()
	fmt.Println(time.Now())
	time.Sleep(100 * time.Second)
	tw.Stop()
}
