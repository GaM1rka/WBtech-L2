package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

// Функция, которая возвращает наш собственный тип customError
func test() *customError {
	// ... do something
	return nil
}

func main() {
	var err error // Объявляется переменная с типом error
	err = test()
	if err != nil {
		println("error")
		return
	}
	println("ok")
}
