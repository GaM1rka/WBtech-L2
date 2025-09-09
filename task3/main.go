package main

import (
	"fmt"
	"os"
)

// Объявление функции, которая возвращает интерфейс error
func Foo() error {
	var err *os.PathError = nil // Создается переменная с типом *os.PathError
	return err
}

func main() {
	err := Foo()
	fmt.Println(err)        // Выведет <nil>, значение, которое получено методом Error(), оно вызывается для err
	fmt.Println(err == nil) // Интерфейс хранит 2 параметра [type, value], поэтому такое сравнение выведет false
}
