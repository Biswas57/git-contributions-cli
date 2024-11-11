package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

// scanGitFolders returns a list of subfolders of `folder` ending with `.git`.
// Returns the base folder of the repo, the .git folder parent.
// Recursively searches in the subfolders by passing an existing `folders` slice.
func scanGitFolders(folders []string, folder string) []string {
	// trim the last `/`
	folder = strings.TrimSuffix(folder, "/")

	// start by opening the folder
	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}

	// read files in last dir inside folder f
	files, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	var path string

	for _, file := range files {
		if file.IsDir() {
			path = folder + "/" + file.Name()

			// if it is a git directory we can add it to the slice of git folders
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				fmt.Print(path + "\n")
				folders = append(folders, path)
				continue
			}

			// ignor if node_modules or vendor because dir is too large like nambo
			if file.Name() == "node_modules" || file.Name() == "vendor" {
				continue
			}

			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}

// recursiveScanFolder starts the recursive search of git repositories
// living in the `folder` subtree
func recursiveFolderFind(folder string) []string {
	folders := make([]string, 0)
	return scanGitFolders(folders, folder)
}

// // findDotFilePath returns the dot file path for the repos list.
// // Creates it and the enclosing folder if it does not exist.
// func findDotFilePath() string {
// 	path, err := os.Getwd()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	dotFile := filepath.Join(path, ".gogitlocalstats")
// 	return dotFile
// }

// parseFileLinesToSlice given a file path string, gets the content
// of each line and parses it to a slice of strings.
func parseExistingRepos() []string {
	f := openFile(gitFile)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines
}

// openFile opens the file located at `filePath`. Creates it if not existing.
func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// file does not exist
			_, err = os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			// other error
			panic(err)
		}
	}

	return f
}

// joinSlice adds the element of the `new` slice
// into the `existing` slice, only if not already there
func joinSlice(existing []string, new []string) []string {
	for _, file := range new {
		if !sliceContains(existing, file) {
			existing = append(existing, file)
		}
	}
	return existing
}

// sliceContains returns bool for if the repo contains a file
func sliceContains(repoSlice []string, fileName string) bool {
	for _, file := range repoSlice {
		if file == fileName {
			return true
		}
	}
	return false
}

// updateDotFile writes content to the file in path `filePath`
// (overwriting existing content)
func updateDotFile(repos []string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(gitFile, []byte(content), 0755)

	fmt.Print(content)
}

// addNewSliceElementsToFile given a slice of strings representing paths, stores them
// to the filesystem
func addNewSliceElementsToFile(newRepos []string) {
	oldRepos := parseExistingRepos()
	repoSlice := joinSlice(oldRepos, newRepos)
	updateDotFile(repoSlice)
}

// scan given a path crawls it and its subfolders
// searching for Git repositories
func scan(path string) {
	fmt.Printf("Found Folders:\n\n")
	repositories := recursiveFolderFind(path)
	addNewSliceElementsToFile(repositories)
	fmt.Printf("\n\nSuccessfully added\n\n")
}
