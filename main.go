package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const defaultCommitTemplate = "{t}({p,}) - {n} {m}"

var (
	ErrNoCommitFileProvided = fmt.Errorf("no commit file provided")
	ErrFailedToTruncateFile = fmt.Errorf("")

	messagePlaceholderRegex = regexp.MustCompile("{(t|n|p|m|i|s)([^}]+)?}")
)

type messageRequirement int

const (
	requirementCommitType messageRequirement = iota
	requirementTicketNumber
	requirementAffectedPackages
	requirementMessage
)

func NewCommitMessage(format string) (string, error) {
	componentCbs := map[string]func(string) (string, error){
		"t": func(_ string) (string, error) { return promptCommitType() },
		"p": func(extra string) (string, error) { return promptRelevantPackages(extra) },
		"n": func(_ string) (string, error) { return promptTicketNumber() },
		"m": func(_ string) (string, error) { return promptMessage("Commit message: ") },

		"i": func(extra string) (string, error) { return promptInt(extra) },
		"s": func(extra string) (string, error) { return promptString(extra) },
	}

	if format == "" {
		format = "{m}"
	}
	parts := messagePlaceholderRegex.FindAllStringSubmatch(format, -1)
	if len(parts) == 0 {
		return "", fmt.Errorf("no template found in commit message format")
	}
	components := []interface{}{}
	for i, _ := range parts {
		if componentCb, found := componentCbs[parts[i][1]]; !found {
			return "", fmt.Errorf("component %v not found", parts[i])
		} else {
			val, err := componentCb(parts[i][2])
			if err != nil {
				return "", err
			}
			components = append(components, val)
		}
	}
	format = messagePlaceholderRegex.ReplaceAllString(format, "%s")
	message := fmt.Sprintf(format, components...)
	message = strings.Join(strings.Fields(message), " ")
	return message, nil
}

/*
Usage: commit-message FILE
*/
func main() {
	f, err := getCommitFileFromArgs()
	if err != nil {
		log.Fatalf("an error occurred: %s", err)
	}
	template, err := getCommitTemplate()
	if err != nil {
		log.Fatalf("an error occurred: %s", err)
	}
	var msg string
	msg, err = NewCommitMessage(template)
	if err != nil {
		log.Fatalf("an error occurred: %s", err)
	}

	err = writeMessageToFile(f, msg)
	if err != nil {
		log.Fatalf("an error occurred: %s", err)
	}
}

func getCommitFileFromArgs() (*os.File, error) {
	if len(os.Args) == 1 {
		return nil, ErrNoCommitFileProvided
	}
	filePath := os.Args[1]
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %s", filePath, err)
	}
	if err := os.Truncate(filePath, 0); err != nil {
		return nil, fmt.Errorf("failed to truncate %s: %s", filePath, err)
	}
	return f, nil
}

func getCommitTemplate() (string, error) {
	app := "git rev-parse --show-toplevel"
	cmd := exec.Command("bash", "-c", app)
	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error while reading modified files: %v", err)
	}
	templateFile := strings.TrimSpace(string(stdout)) + "/.git/commit_template"

	file, err := os.Open(templateFile)
	if err != nil {
		return defaultCommitTemplate, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// We are only interested in the first line
	scanner.Scan()
	template := scanner.Text()
	if template == "" {
		return defaultCommitTemplate, nil
	}

	return template, nil
}

func writeMessageToFile(f *os.File, message string) error {
	if _, err := fmt.Fprintln(f, message); err != nil {
		return err
	}
	return nil
}

func promptInt(prompt string) (string, error) {
	prompt = strings.TrimPrefix(prompt, ";")
	for {
		message := readString(prompt + ": ")
		if _, err := strconv.Atoi(message); err == nil {
			return message, nil
		}
	}
}

func promptString(prompt string) (string, error) {
	prompt = strings.TrimPrefix(prompt, "; ")
	return promptMessage(prompt + ":")
}

func promptCommitType() (string, error) {
	possibleTypes := []string{
		"code", "fix", "chore", "refactor", "test", "build", "doc", "tool",
		"remove", "infra", "hint",
	}
	fmt.Println("Select the commit type:")
	for idx, t := range possibleTypes {
		fmt.Printf("\033[38;5;87m%3d\033[0m - %s\n", idx+1, t)
	}
	return readChoice(possibleTypes), nil
}

func promptRelevantPackages(separator string) (string, error) {
	app := "git diff --staged --name-only | xargs dirname | sort -u"
	cmd := exec.Command("bash", "-c", app)
	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error while reading modified files: %v", err)
	}

	modifiedPackages := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	if len(modifiedPackages) == 1 {
		return modifiedPackages[0], nil
	}

	relevantPackages := []string{}
	fmt.Println("Select the affected package(s):")
	for idx, p := range modifiedPackages {
		fmt.Printf("%4d - %s\n", idx+1, p)
	}
	relevantPackages = readMultipleChoices(modifiedPackages)
	return strings.Join(relevantPackages, separator), nil
}

func promptTicketNumber() (string, error) {
	app := "rev-parse --abbrev-ref HEAD"
	cmd := exec.Command("git", strings.Split(app, " ")...)
	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error while reading current branch name: %s", err)
	}
	re := regexp.MustCompile(`^[A-Z]+-\d+`)
	found := re.FindAll([]byte(strings.ToUpper(string(stdout))), -1)
	var ticketNumber string
	if len(found) == 1 {
		ticketNumber = string(found[0])
	}

	if ticketNumber == "" {
		ticketNumber = strings.ToUpper(readString("Jira ticket: "))
	} else {
		fmt.Printf("Jira ticket %s found, use it?\n", ticketNumber)
		for {
			res := strings.ToLower(readString("(yes/no)> "))
			if res == "y" || res == "yes" {
				break
			} else if res == "n" || res == "no" {
				ticketNumber = ""
				break
			}
		}
	}
	return ticketNumber, nil
}

func promptMessage(prompt string) (string, error) {
	for {
		message := readString(prompt)
		if message != "" {
			return message, nil
		}
	}
}

func readString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	res, _ := reader.ReadString('\n')
	return strings.TrimSpace(res)
}

func readChoice(choices []string) string {
	for {
		idx, err := strconv.Atoi(readString(fmt.Sprintf("(1-%d)> ", len(choices))))
		if err == nil && 1 <= idx && idx <= len(choices) {
			return choices[idx-1]
		}
	}
}

func readMultipleChoices(choices []string) []string {
	for {
		res := []string{}
		hasErrors := false
		userInput := readString(fmt.Sprintf("(1-%d)> ", len(choices)))
		for _, val := range strings.Split(userInput, " ") {
			if val == "" {
				continue
			}
			idx, err := strconv.Atoi(val)
			if err != nil {
				fmt.Printf("invalid value %s\n", val)
				hasErrors = true
				break
			} else if idx < 1 || len(choices) < idx {
				hasErrors = true
				break
			}
			res = append(res, choices[idx-1])
		}
		if !hasErrors {
			return res
		}
	}
}
