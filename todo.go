package todo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/alexeyco/simpletable"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB connection URI
const uri = "mongodb://localhost:27017"

// Database and Collection names
const (
	databaseName   = "todoApp"
	collectionName = "todos"
)

// item represents a single todo item
type item struct {
	Task        string    `bson:"task"`
	Done        bool      `bson:"done"`
	CreatedAt   time.Time `bson:"created_at"`
	CompletedAt time.Time `bson:"completed_at"`
}

// Todos represents a collection of todo items
type Todos struct {
	collection *mongo.Collection
}

// ConnectToDB initializes a connection to the MongoDB collection
func (t *Todos) ConnectToDB() error {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	t.collection = client.Database(databaseName).Collection(collectionName)
	return nil
}

// Add inserts a new task into the MongoDB collection
func (t *Todos) Add(task string) error {
	todo := item{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: time.Time{},
	}

	_, err := t.collection.InsertOne(context.TODO(), todo)
	if err != nil {
		return fmt.Errorf("failed to add task: %v", err)
	}

	return nil
}

// Complete marks a task as done in the MongoDB collection
func (t *Todos) Complete(index int) error {
	todos, err := t.GetAll()
	if err != nil {
		return err
	}

	if index <= 0 || index > len(todos) {
		return fmt.Errorf("invalid index")
	}

	filter := bson.M{"task": todos[index-1].Task}
	update := bson.M{
		"$set": bson.M{
			"done":         true,
			"completed_at": time.Now(),
		},
	}

	_, err = t.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to complete task: %v", err)
	}

	return nil
}

// Delete removes a task from the MongoDB collection
func (t *Todos) Delete(index int) error {
	todos, err := t.GetAll()
	if err != nil {
		return err
	}

	if index <= 0 || index > len(todos) {
		return fmt.Errorf("invalid index")
	}

	filter := bson.M{"task": todos[index-1].Task}
	_, err = t.collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}

	return nil
}

// GetAll retrieves all tasks from the MongoDB collection
func (t *Todos) GetAll() ([]item, error) {
	var todos []item

	cursor, err := t.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tasks: %v", err)
	}

	if err = cursor.All(context.TODO(), &todos); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %v", err)
	}

	return todos, nil
}

// Print displays all tasks in the terminal

func (t *Todos) Print() error {
	clearTerminal()

	// Print header information with color styling
	fmt.Println("\n\n")
	fmt.Println("\x1b[92m" + `
		        ╔═════════════════════════════════════════════╗
		        ║              Welcome to STMTA!              ║
		        ║ Simple Task Management Terminal Application ║
		        ╚═════════════════════════════════════════════╝
	` + "\x1b[0m")

	// Main Table
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "#"},
			{Align: simpletable.AlignCenter, Text: "Task#"},
			{Align: simpletable.AlignCenter, Text: "Done?"},
			{Align: simpletable.AlignCenter, Text: "CreatedAt"},
			{Align: simpletable.AlignCenter, Text: "CompletedAt"},
		},
	}

	todos, err := t.GetAll()
	if err != nil {
		return err
	}

	var cells [][]*simpletable.Cell

	for idx, item := range todos {
		idx++

		task := blue(item.Task)
		done := blue("no")

		createdAtStr := item.CreatedAt.Format(time.RFC822)
		completedAtStr := item.CompletedAt.Format(time.RFC822)

		if item.Done {
			task = green(fmt.Sprintf("\u2705 %s", task))
			done = green("yes")

			createdAtStr = gray(createdAtStr)
			completedAtStr = gray(completedAtStr)
		}

		cells = append(cells, []*simpletable.Cell{
			{Text: fmt.Sprintf("%d", idx)},
			{Text: task},
			{Text: done},
			{Text: createdAtStr},
			{Text: completedAtStr},
		})
	}
	table.Body = &simpletable.Body{Cells: cells}

	// Handle CountPending error
	countPending, err := t.CountPending()
	if err != nil {
		return err
	}

	table.Footer = &simpletable.Footer{Cells: []*simpletable.Cell{
		{Align: simpletable.AlignCenter, Span: 5, Text: red(fmt.Sprintf("You have %d pending todos", countPending))},
	}}

	table.SetStyle(simpletable.StyleUnicode)
	table.Println()

	// Application Footer with color styling
	fmt.Println()
	fmt.Println("\x1b[90m© Powered by SirusWebTech LLC\x1b[0m")

	return nil
}

// CountPending counts the number of pending tasks
func (t *Todos) CountPending() (int, error) {
	todos, err := t.GetAll()
	if err != nil {
		return 0, err
	}

	total := 0
	for _, item := range todos {
		if !item.Done {
			total++
		}
	}
	return total, nil
}

// clearTerminal clears the terminal screen
func clearTerminal() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default: // Unix-based systems (Linux, macOS)
		fmt.Print("\x1b[2J\x1b[H")
	}
}
