package main

// In this file I’m going to use a dependency called go-git which is available on GitHub at
// https://github.com/src-d/go-git. It abstracts the details of dealing with the Git internal
// representation of commits, exposes a nice API, and it’s self-contained (doesn’t need external
// libs like the libgit2 bindings do), which for my program is a good compromise.

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
)

const outOfRange = 99999
const daysInLastSixMonths = 183
const weeksInLastSixMonths = 26
const gitFile = ".gogitlocalstats"

type column []int

// stats calculates and prints the stats for the given email.
func stats(emailPath string) {
	commits := processRepositories(emailPath)
	printCommitsStats(commits)
}

// processRepositories given a user email, returns the commits made in the last 6 months.
func processRepositories(emailPath string) map[int]int {
	repos := parseExistingRepos()

	commits := make(map[int]int)
	for i := 0; i < daysInLastSixMonths; i++ {
		commits[i] = 0
	}

	for _, path := range repos {
		commits = fillCommits(emailPath, path, commits)
	}

	return commits
}

// fillCommits given a repository found in `path`, gets the commits and puts them in the `commits` map.
func fillCommits(emailPath string, path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}

	iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}

	emails := loadEmails(emailPath)
	offset := calcOffset()
	for {
		commit, err := iterator.Next()
		if err != nil {
			if err == io.EOF {
				break // End of commit history
			}
			panic(err)
		}

		// check for my email in a bunch of emails
		if !checkEmail(commit.Author.Email, emails) {
			continue
		}

		daysAgo := countDaysSinceDate(commit.Author.When) + offset
		if daysAgo != outOfRange {
			commits[daysAgo]++
		}
	}

	return commits
}

// loadEmails loads emails from the "my_emails" file into a slice.
func loadEmails(filePath string) []string {
	var emails []string
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		emails = append(emails, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return emails
}

// checkEmail checks if the given email is in the list of authorized emails.
func checkEmail(email string, emails []string) bool {
	for _, e := range emails {
		if e == email {
			return true
		}
	}
	return false
}

// printCommitsStats prints the commit statistics.
func printCommitsStats(commits map[int]int) {
	keys := sortMapIntoSlice(commits)
	columns := buildCols(keys, commits)
	printCells(columns)
}

// buildCols generates a map with rows and columns ready to be printed on screen.
func buildCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column)
	col := column{}

	for _, k := range keys {
		week := int(k / 7) // Determine week in the last 6 months
		day := k % 7       // Determine day in the week

		if day == 0 {
			col = column{}
		}

		col = append(col, commits[k])

		if day == 6 {
			cols[week] = col
		}
	}

	return cols
}

// printCells prints the cells of the graph.
func printCells(cols map[int]column) {
	printMonths()

	for j := 6; j >= 0; j-- {
		for i := weeksInLastSixMonths + 1; i >= 0; i-- {
			if i == weeksInLastSixMonths+1 {
				printDayCol(j)
			}

			if col, ok := cols[i]; ok {
				if i == 0 && j == calcOffset()-1 {
					printCell(col[j], true)
					continue
				} else {
					if len(col) > j {
						printCell(col[j], false)
						continue
					}
				}
			}
			printCell(0, false)
		}
		fmt.Printf("\n")
	}
}

// sortMapIntoSlice returns a slice of indexes of a map, ordered by key.
func sortMapIntoSlice(m map[int]int) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// Helper Functions
// calcOffset determines and returns the number of days missing to fill the last row of the stats graph.
func calcOffset() int {
	var offset int
	weekday := time.Now().Weekday()

	switch weekday {
	case time.Sunday:
		offset = 7
	case time.Saturday:
		offset = 6
	case time.Friday:
		offset = 5
	case time.Thursday:
		offset = 4
	case time.Wednesday:
		offset = 3
	case time.Tuesday:
		offset = 2
	case time.Monday:
		offset = 1
	}

	return offset
}

// getBeginningOfDay given a time.Time calculates the start time of that day.
func getBeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	t = time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return t
}

// countDaysSinceDate counts how many days have passed since the given `date`.
func countDaysSinceDate(date time.Time) int {
	days := 0
	now := getBeginningOfDay(time.Now())

	for date.Before(now) {
		date = date.Add(24 * time.Hour)
		days++
		if days > daysInLastSixMonths {
			return outOfRange
		}
	}
	return days
}

// printCell given a cell value, prints it with a format based on the value amount and `today` flag.
func printCell(val int, today bool) {
	escape := "\033[0;37;30m"
	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m"
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m"
	case val >= 10:
		escape = "\033[1;30;42m"
	}

	if today {
		escape = "\033[1;37;45m"
	}

	if val == 0 {
		fmt.Printf(escape + "  - " + "\033[0m")
		return
	}

	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}

// printDayCol given the day number (0 is Sunday) prints the day name, alternating rows.
func printDayCol(day int) {
	switch day {
	case 5:
		fmt.Printf(" Fri ")
	case 3:
		fmt.Printf(" Wed ")
	case 1:
		fmt.Printf(" Mon ")
	default:
		fmt.Printf("     ")
	}
}

// printMonths prints the month names in the first line.
func printMonths() {
	week := getBeginningOfDay(time.Now()).Add(-(daysInLastSixMonths * 24 * time.Hour))
	month := week.Month()
	fmt.Printf("         ")

	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("    ")
		}

		week = week.Add(7 * 24 * time.Hour)
		if week.After(time.Now()) {
			break
		}
	}

	fmt.Print("\n")
}
