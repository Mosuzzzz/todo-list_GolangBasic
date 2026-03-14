package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"runtime"
	"os/exec"
	"strings"
	"time"
)

type Task struct {
	ID          int         `json:"id"`
	Description string      `json:"description"`
	Status      StatusState `json:"status"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	DeletedAt   string      `json:"deleted_at"`
}
type StatusState string

const (
	StateTodo       StatusState = "todo"
	StateInProgress StatusState = "in_progress"
	StateDone       StatusState = "done"
)

func main() {
	var tasks map[int]Task = ReadFile()
	ShowHelp()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		args := strings.Split(input, " ")
		command := args[0]
		switch command {
		case "add":
			if len(args) < 2 {
				fmt.Println("Usage: add \"task description\"")
				continue
			}

			description := strings.Join(args[1:], " ")

			AddTask(tasks, description)
		case "update":
			if len(args) < 2 {
				fmt.Println("Usage: update <task_id>")
				continue
			}
			id := args[1]
			if len(args) < 3 {
				fmt.Println("Usage: update <task_id> \"new description\"")
				continue
			}
			description := strings.Join(args[2:], " ")
			idInt, err := strconv.Atoi(id)
			if err != nil {
				fmt.Println("Invalid task ID")
				continue
			}
			UpdateTask(tasks, idInt, description)
		case "list":
			listProducts(tasks)
		case "status":
			if len(args) < 3 {
				fmt.Println("Usage: status <task_id> <new_status>")
				continue
			}
			id := args[1]
			status := args[2]
			idInt, err := strconv.Atoi(id)
			if err != nil {
				fmt.Println("Invalid task ID")
				continue
			}
			UpdateStatus(tasks, idInt, StatusState(status))
		case "delete":
			if len(args) < 2 {
				fmt.Println("Usage: delete <task_id>")
				continue
			}
			id := args[1]
			idInt, err := strconv.Atoi(id)
			if err != nil {
				fmt.Println("Invalid task ID")
				continue
			}
			DeleteTask(tasks, idInt)
		case "help":
			ShowHelp()
		case "clear":
			clearScreen()
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Type 'help' for instructions.")
		}
	}
}

func ReadFile() map[int]Task {
	task, err := os.ReadFile("tasks.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return make(map[int]Task)
	}
	taskData := map[int]Task{}
	err = json.Unmarshal(task, &taskData)
	if err != nil {
		fmt.Println("Error unmarshaling data:", err)
		return make(map[int]Task)
	}
	return taskData
}

func savedata(task map[int]Task) {
	taskData, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling data:", err)
	}
	err = os.WriteFile("tasks.json", taskData, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}

func AddTask(task map[int]Task, description string) {

	if description == "" {
		fmt.Println("Description cannot be empty")
		return
	}

	id := getNextID(task)

	task[id] = Task{
		ID:          id,
		Description: description,
		Status:      StateTodo,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	savedata(task)

}

func getNextID(tasks map[int]Task) int {
	maxID := -1
	for id := range tasks {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

func UpdateTask(task map[int]Task, id int, description string) {
	if description == "" {
		fmt.Println("Description cannot be empty")
		return
	}

	if t, ok := task[id]; ok {
		t.Description = description
		t.UpdatedAt = time.Now().Format(time.RFC3339)
		task[id] = t
		savedata(task)
	} else {
		fmt.Println("Task not found")
	}

}

func listProducts(task map[int]Task) {
	if len(task) == 0 {
		fmt.Println("No tasks found")
		return
	}

	KEY := make([]int, 0, len(task))
	for k := range task {
		KEY = append(KEY, k)
	}

	for _, k := range KEY {
		t := task[k]

		if t.DeletedAt != "" {
			continue
		}
		fmt.Printf("ID: %d\nDescription: %s\nStatus: %s\nCreated At: %s\nUpdated At: %s\n\n",
			t.ID, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)
	}

}

func UpdateStatus(task map[int]Task, id int, status StatusState) {
	if !isValidStatus(status) {
		fmt.Println("Invalid status")
		return
	}

	if t, ok := task[id]; ok {
		t.Status = status
		t.UpdatedAt = time.Now().Format(time.RFC3339)
		task[id] = t
		savedata(task)
	} else {
		fmt.Println("Task not found")
	}
}

func isValidStatus(s StatusState) bool {
	switch s {
	case StateTodo, StateInProgress, StateDone:
		return true
	}
	return false
}

func DeleteTask(task map[int]Task, id int) {
	if t, ok := task[id]; ok {

		if t.DeletedAt != "" {
			fmt.Println("Task already deleted")
			return
		}

		t.DeletedAt = time.Now().Format(time.RFC3339)
		task[id] = t

		savedata(task)

		fmt.Println("Task deleted")
	} else {
		fmt.Println("Task not found")
	}
}

func ShowHelp() {
	fmt.Println("Available commands:")
	fmt.Println(" add <description>        - Add new task")
	fmt.Println(" update <id> <desc>       - Update task description")
	fmt.Println(" status <id> <status>     - Change status (todo, in_progress, done)")
	fmt.Println(" delete <id>              - Delete task (soft delete)")
	fmt.Println(" list                     - Show all tasks")
	fmt.Println(" help                     - Show this help message")
	fmt.Println(" clear                    - Clear screen terminal")
	fmt.Println(" exit                     - Exit this program")
	fmt.Println("")
	fmt.Println(`Example:`)
	fmt.Println(` add "Buy milk"`)
	fmt.Println(` update 1 "Buy coffee"`)
	fmt.Println(` status 1 done`)
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Print("\033[2J")
		fmt.Print("\033[H")
	}
}
