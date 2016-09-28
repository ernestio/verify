package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/r3labs/verify/git"
)

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
	for i := len(repos)-1; i >= 0; i-- {
		if strings.TrimSpace(repos[i]) == "" {
			repos = append(repos[:i], repos[i+1:]...)
		}
	}
	return repos
}


func invalid(err string) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	inputFile := os.Args[1]

	fmt.Println("cloning repos")
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	repos := strings.Split(string(data), "\n")
	repos = removeEmpty(repos)

	for _, repo := range syncRepos(repos) {
		fmt.Println(repo.Name())
		// Verify Default Branch
		defaultBranch, _ := repo.Branch()
		if defaultBranch != "develop" {
			invalid("- develop is not default branch!")
		}

		// Verift commit ID's
		err := repo.Checkout("master")
		if err != nil {
			invalid("- master branch does not exist")
		}

		// master commits
		mc, _ := repo.Commits()

		err = repo.Checkout("develop")
		if err != nil {
			invalid("- develop branch does not exist")
		}

		// develop commits
		dc, _ := repo.Commits()

		// Diff the two commit histories
		if !reflect.DeepEqual(mc, dc[len(dc)-len(mc):]) {
			invalid("- branches have diverged")
		}
	}

}
