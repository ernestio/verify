package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"github.com/fatih/color"
//	"reflect"
	"strings"
	"sync"

	"github.com/r3labs/verify/git"
)

// HasCommit ...
func has(commits[]string, id string) bool {
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
	for i := len(repos)-1; i >= 0; i-- {
		if strings.TrimSpace(repos[i]) == "" {
			repos = append(repos[:i], repos[i+1:]...)
		}
	}
	return repos
}

func invalid(err string) {
	color.Red(err)
	//os.Exit(1)
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
		color.Blue(repo.Name())
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

		// develop commits matched with master
		dcm := dc[len(dc)-len(mc):]

		// Diff the two commit histories
		var missingOnMaster []string
		var missingOnDevelop []string

		for _, c := range mc {
			if has(dcm, c) {
				continue
			}
			missingOnDevelop = append(missingOnDevelop, c)
		}

		for _, c := range dcm {
			if has(mc, c) {
				continue
			}
			missingOnMaster = append(missingOnMaster, c)
		}
		
		if len(missingOnDevelop) > 0 || len(missingOnMaster) > 0 {
			invalid("- branches have diverged")
			color.Cyan("    missing on develop:")
			for _, c := range missingOnDevelop {
				color.Magenta("      " + c)			
			}
			color.Cyan("    missing on master:")
			for _, c := range missingOnMaster {
				color.Magenta("      " + c)				
			}
		}

		fmt.Println()
	}

}
