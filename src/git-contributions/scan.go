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

// scan given a path crawls it and its subfolders
// searching for Git repositories
func scan(path string) {
	fmt.Printf("Found Folders:\n\n")
	repositories := recursiveFolderFind(path)
	addNewSliceElementsToFile(repositories)
	fmt.Printf("\n\nSuccessfully added\n\n")
}

// recursiveFolderFind starts the recursive search of git repositories
// living in the `folder` subtree
func recursiveFolderFind(folder string) []string {
	folders := make([]string, 0)
	return scanGitFolders(folders, folder)
}

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
	defer f.Close()

	// read files in last dir inside folder f
	files, err := f.ReadDir(-1)
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

			// ignore if node_modules or vendor because dir is too large
			if file.Name() == "node_modules" || file.Name() == "vendor" {
				continue
			} else if file.Name() == "Pictures" || file.Name() == "Library" || file.Name() == ".Trash" {
				// another special case on my mac Photos -> no access
				continue
			}

			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}

// addNewSliceElementsToFile given a slice of strings representing paths, stores them
// to the filesystem
func addNewSliceElementsToFile(newRepos []string) {
	oldRepos := parseExistingRepos()
	repoSlice := joinSlice(oldRepos, newRepos)
	updateDotFile(repoSlice)
}

// parseExistingRepos parses the content of each line in `gitFile`
// and returns a slice of strings with the existing repositories.
func parseExistingRepos() []string {
	f := openFile(gitFile)
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		panic(err)
	}

	return lines
}

// openFile opens the file located at `filePath` for reading and writing.
// Creates the file if it does not exist.
func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// file does not exist
			f, err = os.Create(filePath)
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

// updateDotFile writes content to the file in `gitFile`, overwriting existing content.
func updateDotFile(repos []string) {
	content := strings.Join(repos, "\n")
	err := os.WriteFile(gitFile, []byte(content), 0644)
	if err != nil {
		log.Fatalf("error writing to file: %v", err)
	}

	fmt.Print(content)
}

// joinSlice adds elements from `new` slice into the `existing` slice only if
// not already present in `existing`.
func joinSlice(existing []string, new []string) []string {
	for _, file := range new {
		if !sliceContains(existing, file) {
			existing = append(existing, file)
		}
	}
	return existing
}

// sliceContains checks if a `fileName` is present in `repoSlice`.
func sliceContains(repoSlice []string, fileName string) bool {
	for _, file := range repoSlice {
		if file == fileName {
			return true
		}
	}
	return false
}
