package main

import (
	"fmt"
    "log"
	"time"
    "os"

	"net/http"
    "html/template"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error
var templates *template.Template
var tasksTemplate *template.Template
var insertTemplate *template.Template
var editTemplate *template.Template

type Task struct {
    Id int
    Title string
    Content string
    Created string
}

type Context struct {
    Tasks []Task
    Name string
    Search string
    Message string
}

func main() {

	db, err = sql.Open(
        "mysql",
        "root:@9299Feb!!@tcp(127.0.0.1:3306)/todo_database?parseTime=true")

	if err != nil {
		fmt.Println("error occured")
		panic(err.Error())
	}

	defer db.Close()

    PORT := ":8080"

    styles := http.FileServer(http.Dir("./public/templates/stylesheets"))
    http.Handle("/styles/", http.StripPrefix("/styles/", styles))

    http.HandleFunc("/", ShowAllTasksFunc)
    http.HandleFunc("/add", AddTaskFunc)
    http.HandleFunc("/edit/", EditTaskFunc)
    http.HandleFunc("/editresult/", EditResultFunc)
    http.HandleFunc("/delete/", DeleteTaskFunc)
    log.Fatal(http.ListenAndServe(PORT, nil))

}

func ShowAllTasksFunc(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        context := GetTasks()
        PopulateTemplates()
        tasksTemplate.Execute(w, context)
    } else {
        http.Redirect(w, r, "/", http.StatusFound)
    }
}

func GetTasks() Context {
    var task []Task
    var context Context
    var TaskID int
    var TaskContent string
    var TaskTitle string
    var TaskCreated time.Time
    var getTaskSql string
	getTaskSql = "select id, title, context, created_date from task;"

    rows, err := db.Query(getTaskSql)

	if err != nil {
		panic(err.Error())
	}

	defer rows.Close()

    for rows.Next() {

		err = rows.Scan(&TaskID, &TaskTitle, &TaskContent, &TaskCreated)
		if err != nil {
			panic(err.Error())
		}

        TaskCreated = TaskCreated.Local() // convert time to local time

        a := Task {
            Id: TaskID,
            Title: TaskTitle,
            Content: TaskContent,
            Created: TaskCreated.Format(time.UnixDate)[0:20],
        }

        task = append(task, a)

    }
    context = Context{Tasks: task}
    return context
}

func AddTaskFunc(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        insertTemplate.Execute(w, nil)
        return
    }

    r.ParseForm()
    header := r.FormValue("headerName")
    content := r.FormValue("contentName")
    priority := r.FormValue("priority")
    finishDate := r.FormValue("finishDate")

    truth := AddTask(header, content, priority, finishDate)
    if truth != nil {
        log.Fatal("Error adding task")
    }
    http.Redirect(w, r, "/", http.StatusFound)
}

func AddTask(title, content, priority, finishDate string) error {
    stmt, err := db.Prepare(
		`insert into task
	       (
	           title, context, created_date,
	           last_modified_at, finish_date, priority,
	           cat_id, task_status_id, due_date, user_id, hide)
	       values
	       (
	           ?, ?, current_timestamp(),
	           current_timestamp(),
	           ?, ?, ?, ?, NULL, ?, ?
	       );`)

	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(title, content, finishDate, priority, 1, 1, 1, 0)

	if err != nil {
		panic(err.Error())
	}
    return err
}

func EditTaskFunc(w http.ResponseWriter, r *http.Request) {
    var task Task
    var TaskID int
    var TaskContent string
    var TaskTitle string
    var TaskCreated time.Time

    r.ParseForm()
    id := r.FormValue("idvalue")

    getTaskSql := "select id, title, context, created_date from task where id=?;"

    row := db.QueryRow(getTaskSql, id)

	if err != nil {
		// panic(err.Error())

        fmt.Println(err)

	}

    err = row.Scan(&TaskID, &TaskTitle, &TaskContent, &TaskCreated)
    if err != nil {
        panic(err.Error())
    }

    task = Task {
        Id: TaskID,
        Title: TaskTitle,
        Content: TaskContent,
        Created: TaskCreated.Format(time.UnixDate)[0:20],
    }

    editTemplate.Execute(w, task)
}

func EditResultFunc(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    id := r.FormValue("idvalue")
    title := r.FormValue("headerName")
    content := r.FormValue("contentName")

    stmt, err := db.Prepare(`update task set title=?, context=? where id=?;`)

    if err != nil {
        panic(err.Error())
    }

    _, err = stmt.Exec(title, content, id)

	if err != nil {
		panic(err.Error())
	}

    http.Redirect(w, r, "/", http.StatusFound)
}

func DeleteTaskFunc(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    id := r.FormValue("idvalue")
    stmt, err := db.Prepare(`delete from task where id=?`)

    if err != nil {
        panic(err.Error())
    }

    _, err = stmt.Exec(id)

	if err != nil {
		panic(err.Error())
	}

    http.Redirect(w, r, "/", http.StatusFound)
}

func PopulateTemplates() {
    templates = template.Must(template.ParseGlob("./public/templates/*.html"))
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    tasksTemplate = templates.Lookup("tasks.html")
    insertTemplate = templates.Lookup("insert.html")
    editTemplate = templates.Lookup("edit.html")
}
