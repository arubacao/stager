package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os"
	"github.com/gocarina/gocsv"
	"net/url"
	"path"
	"strings"
	"path/filepath"
	"os/exec"
	"github.com/iancoleman/strcase"
)

const (
	configFile = "config.json"
	studentsFile = "students.csv"
)

type Config struct {
	Url string
	Username string
	Password string
	Deadline string
	SquashAfter string `json:"squash_after"`
}

type Student struct {
	Id      string `csv:"id"`
	Name    string `csv:"name"`
}

type Operator interface {
	Run(repo string, config Config) (string, error)
}

type Operation struct {
	Operator Operator
}

func (o *Operation) Operate(repo string, config Config) (string, error) {
	return o.Operator.Run(repo, config)
}

type DeadlineOperation struct{}

func (DeadlineOperation) Run(repo string, config Config) (string, error) {
	lastSha, err := commander("git",
		"-C", repo,
		"log", "-n1", `--pretty=format:"%H"`, `--before="`+config.Deadline+`"`)
	checkError(err)
	checkout, err := commander("git",
		"-C", repo,
		"reset",
		"--hard",
		trimQuote(lastSha))
	fmt.Println(checkout)
	return checkout, err
}

type SquashOperation struct {}

func (SquashOperation) Run(repo string, config Config) (string, error) {
	reset, err := commander("git",
		"-C", repo,
		"reset",
		"--hard",
		config.SquashAfter)
	checkError(err)
	fmt.Println(reset)
	squash, err := commander("git",
		"-C", repo,
		"merge",
		"--squash",
		"HEAD@{1}")
	checkError(err)
	fmt.Println(squash)
	commit, err := commander("git",
		"-C", repo,
		"commit",
		"--no-edit")
	fmt.Println(commit)
	return commit, err
}

func main() {
	config := getConfig(configFile)
	students := getStudents(studentsFile)
	repos := make([]string, 0)
	for _, student := range students {
		repos = append(repos, cloneRepo(config, student))
	}
	operations := []Operation{
		{DeadlineOperation{}},
		{SquashOperation{}},
	}
	for _, repo := range repos {
		for _, operation := range operations {
			operation.Operate(repo, config)
		}
	}
}

func cloneRepo(config Config, student *Student) string {
	repoUrl := fmt.Sprintf(config.Url, config.Username, config.Password, student.Id)
	targetDir := getTargetDirectory(repoUrl, student.Name)
	output, err := commander("git", "clone", repoUrl, targetDir)
	checkError(err)
	fmt.Println("Cloning Repo for: ", student.Name)
	fmt.Println(output)
	return targetDir
}

func getTargetDirectory(repoUrl, studentName string) string {
	u, _ := url.Parse(repoUrl)
	ps := path.Base(u.Path)
	repoBase := strings.TrimSuffix(ps, filepath.Ext(ps))
	return repoBase + "_" + strcase.ToSnake(studentName)
}

func getStudents(filename string) []*Student {
	studentsFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	checkError(err)
	defer studentsFile.Close()

	var students []*Student
	err = gocsv.UnmarshalFile(studentsFile, &students)
	checkError(err)

	return students
}

func getConfig(filename string) Config {
	var config Config
	configContent, err := ioutil.ReadFile(filename)
	checkError(err)
	json.Unmarshal([]byte(configContent), &config)
	return config
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func commander(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()

	return string(output), err
}

func trimQuote(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}