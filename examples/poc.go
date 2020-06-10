package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirfilip/workerpool"
)

type Task struct {
	Name string
}

func (t Task) Run(ctx context.Context) error {
	fmt.Printf("Task %s Starting\n", t.Name)
	select {
	case <-ctx.Done():
		fmt.Printf("Task %s canceled\n", t.Name)
	case <-time.After(10 * time.Second):
		fmt.Printf("Task %s Finished\n", t.Name)
	}
	return nil
}

func main() {
	var err error
	ctx := context.Background()

	wp := workerpool.New(10)
	wp.Start(ctx)

	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)
		task := Task{Name: fmt.Sprintf("w%d", i)}
		err = wp.Execute(task)
		if err != nil {
			panic(err)
		}
	}
	wp.Wait()
	fmt.Println("Done.")
}
