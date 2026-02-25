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




func (i *Interpreter) evaluate() error {
	i.current = 0
	for i.current < len(i.tokens) {
		if err := i.parseStatement(); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) parseStatement() error {
	if i.current >= len(i.tokens) {
		return nil
	}

	token := i.tokens[i.current]

	switch token.Type {
	case "VAR":
		return i.parseVarDecl()
	case "IF":
		return i.parseIf()
	case "FOR":
		return i.parseFor()
	case "FMT.PRINTLN":
		return i.parsePrintln()
	case "IDENTIFIER":
		return i.parseAssignment()
	case "LBRACE":
		return i.parseBlock()
	default:
		i.current++
		return nil
	}
}

func (i *Interpreter) parseVarDecl() error {
	i.current++ 
	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "IDENTIFIER" {
		return fmt.Errorf("expected identifier after 'var'")
	}
	name := i.tokens[i.current].Literal
	i.current++

	
	if i.current >= len(i.tokens) {
		return fmt.Errorf("unexpected end of input")
	}

	typeToken := i.tokens[i.current]
	varType := ""


	if typeToken.Type == "INT" || typeToken.Type == "STRING" || typeToken.Type == "BOOL" {
		varType = typeToken.Literal
		i.current++
	}

	
	if i.current < len(i.tokens) && i.tokens[i.current].Type == "ASSIGN" {
		i.current++ 
		val, err := i.parseExpression()
		if err != nil {
			return err
		}

		if varType == "" {
			varType = val.Type()
		}

		i.env.Set(name, val)
	} else if varType == "" {
		return fmt.Errorf("variable %s must have a type or value", name)
	} else {
		
		switch varType {
		case "int":
			i.env.Set(name, IntValue(0))
		case "string":
			i.env.Set(name, StringValue(""))
		case "bool":
			i.env.Set(name, BoolValue(false))
		}
	}


	if i.current < len(i.tokens) && i.tokens[i.current].Type == "SEMICOLON" {
		i.current++
	}

	return nil
}

func (i *Interpreter) parseAssignment() error {
	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "IDENTIFIER" {
		return fmt.Errorf("expected identifier")
	}
	name := i.tokens[i.current].Literal
	i.current++

	if i.current >= len(i.tokens) {
		return fmt.Errorf("unexpected end of input")
	}

	op := i.tokens[i.current].Type
	if op != "ASSIGN" && op != ":=" {
		return fmt.Errorf("expected assignment operator, got %s", op)
	}
	i.current++

	val, err := i.parseExpression()
	if err != nil {
		return err
	}

	i.env.Set(name, val)


	if i.current < len(i.tokens) && i.tokens[i.current].Type == "SEMICOLON" {
		i.current++
	}

	return nil
}

func (i *Interpreter) parseIf() error {
	i.current++
	condition, err := i.parseExpression()
	if err != nil {
		return err
	}

	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "LBRACE" {
		return fmt.Errorf("expected '{' after if condition")
	}


	condBool := i.toBool(condition)

	if condBool {
		return i.parseBlock()
	}

	skip := 1
	i.current++ 
	for i.current < len(i.tokens) && skip > 0 {
		if i.tokens[i.current].Type == "LBRACE" {
			skip++
		} else if i.tokens[i.current].Type == "RBRACE" {
			skip--
		}
		i.current++
	}


	if i.current < len(i.tokens) && i.tokens[i.current].Type == "ELSE" {
		i.current++
		if i.current >= len(i.tokens) || i.tokens[i.current].Type != "LBRACE" {
			return fmt.Errorf("expected '{' after else")
		}
		return i.parseBlock()
	}

	return nil
}

func (i *Interpreter) parseFor() error {
	i.current++ 


	if i.current < len(i.tokens) && i.tokens[i.current].Type == "LBRACE" {

		return i.parseBlock()
	}


	condition, err := i.parseExpression()
	if err != nil {
		return err
	}

	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "LBRACE" {
		return fmt.Errorf("expected '{' after for condition")
	}


	if i.toBool(condition) {
		return i.parseBlock()
	}

	skip := 1
	i.current++ // skip '{'
	for i.current < len(i.tokens) && skip > 0 {
		if i.tokens[i.current].Type == "LBRACE" {
			skip++
		} else if i.tokens[i.current].Type == "RBRACE" {
			skip--
		}
		i.current++
	}

	return nil
}

func (i *Interpreter) parsePrintln() error {
	i.current++ 
	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "LPAREN" {
		return fmt.Errorf("expected '(' after fmt.Println")
	}
	i.current++ 

	var args []string


	for i.current < len(i.tokens) && i.tokens[i.current].Type != "RPAREN" {
		val, err := i.parseExpression()
		if err != nil {
			return err
		}
		args = append(args, val.String())

		if i.current < len(i.tokens) && i.tokens[i.current].Type == "COMMA" {
			i.current++
		}
	}

	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "RPAREN" {
		return fmt.Errorf("expected ')'")
	}
	i.current++

	output := strings.Join(args, " ")
	fmt.Println(output)
	i.output = append(i.output, output)

	if i.current < len(i.tokens) && i.tokens[i.current].Type == "SEMICOLON" {
		i.current++
	}

	return nil
}

func (i *Interpreter) parseExpression() (Value, error) {
	return i.parseOr()
}

func (i *Interpreter) parseOr() (Value, error) {
	left, err := i.parseAnd()
	if err != nil {
		return nil, err
	}

	for i.current < len(i.tokens) && i.tokens[i.current].Type == "||" {
		i.current++
		right, err := i.parseAnd()
		if err != nil {
			return nil, err
		}
		left = BoolValue(i.toBool(left) || i.toBool(right))
	}

	return left, nil
}

func (i *Interpreter) parseAnd() (Value, error) {
	left, err := i.parseComparison()
	if err != nil {
		return nil, err
	}

	for i.current < len(i.tokens) && i.tokens[i.current].Type == "&&" {
		i.current++
		right, err := i.parseComparison()
		if err != nil {
			return nil, err
		}
		left = BoolValue(i.toBool(left) && i.toBool(right))
	}

	return left, nil
}

func (i *Interpreter) parseComparison() (Value, error) {
	left, err := i.parseAddition()
	if err != nil {
		return nil, err
	}

	for i.current < len(i.tokens) {
		op := i.tokens[i.current].Type
		if op != "==" && op != "!=" && op != "<" && op != ">" && op != "<=" && op != ">=" {
			break
		}
		i.current++

		right, err := i.parseAddition()
		if err != nil {
			return nil, err
		}

		result := i.compare(left, op, right)
		left = BoolValue(result)
	}

	return left, nil
}

func (i *Interpreter) parseAddition() (Value, error) {
	left, err := i.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for i.current < len(i.tokens) {
		op := i.tokens[i.current].Type
		if op != "+" && op != "-" {
			break
		}
		i.current++

		right, err := i.parseMultiplication()
		if err != nil {
			return nil, err
		}

		if op == "+" {
			if sl, ok := left.(StringValue); ok {
				left = StringValue(sl + StringValue(right.String()))
			} else {
				left = IntValue(i.toInt(left) + i.toInt(right))
			}
		} else {
			left = IntValue(i.toInt(left) - i.toInt(right))
		}
	}

	return left, nil
}

func (i *Interpreter) parseMultiplication() (Value, error) {
	left, err := i.parseUnary()
	if err != nil {
		return nil, err
	}

	for i.current < len(i.tokens) {
		op := i.tokens[i.current].Type
		if op != "*" && op != "/" {
			break
		}
		i.current++

		right, err := i.parseUnary()
		if err != nil {
			return nil, err
		}

		if op == "*" {
			left = IntValue(i.toInt(left) * i.toInt(right))
		} else {
			left = IntValue(i.toInt(left) / i.toInt(right))
		}
	}

	return left, nil
}

func (i *Interpreter) parseUnary() (Value, error) {
	if i.current >= len(i.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	op := i.tokens[i.current].Type
	if op == "!" || op == "-" {
		i.current++
		val, err := i.parseUnary()
		if err != nil {
			return nil, err
		}
		if op == "!" {
			return BoolValue(!i.toBool(val)), nil
		} else {
			return IntValue(-i.toInt(val)), nil
		}
	}

	return i.parsePrimary()
}

func (i *Interpreter) parsePrimary() (Value, error) {
	if i.current >= len(i.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	token := i.tokens[i.current]

	switch token.Type {
	case "INT_LIT":
		i.current++
		val, _ := strconv.ParseInt(token.Literal, 10, 64)
		return IntValue(val), nil

	case "STRING_LIT":
		i.current++
		return StringValue(token.Literal), nil

	case "TRUE":
		i.current++
		return BoolValue(true), nil

	case "FALSE":
		i.current++
		return BoolValue(false), nil

	case "IDENTIFIER":
		i.current++
		return i.env.Get(token.Literal)

	case "LPAREN":
		i.current++
		val, err := i.parseExpression()
		if err != nil {
			return nil, err
		}
		if i.current >= len(i.tokens) || i.tokens[i.current].Type != "RPAREN" {
			return nil, fmt.Errorf("expected ')'")
		}
		i.current++
		return val, nil

	default:
		return nil, fmt.Errorf("unexpected token: %s", token.Type)
	}
}

func (i *Interpreter) parseBlock() error {
	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "LBRACE" {
		return fmt.Errorf("expected '{'")
	}
	i.current++ 

	for i.current < len(i.tokens) && i.tokens[i.current].Type != "RBRACE" {
		if err := i.parseStatement(); err != nil {
			return err
		}
	}

	if i.current >= len(i.tokens) || i.tokens[i.current].Type != "RBRACE" {
		return fmt.Errorf("expected '}'")
	}
	i.current++ 

	return nil
}

func (i *Interpreter) toInt(v Value) int64 {
	switch val := v.(type) {
	case IntValue:
		return int64(val)
	case BoolValue:
		if val {
			return 1
		}
		return 0
	case StringValue:
		s, _ := strconv.ParseInt(string(val), 10, 64)
		return s
	}
	return 0
}

func (i *Interpreter) toBool(v Value) bool {
	switch val := v.(type) {
	case BoolValue:
		return bool(val)
	case IntValue:
		return val != 0
	case StringValue:
		return len(val) > 0
	}
	return false
}




func (i *Interpreter) compare(left Value, op string, right Value) bool {
	switch op {
	case "==":
		return left.String() == right.String()
	case "!=":
		return left.String() != right.String()
	case "<":
		return i.toInt(left) < i.toInt(right)
	case ">":
		return i.toInt(left) > i.toInt(right)
	case "<=":
		return i.toInt(left) <= i.toInt(right)
	case ">=":
		return i.toInt(left) >= i.toInt(right)
	}
	return false
}


func (i *Interpreter) Run(input string) error {
	if err := i.Tokenize(input); err != nil {
		return err
	}
	return i.evaluate()
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
