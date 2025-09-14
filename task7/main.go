package main

import (
	"fmt"
	"math/rand"
	"time"
)

func asChan(vs ...int) <-chan int {
	c := make(chan int)
	go func() {
		for _, v := range vs {
			c <- v                                                        // Отправляем в канал числа из слайса vs
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // Задерживаем дальнейшее записывание рандомным числом
		}
		close(c) // После заполнения канала закрываем его. Сигнал о том, что в канал больше не будет проводиться запись
	}()
	return c
}

func merge(a, b <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		for {
			// Select будет проверять из какого канала можно считать данные, она будет работать параллельно отправке данных в канал в asChan
			select {
			// Если канал в канале а есть что то для чтения, то может выполниться этот case
			case v, ok := <-a:
				if ok {
					c <- v
				} else {
					a = nil
				}
			// Если в канале b есть что то для чтения, то может выполниться этот кейс
			case v, ok := <-b:
				if ok {
					c <- v
				} else {
					b = nil
				}
			}
			// Когда мы считали и закрыли каналы a и b, то они будут равны nil -> закрываем канал c, больше ничего не будет записываться
			if a == nil && b == nil {
				close(c)
				return
			}
		}
	}()
	return c
}

func main() {
	rand.Seed(time.Now().Unix())
	a := asChan(1, 3, 5, 7)
	b := asChan(2, 4, 6, 8)
	c := merge(a, b)
	for v := range c {
		fmt.Print(v)
	}
}
