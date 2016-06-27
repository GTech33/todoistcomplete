package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/fatih/color"
)

type Task struct {
	DateCompleted string
	Name          string
	Project       string
}

type Tasks []Task

func (slice Tasks) Len() int {
	return len(slice)
}

func (slice Tasks) Less(i, j int) bool {
	return slice[i].Project < slice[j].Project
}

func (slice Tasks) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func filter(vs []Task, f func(string) bool) []Task {
	var vsf []Task
	for _, v := range vs {
		if f(v.Project) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func main() {
	sinceDate := flag.String("SinceDate", "", "Completed tasks since a date. Date must be 01/02/2006 format.")
	untilDate := flag.String("UntilDate", "", "Completed tasks until a date. Date must be 01/02/2006 format.")
	token := flag.String("Token", "", "Todoist API token")

	flag.Parse()

	sinceFormat, err := time.Parse("01/02/2006", *sinceDate)
	untilFormat, err := time.Parse("01/02/2006", *untilDate)
	since := sinceFormat.Format("2006-01-02T15:04")
	until := untilFormat.Format("2006-01-02T") + "23:59"

	// fmt.Println("sinceDate:", since)
	// fmt.Println("untilDate:", until)

	body := url.Values{}
	body.Add("token", *token)
	body.Add("limit", "50")
	body.Add("since", since)
	body.Add("until", until)

	client := &http.Client{}
	request, err := http.NewRequest("POST", "https://todoist.com/API/v7/get_all_completed_items", strings.NewReader(body.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	data, err := simplejson.NewFromReader(response.Body)
	items, err := data.Get("items").Array()
	projects, err := data.Get("projects").Map()

	projectNameLookup := make(map[string]string)
	for projectID, v := range projects {
		p := v.(map[string]interface{})
		projectNameLookup[projectID] = p["name"].(string)
	}

	var tasks []Task
	for _, item := range items {
		n := item.(map[string]interface{})
		tasks = append(tasks, Task{
			DateCompleted: n["completed_date"].(string),
			Name:          n["content"].(string),
			Project:       projectNameLookup[(n["project_id"].(json.Number)).String()],
		})
	}

	sort.Sort(Tasks(tasks))
	last := ""
	var projectTasks []Task
	for _, t := range tasks {
		if !strings.EqualFold(last, t.Project) {
			color.Set(color.FgRed)
			last = t.Project
			fmt.Println(last)
			projectTasks = filter(tasks, func(projectName string) bool {
				return strings.EqualFold(projectName, last)
			})
			for _, task := range projectTasks {
				fmt.Printf("\t%s (%s)\n", task.Name, task.DateCompleted)
			}
			color.Unset()
		}
	}

	if err != nil {
		fmt.Println(err)
	}
}
