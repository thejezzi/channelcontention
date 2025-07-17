package main

import (
	"sync"
	"testing"
)

func BenchmarkChannelLifecycleParallel(b *testing.B) {
	for b.Loop() {
		var wg sync.WaitGroup
		wg.Add(1000)

		for range 1000 {
			go func() {
				defer wg.Done()

				ch := make(chan int)

				go func() {
					ch <- 42
					close(ch)
				}()

				sum := 0
				for v := range ch {
					sum += v
				}
			}()
		}

		wg.Wait()
	}
}

func BenchmarkChannelPool(b *testing.B) {
	const poolSize = 1000

	pool := make(chan chan int, poolSize)
	for range poolSize {
		pool <- make(chan int)
	}

	for b.Loop() {

		var wg sync.WaitGroup
		wg.Add(poolSize)

		for range poolSize {
			go func() {
				defer wg.Done()

				ch := <-pool

				go func(c chan int) {
					c <- 42
				}(ch)

				sum := 0
				val := <-ch
				sum += val

				pool <- ch
			}()
		}

		wg.Wait()
	}
}

func BenchmarkSyncPoolChannelLifecycle(b *testing.B) {
	const poolSize = 1000

	pool := sync.Pool{
		New: func() any {
			return make(chan int, 1)
		},
	}

	for b.Loop() {
		var wg sync.WaitGroup
		wg.Add(poolSize)

		for range poolSize {
			go func() {
				defer wg.Done()

				ch := pool.Get().(chan int)

				go func(c chan int) {
					c <- 42
				}(ch)

				sum := 0
				val := <-ch
				sum += val

				pool.Put(ch)
			}()
		}

		wg.Wait()
	}
}
