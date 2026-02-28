# Simple Go Interpreter 

## Описание

**Simple Go Interpreter** — это интерпретатор для упрощенного подмножества языка программирования Go, разработанный с использованием ANTLR4.

Интерпретатор поддерживает следующие конструкции:
- Объявление переменных (`var x int = 5`)
- Короткое объявление (`x := 5`)
- Базовые типы данных: `int`, `string`, `bool`
- Математические операции: `+`, `-`, `*`, `/`
- Операции сравнения: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Логические операции: `&&`, `||`, `!`
- Условные конструкции: `if-else`
- Циклы: `for`
- Вывод на экран: `fmt.Println()`


## Требования к установке

### Системные требования
- Go 1.21 или выше
- Git

### Зависимости
Проект использует ANTLR4 Go runtime:
```
github.com/antlr4-go/antlr/v4 v4.13.0
```

## Инструкции по сборке

### 1. Клонирование/получение проекта
```bash
cd translator
```

### 2. Загрузка зависимостей
```bash
go mod download
go mod tidy
```

### 3. Компиляция (сборка)
```bash
go build -o translator main.go
```

Исполняемый файл `translator` (или `translator.exe` на Windows) будет создан в текущем каталоге.

### 4. Проверка сборки
```bash
# Вывести справку
./translator

# Должно появиться:
# Simple Go Interpreter
# 
# Usage:
#   translator                    - Start REPL
#   translator FILE               - Run file
#   translator -c CODE            - Run code
```
