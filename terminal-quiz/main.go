package main

import (
	"flag"
	"fmt"
	"os"
	"math/rand"
	"time"
	"strings"
	"encoding/csv"
)

type problem struct {
	question string
	answer string
}

func readCSV(csvFilePath string) ([][]string, error) {
	var lines [][]string
	f, err := os.Open(csvFilePath)
	if err != nil {
		return lines, err
    }
    defer f.Close()

	lines, err = csv.NewReader(f).ReadAll()
    if err != nil {
        return lines, err
	}
	return lines, nil
}

func parseLines(lines [][]string) []problem {
	var problems = make([]problem, len(lines))
	for i, line := range lines {
		problems[i] = problem{
			question: line[0],
			answer: strings.TrimSpace(line[1]),
		}
	}
	return problems
}

func main() {
	csvFilePath := flag.String("csv", "problems.csv", "a csv file in the format of 'question,answer'")
	timeLimit := flag.Int("limit", 30, "the time limit for the quiz in seconds")
	shuffleMode := flag.Bool("shuffle", false, "shuffle the quiz order each time it is run")
	flag.Parse()

	lines, err := readCSV(*csvFilePath)
	if err != nil {
		fmt.Println("Failed to parse the provided CSV file.")
	}

	problems := parseLines(lines)
	if *shuffleMode == true {
		rand.Shuffle(len(problems), func(i, j int) {
			problems[i], problems[j] = problems[j], problems[i]
		})
	}

	correct := 0
	timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)

	problemloop:
	for index, problem := range problems {
		fmt.Printf("Problem #%d: %s = \n", index + 1, problem.question)

		answerCh := make(chan string)
		go func() {
			var answer string
			fmt.Scanf("%s\n", &answer)
			answerCh <- answer
		}()

		select {
		case <-timer.C:
			fmt.Println()
			break problemloop
		case answer := <-answerCh:
			if problem.answer == answer {
				correct++
			}
		}
	}
	fmt.Printf("You scored %d out of %d.\n", correct, len(problems))
}
