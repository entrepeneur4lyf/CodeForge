package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Language represents a supported programming language
type Language struct {
	Name           string
	Extensions     []string
	BuildCommand   []string
	TestCommand    []string
	RunCommand     []string
	Compiler       string
	PackageManager string
	LSPServer      string
}

// SupportedLanguages defines all languages from Phase 1 specification
var SupportedLanguages = map[string]Language{
	"go": {
		Name:           "Go",
		Extensions:     []string{".go"},
		BuildCommand:   []string{"go", "build", "./..."},
		TestCommand:    []string{"go", "test", "./..."},
		RunCommand:     []string{"go", "run"},
		Compiler:       "go",
		PackageManager: "go mod",
		LSPServer:      "gopls",
	},
	"rust": {
		Name:           "Rust",
		Extensions:     []string{".rs"},
		BuildCommand:   []string{"cargo", "build"},
		TestCommand:    []string{"cargo", "test"},
		RunCommand:     []string{"cargo", "run"},
		Compiler:       "rustc",
		PackageManager: "cargo",
		LSPServer:      "rust-analyzer",
	},
	"python": {
		Name:           "Python",
		Extensions:     []string{".py"},
		BuildCommand:   []string{"python", "-m", "build"},
		TestCommand:    []string{"pytest"},
		RunCommand:     []string{"python"},
		Compiler:       "python3",
		PackageManager: "pip",
		LSPServer:      "pylsp",
	},
	"javascript": {
		Name:           "JavaScript",
		Extensions:     []string{".js", ".mjs"},
		BuildCommand:   []string{"npm", "run", "build"},
		TestCommand:    []string{"npm", "test"},
		RunCommand:     []string{"node"},
		Compiler:       "node",
		PackageManager: "npm",
		LSPServer:      "typescript-language-server",
	},
	"typescript": {
		Name:           "TypeScript",
		Extensions:     []string{".ts", ".tsx"},
		BuildCommand:   []string{"tsc"},
		TestCommand:    []string{"npm", "test"},
		RunCommand:     []string{"ts-node"},
		Compiler:       "tsc",
		PackageManager: "npm",
		LSPServer:      "typescript-language-server",
	},
	"java": {
		Name:           "Java",
		Extensions:     []string{".java"},
		BuildCommand:   []string{"mvn", "compile"},
		TestCommand:    []string{"mvn", "test"},
		RunCommand:     []string{"java"},
		Compiler:       "javac",
		PackageManager: "maven",
		LSPServer:      "jdtls",
	},
	"cpp": {
		Name:           "C++",
		Extensions:     []string{".cpp", ".cc", ".cxx", ".hpp", ".h"},
		BuildCommand:   []string{"cmake", "--build", "build"},
		TestCommand:    []string{"ctest"},
		RunCommand:     []string{"./main"},
		Compiler:       "g++",
		PackageManager: "vcpkg",
		LSPServer:      "clangd",
	},
}

// DetectLanguage determines the language based on file extension
func DetectLanguage(filePath string) (Language, error) {
	ext := filepath.Ext(filePath)

	for _, lang := range SupportedLanguages {
		for _, langExt := range lang.Extensions {
			if ext == langExt {
				return lang, nil
			}
		}
	}

	return Language{}, fmt.Errorf("unsupported file extension: %s", ext)
}

// detectProjectLanguage detects the primary language of a project
func detectProjectLanguage(projectPath string) (Language, error) {
	// Check for language-specific files in order of priority
	languageFiles := map[string]string{
		"go.mod":           "go",
		"Cargo.toml":       "rust",
		"package.json":     "javascript",
		"tsconfig.json":    "typescript",
		"pom.xml":          "java",
		"CMakeLists.txt":   "cpp",
		"requirements.txt": "python",
		"setup.py":         "python",
		"pyproject.toml":   "python",
	}

	for file, langKey := range languageFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			if lang, exists := SupportedLanguages[langKey]; exists {
				return lang, nil
			}
		}
	}

	return Language{}, fmt.Errorf("could not detect project language in %s", projectPath)
}

// Build executes the build command for the detected language
func Build(projectPath string) ([]byte, error) {
	// Try to detect language from project structure
	lang, err := detectProjectLanguage(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project language: %w", err)
	}

	return BuildWithLanguage(projectPath, lang)
}

// BuildWithLanguage builds a project with a specific language
func BuildWithLanguage(projectPath string, lang Language) ([]byte, error) {
	cmd := exec.Command(lang.BuildCommand[0], lang.BuildCommand[1:]...)
	cmd.Dir = projectPath
	return cmd.CombinedOutput()
}

// BuildGo maintains backward compatibility
func BuildGo() ([]byte, error) {
	return Build(".")
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
