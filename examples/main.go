package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirfilip/workerpool"
)

type Worker struct {
	Name string
}

func (w Worker) Run(ctx context.Context) error {
	fmt.Printf("Worker %s Starting\n", w.Name)
	select {
	case <-ctx.Done():
		fmt.Printf("Worker %s canceled\n", w.Name)
	case <-time.After(5 * time.Second):
		fmt.Printf("Worker %s Finished\n", w.Name)
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
		worker := Worker{Name: fmt.Sprintf("w%d", i)}
		err = wp.AddWork(worker)
		if err != nil {
			panic(err)
		}
	}
	wp.Shutdown()
	fmt.Println("Done.")
}
