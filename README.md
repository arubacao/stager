# git-bulk-dl
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE.md)
[![Travis](https://img.shields.io/travis/arubacao/git-bulk-dl.svg?style=flat-square)](https://travis-ci.org/arubacao/git-bulk-dl)
[![Go Report Card](https://goreportcard.com/badge/github.com/arubacao/git-bulk-dl?style=flat-square)](https://goreportcard.com/report/github.com/arubacao/git-bulk-dl)
[![Godoc](https://godoc.org/github.com/arubacao/git-bulk-dl?status.svg&style=flat-square)](http://godoc.org/github.com/arubacao/git-bulk-dl)

`git-bulk-dl` is a tool designed to help [ls1intum](https://wwwbruegge.in.tum.de/lehrstuhl_1/) tutors streamline code homework correction submitted to [ArTEMiS](https://artemis.ase.in.tum.de).
It downloads and prepares a selected list of student repositories to your local machine.

## Features
- Download selected list of student repositories
- Append student names to folder names
- Remove code committed after the deadline 
- Squash student commits into a single commit 
- Dead-simple and reusable configuration
- Automatic build pipeline for cross-platform executables [(see here)](http://github.com/arubacao/git-bulk-dl/releases)

### More Info

#### Append student names to repo folders

![Rename Repos](/assets/rename.png?raw=true "rename repos")

#### Enforce deadline

Some students try to add more commits after the deadline.  
`git-bulk-dl` removes late commits.  
_Note: This is not bullet proof, since git timestamps can be manipulated!_  
![Enforce deadline](/assets/due-date.png?raw=true "Enforce deadline")

#### Squash student commits

Students add code to an existing codebase provided from the instructors.
It is a lot easier to correct their homework, if the changes made by the students are immediately visible.
Therefore, the tools squashes commits after a giving SHA hash. The original history is still available.  
![Squash commits](/assets/squash.png?raw=true "Squash student commits")  

## Install
### Pre-compiled executables (recommended)
Get them [here](http://github.com/arubacao/git-bulk-dl/releases).

### Source
You need `go` installed and `GOBIN` in your `PATH`. Once that is done, run the
command from the repos root folder:
```shell
$ go get -d -t -v ./...
$ go run main.go
```

## Configuration
### config.json

Rename `example.config.json` to `config.json`
```$shell
cp example.config.json config.json
```

```$json
{
  // Copy & paste the url for the repo from https://artemis.ase.in.tum.de
  // The first 2 %s are placeholders for your TUM credentials
  // The last %s is a placeholder for the students LRZ id
  "url": "https://%s:%s@repobruegge.in.tum.de/scm/eist2018l02bumperss03/eist2018-l02-bumpers-sprint03-exercise-%s.git",
  // Your LRZ id
  "username": "tutor-lrz-idga12dub",
  // Your TUM password
  "password": "my-secret-password",
  // The homework deadline in 'Y-m-d H:i:s' (DateTime)  
  "deadline": "2018-29-04 23:59:59",
  // The SHA hash of Stephan Krusches last commit
  "squash_after": "47ad218377d8b2509c6293823cc6ff2f87ca770a"
}
```

### students.csv

Rename `example.students.csv` to `students.csv`
```$shell
cp example.students.csv students.csv
```

Open the first and add your students LRZ ids and names 

```$csv
name,id
John Doe,ga77ugu
...
```

## Usage

1. Place the executable, `config.json` and `students.csv` into a desired folder.
2. a) double click the executable or b) execute from terminal (recommended)
    ```$bash
    $ cd ~/homework3correction
    $ ./git-bulk-dl
    ```
3. ...
4. Profit

## Implementation

The different operations have been implement with the help of the strategy pattern (s. illustration below).
<p><a href="https://commons.wikimedia.org/wiki/File:W3sDesign_Strategy_Design_Pattern_UML.jpg#/media/File:W3sDesign_Strategy_Design_Pattern_UML.jpg"><img src="https://upload.wikimedia.org/wikipedia/commons/4/45/W3sDesign_Strategy_Design_Pattern_UML.jpg" alt="W3sDesign Strategy Design Pattern UML.jpg"></a><br>By <a href="//commons.wikimedia.org/wiki/User:Vanderjoe" title="User:Vanderjoe">Vanderjoe</a> - <span class="int-own-work" lang="en">Own work</span>, <a href="https://creativecommons.org/licenses/by-sa/4.0" title="Creative Commons Attribution-Share Alike 4.0">CC BY-SA 4.0</a>, <a href="https://commons.wikimedia.org/w/index.php?curid=60733582">Link</a></p>

The different operations/strategies are:
1. PullOperation  
    PullOperation ensures that each accessible repository is up-to-date 
    and in a clean state. This is useful for already locally available repo folders.
2. DeadlineOperation  
    DeadlineOperation ensures that commits after a given deadline are not applied in the local repository.
    This is useful, since BitBucket does not enforce any deadline whatsoever.
3. SquashOperation
    SquashOperation squashes all commits after a given SHA hash.
    This is useful to visualise all changes a student made in a single commit.
    
## Todos:  

_This is the first draft of the tool. Bugs are likely to occur._  

- Better error handling