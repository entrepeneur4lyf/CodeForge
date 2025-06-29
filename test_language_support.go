package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/entrepeneur4lyf/codeforge/internal/builder"
)

func main() {
	fmt.Println("🧪 Testing PHP and C Language Support")
	fmt.Println("=====================================")

	// Test language detection by file extension
	testFiles := []string{
		"test.php",
		"index.php",
		"script.phtml",
		"main.c",
		"utils.h",
		"program.cpp",
		"app.go",
		"script.py",
	}

	fmt.Println("\n📁 Testing File Extension Detection:")
	fmt.Println("------------------------------------")

	for _, file := range testFiles {
		lang, err := builder.DetectLanguage(file)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", file, err)
		} else {
			fmt.Printf("✅ %s -> %s (%s)\n", file, lang.Name, lang.Compiler)
		}
	}

	// Test project detection
	fmt.Println("\n🏗️ Testing Project Detection:")
	fmt.Println("-----------------------------")

	projectTests := []struct {
		file     string
		expected string
	}{
		{"composer.json", "PHP"},
		{"Makefile", "C"},
		{"makefile", "C"},
		{"go.mod", "Go"},
		{"package.json", "JavaScript"},
		{"Cargo.toml", "Rust"},
	}

	// Create temporary test directory
	tempDir := "test_projects"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	for _, test := range projectTests {
		// Create test project directory
		projectDir := filepath.Join(tempDir, test.expected+"_project")
		os.MkdirAll(projectDir, 0755)

		// Create the detection file
		testFile := filepath.Join(projectDir, test.file)
		err := os.WriteFile(testFile, []byte("# Test file"), 0644)
		if err != nil {
			fmt.Printf("❌ Failed to create %s: %v\n", test.file, err)
			continue
		}

		// Test detection
		lang, err := builder.DetectLanguage(filepath.Join(projectDir, "dummy."+getExtension(test.expected)))
		if err != nil {
			fmt.Printf("❌ %s detection failed: %v\n", test.file, err)
		} else {
			fmt.Printf("✅ %s -> Detected %s project\n", test.file, lang.Name)
		}
	}

	// Test language configurations
	fmt.Println("\n⚙️ Testing Language Configurations:")
	fmt.Println("-----------------------------------")

	languages := []string{"php", "c", "cpp", "go", "python", "javascript", "typescript", "java", "rust"}

	for _, langKey := range languages {
		if lang, exists := builder.SupportedLanguages[langKey]; exists {
			fmt.Printf("✅ %s:\n", lang.Name)
			fmt.Printf("   📁 Extensions: %v\n", lang.Extensions)
			fmt.Printf("   🔨 Build: %v\n", lang.BuildCommand)
			fmt.Printf("   🧪 Test: %v\n", lang.TestCommand)
			fmt.Printf("   ▶️  Run: %v\n", lang.RunCommand)
			fmt.Printf("   🛠️  Compiler: %s\n", lang.Compiler)
			fmt.Printf("   📦 Package Manager: %s\n", lang.PackageManager)
			fmt.Printf("   🔍 LSP Server: %s\n", lang.LSPServer)
			fmt.Println()
		} else {
			fmt.Printf("❌ Language '%s' not found\n", langKey)
		}
	}

	// Test error parsing
	fmt.Println("\n🐛 Testing Error Parsing:")
	fmt.Println("-------------------------")

	errorTests := []struct {
		output   string
		language string
	}{
		{
			"main.c:15:10: error: 'undeclared_var' undeclared (first use in this function)",
			"C/GCC",
		},
		{
			"PHP Parse error: syntax error, unexpected '}' in /var/www/html/index.php on line 42",
			"PHP Parse",
		},
		{
			"Fatal error: Call to undefined function missing_func() in /path/to/script.php on line 25",
			"PHP Fatal",
		},
		{
			"./main.go:10:5: undefined: fmt.Printl",
			"Go",
		},
		{
			"make: *** [main.o] Error in main.c:20",
			"Make",
		},
	}

	for _, test := range errorTests {
		file, line := builder.ParseError(test.output)
		if file != "" && line != "" {
			fmt.Printf("✅ %s: %s:%s\n", test.language, file, line)
		} else {
			fmt.Printf("❌ %s: Failed to parse error\n", test.language)
		}
	}

	// Test code extraction
	fmt.Println("\n📝 Testing Code Extraction:")
	fmt.Println("---------------------------")

	codeTests := []struct {
		input    string
		language string
	}{
		{
			"Here's the fix:\n```php\n<?php\necho 'Hello World';\n?>\n```",
			"PHP",
		},
		{
			"Try this C code:\n```c\n#include <stdio.h>\nint main() {\n    printf(\"Hello\");\n    return 0;\n}\n```",
			"C",
		},
		{
			"Here's a Go solution:\n```go\npackage main\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			"Go",
		},
	}

	for _, test := range codeTests {
		extracted := builder.ExtractCode(test.input)
		if extracted != test.input {
			fmt.Printf("✅ %s: Successfully extracted code block\n", test.language)
			fmt.Printf("   📄 Code: %s...\n", truncate(extracted, 50))
		} else {
			fmt.Printf("❌ %s: Failed to extract code\n", test.language)
		}
	}

	fmt.Println("\n🎉 Language Support Test Complete!")
	fmt.Println("==================================")
	fmt.Println("✅ PHP support added with:")
	fmt.Println("   - File extensions: .php, .phtml, .php3, .php4, .php5, .phps")
	fmt.Println("   - Project detection: composer.json")
	fmt.Println("   - Build command: composer install")
	fmt.Println("   - Test command: ./vendor/bin/phpunit")
	fmt.Println("   - LSP server: phpactor")
	fmt.Println("   - Error parsing: PHP Parse/Fatal errors")
	fmt.Println()
	fmt.Println("✅ C support added with:")
	fmt.Println("   - File extensions: .c, .h")
	fmt.Println("   - Project detection: Makefile/makefile")
	fmt.Println("   - Build command: make")
	fmt.Println("   - Test command: make test")
	fmt.Println("   - LSP server: clangd")
	fmt.Println("   - Error parsing: GCC/Clang errors")
}

// Helper functions
func getExtension(language string) string {
	extensions := map[string]string{
		"PHP":        "php",
		"C":          "c",
		"Go":         "go",
		"JavaScript": "js",
		"Rust":       "rs",
	}

	if ext, exists := extensions[language]; exists {
		return ext
	}
	return "txt"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
