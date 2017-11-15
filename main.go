package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"strings"
	"sync"

	"github.com/fatih/color"

	"github.com/r3labs/verify/git"
)

var exit int

// HasCommit ...
func has(commits []string, id string) bool {
	for _, c := range commits {
		if c == id {
			return true
		}
	}
	return false
}

func syncRepos(repos []string) []*git.Repo {
	var wg sync.WaitGroup
	wg.Add(len(repos))

	r := make([]*git.Repo, len(repos))

	for i := 0; i < len(repos); i++ {

		go func(wg *sync.WaitGroup, repo string, r []*git.Repo, i int) {
			defer wg.Done()
			var err error
			r[i], err = git.Clone(repo, "/tmp/verify/")
			if err != nil {
				panic(err)
			}
		}(&wg, repos[i], r, i)

	}

	wg.Wait()

	return r
}

func removeEmpty(repos []string) []string {
	for i := len(repos) - 1; i >= 0; i-- {
		if strings.TrimSpace(repos[i]) == "" {
			repos = append(repos[:i], repos[i+1:]...)
		}
	}
	return repos
}

func invalid(err string) {
	color.Red("error\n")
	color.Red(err)
	exit = 1
}

func main() {
	inputFile := os.Args[1]

	color.Blue("verifying repositories")
	fmt.Println()
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	repos := strings.Split(string(data), "\n")
	repos = removeEmpty(repos)

	for _, repo := range syncRepos(repos) {
		fmt.Print(repo.Name() + "... ")

		// Verify Default Branch
		defaultBranch, _ := repo.Branch()
		if defaultBranch != "develop" {
			invalid("- develop is not default branch!")
			continue
		}

		// Diff the two commit histories
		diverged, err := repo.Diverged("origin/develop", "origin/master")
		if err != nil {
			invalid("- " + err.Error())
			continue
		}
		if diverged {
			invalid("- environments have diverged https://github.com/" + repo.Path() + "/compare/develop...master?expand=1")
			continue
		}
		color.Green("ok")
	}

	if exit != 0 {
		os.Exit(1)
	}
}
