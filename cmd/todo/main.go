package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	todo "github.com/sirusxcoder/stmta"
)

func main() {
	add := flag.Bool("add", false, "Add a new todo")
	complete := flag.Int("complete", 0, "Mark a todo as completed")
	del := flag.Int("del", 0, "Deletes a todo")
	list := flag.Bool("list", false, "List all todos")

	flag.Parse()

	todos := &todo.Todos{}

	// Connect to MongoDB
	if err := todos.ConnectToDB(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to DB:", err)
		os.Exit(1)
	}

	switch {
	case *add:
		task, err := getInput(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		if err := todos.Add(task); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to add task:", err)
			os.Exit(1)
		}

	case *complete > 0:
		if err := todos.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to mark task as completed:", err)
			os.Exit(1)
		}

	case *del > 0:
		if err := todos.Delete(*del); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to delete task:", err)
			os.Exit(1)
		}

	case *list:
		if err := todos.Print(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to list tasks:", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintln(os.Stdout, "invalid command")
		os.Exit(0)
	}
}

func getInput(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}

	text := scanner.Text()

	if len(text) == 0 {
		return "", errors.New(`empty todo is not allowed`)
	}
	return text, nil
}
