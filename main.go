package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)
type Value interface {
	Type() string
	String() string
}

type IntValue int64

func (v IntValue) Type() string   { return "int" }
func (v IntValue) String() string { return fmt.Sprintf("%d", v) }


type StringValue string

func (v StringValue) Type() string   { return "string" }
func (v StringValue) String() string { return string(v) }

type BoolValue bool

func (v BoolValue) Type() string   { return "bool" }
func (v BoolValue) String() string { return fmt.Sprintf("%v", v) }

type Environment struct {
	vars map[string]Value
}

func NewEnvironment() *Environment {
	return &Environment{
		vars: make(map[string]Value),
	}
}

func (e *Environment) Set(name string, value Value) {
	e.vars[name] = value
}

func (e *Environment) Get(name string) (Value, error) {
	if v, ok := e.vars[name]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("undefined variable: %s", name)
}

type Interpreter struct {
	env     *Environment
	tokens  []Token
	current int
	output  []string
}

type Token struct {
	Type    string
	Literal string
	Line    int
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		env:     NewEnvironment(),
		tokens:  []Token{},
		current: 0,
		output:  []string{},
	}
}

func (i *Interpreter) Tokenize(input string) error {
	keywords := map[string]bool{
		"var": true, "int": true, "string": true, "bool": true,
		"if": true, "else": true, "for": true, "func": true,
		"fmt.Println": true, "true": true, "false": true,
	}

	line := 1
	for len(input) > 0 {
		if input[0] == ' ' || input[0] == '\t' {
			input = input[1:]
			continue
		}
		if input[0] == '\n' {
			line++
			input = input[1:]
			continue
		}

		if strings.HasPrefix(input, "//") {
			for len(input) > 0 && input[0] != '\n' {
				input = input[1:]
			}
			continue
		}

		if strings.HasPrefix(input, "/*") {
			input = input[2:]
			for len(input) >= 2 && !strings.HasPrefix(input, "*/") {
				if input[0] == '\n' {
					line++
				}
				input = input[1:]
			}
			if len(input) >= 2 {
				input = input[2:]
			}
			continue
		}

		if input[0] == '"' {
			input = input[1:]
			literal := ""
			for len(input) > 0 && input[0] != '"' {
				if input[0] == '\\' && len(input) > 1 {
					input = input[1:]
					switch input[0] {
					case 'n':
						literal += "\n"
					case 't':
						literal += "\t"
					case '\\':
						literal += "\\"
					case '"':
						literal += "\""
					default:
						literal += string(input[0])
					}
					input = input[1:]
				} else {
					literal += string(input[0])
					input = input[1:]
				}
			}
			if len(input) > 0 {
				input = input[1:]
			}
			i.tokens = append(i.tokens, Token{"STRING_LIT", literal, line})
			continue
		}

		if input[0] >= '0' && input[0] <= '9' {
			literal := ""
			for len(input) > 0 && input[0] >= '0' && input[0] <= '9' {
				literal += string(input[0])
				input = input[1:]
			}
			i.tokens = append(i.tokens, Token{"INT_LIT", literal, line})
			continue
		}

		if (input[0] >= 'a' && input[0] <= 'z') || (input[0] >= 'A' && input[0] <= 'Z') || input[0] == '_' {
			literal := ""
			for len(input) > 0 && ((input[0] >= 'a' && input[0] <= 'z') || (input[0] >= 'A' && input[0] <= 'Z') || input[0] >= '0' && input[0] <= '9' || input[0] == '_' || input[0] == '.') {
				literal += string(input[0])
				input = input[1:]
			}

			if keywords[literal] {
				i.tokens = append(i.tokens, Token{strings.ToUpper(literal), literal, line})
			} else {
				i.tokens = append(i.tokens, Token{"IDENTIFIER", literal, line})
			}
			continue
		}

		if len(input) >= 2 {
			twoChar := input[:2]
			switch twoChar {
			case ":=", "==", "!=", "<=", ">=", "&&", "||":
				i.tokens = append(i.tokens, Token{twoChar, twoChar, line})
				input = input[2:]
				continue
			}
		}

		switch input[0] {
		case '(':
			i.tokens = append(i.tokens, Token{"LPAREN", "(", line})
		case ')':
			i.tokens = append(i.tokens, Token{"RPAREN", ")", line})
		case '{':
			i.tokens = append(i.tokens, Token{"LBRACE", "{", line})
		case '}':
			i.tokens = append(i.tokens, Token{"RBRACE", "}", line})
		case '[':
			i.tokens = append(i.tokens, Token{"LBRACKET", "[", line})
		case ']':
			i.tokens = append(i.tokens, Token{"RBRACKET", "]", line})
		case ';':
			i.tokens = append(i.tokens, Token{"SEMICOLON", ";", line})
		case ',':
			i.tokens = append(i.tokens, Token{"COMMA", ",", line})
		case '=':
			i.tokens = append(i.tokens, Token{"ASSIGN", "=", line})
		case '+':
			i.tokens = append(i.tokens, Token{"PLUS", "+", line})
		case '-':
			i.tokens = append(i.tokens, Token{"MINUS", "-", line})
		case '*':
			i.tokens = append(i.tokens, Token{"STAR", "*", line})
		case '/':
			i.tokens = append(i.tokens, Token{"DIV", "/", line})
		case '<':
			i.tokens = append(i.tokens, Token{"LT", "<", line})
		case '>':
			i.tokens = append(i.tokens, Token{"GT", ">", line})
		case '!':
			i.tokens = append(i.tokens, Token{"NOT", "!", line})
		case '.':
			i.tokens = append(i.tokens, Token{"DOT", ".", line})
		default:
			return fmt.Errorf("unexpected character: %c at line %d", input[0], line)
		}
		input = input[1:]
	}

	return nil
}


func main() {
	if len(os.Args) < 2 {
		fmt.Println("Simple Go Interpreter")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  translator                    - Start REPL")
		fmt.Println("  translator FILE               - Run file")
		fmt.Println("  translator -c CODE            - Run code")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  translator -c \"var x int = 5; fmt.Println(x);\"")
		os.Exit(1)
	}

	interp := NewInterpreter()

	if os.Args[1] == "-c" && len(os.Args) > 2 {
		code := os.Args[2]
		if err := interp.Run(code); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		filename := os.Args[1]
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		if err := interp.Run(string(content)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
