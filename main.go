/*
 * Copyright 2018 Christopher Lass
 *
 * MIT Licence
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/iancoleman/strcase"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	configFile   = "config.json"
	studentsFile = "students.csv"
)

// Config holds data from config.json
type Config struct {
	URL         string
	Username    string
	Password    string
	Deadline    string
	SquashAfter string `json:"squash_after"`
}

// Student holds data from students.csv
type Student struct {
	ID   string `csv:"id"`
	Name string `csv:"name"`
}

// Operator Interface : Implementation of an interchangeable operator object that operates on the students context
type Operator interface {
	Run(repo string, student Student, config Config) (string, error)
}

// Operation struct
type Operation struct {
	Operator Operator
}

// Operate func adds verbosity and calls run func
func (o *Operation) Operate(repo string, student Student, config Config) (string, error) {
	fmt.Println("--------------------------------------------------------------")
	fmt.Printf("Execute %s for %s \n", reflect.TypeOf(o.Operator), student.Name)
	fmt.Println("--------------------------------------------------------------")

	return o.Operator.Run(repo, student, config)
}

// PullOperation ensures that each accessible repository is up-to-date
// and in a clean state. This is useful for already locally available repo folders.
type PullOperation struct{}

// Run executes the PullOperation
func (PullOperation) Run(repo string, student Student, config Config) (string, error) {
	fetch, err := commander("git",
		"-C", repo,
		"fetch", "--all")
	fmt.Println(fetch)
	checkGitError(fetch, err)
	reset, err := commander("git",
		"-C", repo,
		"reset", "--hard", "origin/master")
	fmt.Println(reset)
	checkGitError(reset, err)

	return reset, err
}

// DeadlineOperation ensures that commits after a given deadline are not applied in the local repository.
// This is useful, since BitBucket does not enforce any deadline whatsoever.
type DeadlineOperation struct{}

// Run executes the DeadlineOperation
func (DeadlineOperation) Run(repo string, student Student, config Config) (string, error) {
	lastSha, err := commander("git",
		"-C", repo,
		"log", "-n1", `--pretty=format:"%H"`, `--before="`+config.Deadline+`"`)
	fmt.Println(lastSha)
	checkGitError(lastSha, err)
	checkout, err := commander("git",
		"-C", repo,
		"reset",
		"--hard",
		trimQuote(lastSha))
	fmt.Println(checkout)
	checkGitError(checkout, err)

	return checkout, err
}

// SquashOperation squashes all commits after a given SHA hash.
// This is useful to visualise all changes a student made in a single commit.
type SquashOperation struct{}

// Run executes the SquashOperation
func (SquashOperation) Run(repo string, student Student, config Config) (string, error) {
	reset, err := commander("git",
		"-C", repo,
		"reset",
		"--hard",
		config.SquashAfter)
	fmt.Println(reset)
	checkGitError(reset, err)
	squash, err := commander("git",
		"-C", repo,
		"merge",
		"--squash",
		"HEAD@{1}")
	fmt.Println(squash)
	checkGitError(squash, err)
	commit, err := commander("git",
		"-C", repo,
		"commit",
		"--no-edit")
	fmt.Println(commit)
	checkGitError(commit, err)

	return commit, err
}

func main() {
	config := getConfig(configFile)
	students := getStudents(studentsFile)
	operations := []Operation{
		{PullOperation{}},
		{DeadlineOperation{}},
		{SquashOperation{}},
	}
	for _, student := range students {
		repo := cloneRepo(config, student)

		if info, err := os.Stat(repo); err != nil || !info.IsDir() {
			fmt.Printf("Student: %s - No local repository: %s \n", student.Name, repo)
			continue
		}

		for _, operation := range operations {
			operation.Operate(repo, student, config)
		}
	}
}

func cloneRepo(config Config, student Student) string {
	fmt.Println("Cloning Repo for: ", student.Name)
	repoURL := fmt.Sprintf(config.URL, config.Username, config.Password, student.ID)
	targetDir := getTargetDirectory(repoURL, student.Name)
	output, err := commander("git", "clone", repoURL, targetDir)
	checkGitError(output, err)
	fmt.Println(output)

	return targetDir
}

func getTargetDirectory(repoURL, studentName string) string {
	u, _ := url.Parse(repoURL)
	ps := path.Base(u.Path)
	repoBase := strings.TrimSuffix(ps, filepath.Ext(ps))

	return repoBase + "_" + strcase.ToSnake(studentName)
}

func getStudents(filename string) []Student {
	studentsFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	checkError(err)
	defer studentsFile.Close()

	var students []Student
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

func checkGitError(message string, err error) {
	if err == nil {
		return
	}
	if strings.Contains(message, "already exists") {
		return
	}
	if strings.Contains(message, "does not exist") {
		return
	}
	if strings.Contains(message, "nothing to commit, working tree clean") {
		return
	}
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func commander(executable string, args ...string) (string, error) {
	fmt.Printf("Execute: %s %s \n", executable, strings.Join(args, " "))
	cmd := exec.Command(executable, args...)
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
