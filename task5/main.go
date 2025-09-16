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
	var err error // Объявляется интерфейсная переменная error
	err = test() // Возвращается интерфейсное значение, тип, которого равен *customError, а значение - nil
	// Так как интерфейс имеет только 2 параметра, поэтому условие сработает и выведется "error"
	if err != nil {
		println("error")
		return
	}
	println("ok")
}
