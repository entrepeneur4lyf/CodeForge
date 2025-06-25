package builder

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func BuildGo() ([]byte, error) {
	cmd := exec.Command("go", "build", "./...")
	return cmd.CombinedOutput()
}

func ApplyFix(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

func ParseError(output string) (string, string) {
	re := regexp.MustCompile(`(?m)^(# .*?)
(.*?):(\d+):(\d+): (.*)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) > 2 {
		filePath := strings.TrimSpace(matches[2])
		lineNumber := matches[3]
		return filePath, lineNumber
	}

	return "", ""
}

func ExtractCode(response string) string {
	re := regexp.MustCompile("(?s)```go\n(.*)```")
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		return matches[1]
	}
	return response
}

func GenerateDiff(filePath string, newContent string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(content), newContent, false)

	return dmp.DiffPrettyText(diffs), nil
}
