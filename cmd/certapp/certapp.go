package main

import (
	"context"
	"os/signal"
	"sync"

	"github.com/lcahid/idk/internal/service"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background())
	wg := sync.WaitGroup{}
	httpService, err := service.ServiceInit(ctx, &wg)
	if err != nil {
		panic(err)
	}
	defer httpService.Close()
	wg.Wait()
}

