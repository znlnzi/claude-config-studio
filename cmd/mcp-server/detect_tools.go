package main

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── Types ──────────────────────────────────────────────────

type languageInfo struct {
	Name      string `json:"name"`
	Indicator string `json:"indicator"`
}

type frameworkInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version,omitempty"`
	Indicator string `json:"indicator"`
}

type testFrameworkInfo struct {
	Name      string `json:"name"`
	Indicator string `json:"indicator"`
}

type packageManagerInfo struct {
	Name      string `json:"name"`
	Indicator string `json:"indicator"`
}

type gitInfo struct {
	Initialized   bool     `json:"initialized"`
	RecentAuthors []string `json:"recent_authors"`
	IsSolo        bool     `json:"is_solo"`
	Branch        string   `json:"branch"`
}

type claudeConfigInfo struct {
	HasClaudeDir bool            `json:"has_claude_dir"`
	HasSetupMeta bool            `json:"has_setup_meta"`
	SetupMeta    json.RawMessage `json:"setup_meta"`
}

type projectDetection struct {
	ProjectPath    string              `json:"project_path"`
	Name           string              `json:"name"`
	Languages      []languageInfo      `json:"languages"`
	Frameworks     []frameworkInfo      `json:"frameworks"`
	TestFrameworks []testFrameworkInfo  `json:"test_frameworks"`
	PackageManager *packageManagerInfo  `json:"package_manager"`
	Git            *gitInfo            `json:"git"`
	ClaudeConfig   *claudeConfigInfo   `json:"claude_config"`
}

// ─── Tool Definition ────────────────────────────────────────

func buildDetectProjectTool() mcp.Tool {
	return mcp.NewTool(
		"detect_project",
		mcp.WithDescription("Detect project characteristics for setup wizard. Scans for language, framework, test framework, package manager, and git status."),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Absolute project path to scan"),
		),
	)
}

// ─── Handler ────────────────────────────────────────────────

func handleDetectProject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError("project_path is required"), nil
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(errPathNotFound(projectPath)), nil
	}

	result := projectDetection{
		ProjectPath:    projectPath,
		Name:           filepath.Base(projectPath),
		Languages:      detectLanguages(projectPath),
		Frameworks:     detectFrameworks(projectPath),
		TestFrameworks: detectTestFrameworks(projectPath),
		PackageManager: detectPackageManager(projectPath),
		Git:            detectGitInfo(projectPath),
		ClaudeConfig:   detectClaudeConfig(projectPath),
	}

	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// ─── Language Detection ─────────────────────────────────────

var languageIndicators = []struct {
	file     string
	language string
	glob     bool // if true, use glob pattern matching
}{
	{"tsconfig.json", "TypeScript", false},
	{"package.json", "JavaScript", false},
	{"go.mod", "Go", false},
	{"pyproject.toml", "Python", false},
	{"setup.py", "Python", false},
	{"requirements.txt", "Python", false},
	{"Cargo.toml", "Rust", false},
	{"pom.xml", "Java", false},
	{"build.gradle", "Java", false},
	{"Gemfile", "Ruby", false},
	{"composer.json", "PHP", false},
	{"Package.swift", "Swift", false},
	{"pubspec.yaml", "Dart/Flutter", false},
	{"*.csproj", "C#", true},
	{"*.sln", "C#", true},
}

func detectLanguages(projectPath string) []languageInfo {
	var langs []languageInfo
	seen := map[string]bool{}

	for _, ind := range languageIndicators {
		if seen[ind.language] {
			continue
		}

		if ind.glob {
			matches, _ := filepath.Glob(filepath.Join(projectPath, ind.file))
			if len(matches) > 0 {
				seen[ind.language] = true
				langs = append(langs, languageInfo{
					Name:      ind.language,
					Indicator: ind.file,
				})
			}
		} else {
			if fileExists(filepath.Join(projectPath, ind.file)) {
				seen[ind.language] = true
				langs = append(langs, languageInfo{
					Name:      ind.language,
					Indicator: ind.file,
				})
			}
		}
	}

	return langs
}

// ─── Framework Detection ────────────────────────────────────

var jsFrameworks = []struct {
	dep       string
	framework string
}{
	{"next", "Next.js"},
	{"react", "React"},
	{"vue", "Vue.js"},
	{"@angular/core", "Angular"},
	{"svelte", "Svelte"},
	{"express", "Express.js"},
	{"nuxt", "Nuxt.js"},
	{"gatsby", "Gatsby"},
}

var goFrameworks = []struct {
	dep       string
	framework string
}{
	{"github.com/gin-gonic/gin", "Gin"},
	{"github.com/labstack/echo", "Echo"},
	{"github.com/gofiber/fiber", "Fiber"},
}

var pythonFrameworks = []struct {
	dep       string
	framework string
}{
	{"fastapi", "FastAPI"},
	{"django", "Django"},
	{"flask", "Flask"},
}

func detectFrameworks(projectPath string) []frameworkInfo {
	var frameworks []frameworkInfo

	// JS/TS frameworks from package.json
	pkgPath := filepath.Join(projectPath, "package.json")
	if fileExists(pkgPath) {
		deps, devDeps := parsePackageJSONDeps(pkgPath)
		allDeps := mergeMaps(deps, devDeps)
		for _, fw := range jsFrameworks {
			if ver, ok := allDeps[fw.dep]; ok {
				frameworks = append(frameworks, frameworkInfo{
					Name:      fw.framework,
					Version:   ver,
					Indicator: fw.dep + " in package.json",
				})
			}
		}
	}

	// Go frameworks from go.mod
	goModPath := filepath.Join(projectPath, "go.mod")
	if fileExists(goModPath) {
		goModDeps := parseGoModDeps(goModPath)
		for _, fw := range goFrameworks {
			for _, dep := range goModDeps {
				if strings.HasPrefix(dep, fw.dep) {
					frameworks = append(frameworks, frameworkInfo{
						Name:      fw.framework,
						Indicator: fw.dep + " in go.mod",
					})
					break
				}
			}
		}
	}

	// Python frameworks from pyproject.toml
	pyprojectPath := filepath.Join(projectPath, "pyproject.toml")
	if fileExists(pyprojectPath) {
		pyDeps := parsePyprojectDeps(pyprojectPath)
		for _, fw := range pythonFrameworks {
			for _, dep := range pyDeps {
				if strings.EqualFold(dep, fw.dep) {
					frameworks = append(frameworks, frameworkInfo{
						Name:      fw.framework,
						Indicator: fw.dep + " in pyproject.toml",
					})
					break
				}
			}
		}
	}

	return frameworks
}

// ─── Test Framework Detection ───────────────────────────────

var jsTestFrameworks = []struct {
	dep  string
	name string
}{
	{"jest", "Jest"},
	{"vitest", "Vitest"},
	{"mocha", "Mocha"},
	{"playwright", "Playwright"},
	{"@playwright/test", "Playwright"},
	{"cypress", "Cypress"},
}

func detectTestFrameworks(projectPath string) []testFrameworkInfo {
	var testFws []testFrameworkInfo
	seen := map[string]bool{}

	// JS/TS test frameworks from package.json devDependencies
	pkgPath := filepath.Join(projectPath, "package.json")
	if fileExists(pkgPath) {
		_, devDeps := parsePackageJSONDeps(pkgPath)
		deps, _ := parsePackageJSONDeps(pkgPath)
		allDeps := mergeMaps(deps, devDeps)
		for _, fw := range jsTestFrameworks {
			if seen[fw.name] {
				continue
			}
			if _, ok := allDeps[fw.dep]; ok {
				seen[fw.name] = true
				testFws = append(testFws, testFrameworkInfo{
					Name:      fw.name,
					Indicator: fw.dep + " in package.json",
				})
			}
		}
	}

	// Python: pytest
	pyprojectPath := filepath.Join(projectPath, "pyproject.toml")
	if fileExists(pyprojectPath) {
		pyDeps := parsePyprojectDeps(pyprojectPath)
		for _, dep := range pyDeps {
			if strings.EqualFold(dep, "pytest") {
				testFws = append(testFws, testFrameworkInfo{
					Name:      "pytest",
					Indicator: "pytest in pyproject.toml",
				})
				break
			}
		}
	}

	// Go: built-in go test
	if fileExists(filepath.Join(projectPath, "go.mod")) {
		testFws = append(testFws, testFrameworkInfo{
			Name:      "go test",
			Indicator: "go.mod (built-in)",
		})
	}

	// Rust: built-in cargo test
	if fileExists(filepath.Join(projectPath, "Cargo.toml")) {
		testFws = append(testFws, testFrameworkInfo{
			Name:      "cargo test",
			Indicator: "Cargo.toml (built-in)",
		})
	}

	return testFws
}

// ─── Package Manager Detection ──────────────────────────────

var packageManagerIndicators = []struct {
	file    string
	manager string
}{
	{"pnpm-lock.yaml", "pnpm"},
	{"yarn.lock", "yarn"},
	{"bun.lockb", "bun"},
	{"bun.lock", "bun"},
	{"package-lock.json", "npm"},
	{"poetry.lock", "poetry"},
	{"uv.lock", "uv"},
	{"Pipfile.lock", "pipenv"},
	{"go.sum", "go modules"},
	{"Cargo.lock", "cargo"},
	{"Gemfile.lock", "bundler"},
}

func detectPackageManager(projectPath string) *packageManagerInfo {
	for _, ind := range packageManagerIndicators {
		if fileExists(filepath.Join(projectPath, ind.file)) {
			return &packageManagerInfo{
				Name:      ind.manager,
				Indicator: ind.file,
			}
		}
	}
	return nil
}

// ─── Git Detection ──────────────────────────────────────────

func detectGitInfo(projectPath string) *gitInfo {
	// Check if git is initialized
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return &gitInfo{Initialized: false}
	}

	info := &gitInfo{Initialized: true}

	// Get current branch
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectPath
	if out, err := cmd.Output(); err == nil {
		info.Branch = strings.TrimSpace(string(out))
	}

	// Get recent authors (unique)
	cmd = exec.Command("git", "log", "--format=%ae", "-20")
	cmd.Dir = projectPath
	if out, err := cmd.Output(); err == nil {
		authorSet := map[string]bool{}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				authorSet[line] = true
			}
		}
		for author := range authorSet {
			info.RecentAuthors = append(info.RecentAuthors, author)
		}
		info.IsSolo = len(info.RecentAuthors) <= 1
	}

	return info
}

// ─── Claude Config Detection ────────────────────────────────

func detectClaudeConfig(projectPath string) *claudeConfigInfo {
	claudeDir := filepath.Join(projectPath, ".claude")
	info := &claudeConfigInfo{
		HasClaudeDir: fileExists(claudeDir),
	}

	metaPath := filepath.Join(claudeDir, ".setup-meta.json")
	if fileExists(metaPath) {
		info.HasSetupMeta = true
		if data, err := os.ReadFile(metaPath); err == nil {
			info.SetupMeta = json.RawMessage(data)
		}
	}

	return info
}

// ─── Dependency Parsing Helpers ─────────────────────────────

func parsePackageJSONDeps(path string) (deps map[string]string, devDeps map[string]string) {
	deps = map[string]string{}
	devDeps = map[string]string{}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	if pkg.Dependencies != nil {
		deps = pkg.Dependencies
	}
	if pkg.DevDependencies != nil {
		devDeps = pkg.DevDependencies
	}
	return
}

func parseGoModDeps(path string) []string {
	var deps []string

	file, err := os.Open(path)
	if err != nil {
		return deps
	}
	defer file.Close()

	inRequire := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "require (") || strings.HasPrefix(line, "require(") {
			inRequire = true
			continue
		}
		if inRequire {
			if line == ")" {
				inRequire = false
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				deps = append(deps, parts[0])
			}
		}
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps = append(deps, parts[1])
			}
		}
	}

	return deps
}

func parsePyprojectDeps(path string) []string {
	var deps []string

	file, err := os.Open(path)
	if err != nil {
		return deps
	}
	defer file.Close()

	inDeps := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Match [project] dependencies or [tool.poetry.dependencies]
		if line == "dependencies = [" ||
			line == "[tool.poetry.dependencies]" ||
			line == "[project.dependencies]" {
			inDeps = true
			continue
		}

		if inDeps {
			if line == "]" || (strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]")) {
				inDeps = false
				continue
			}
			cleaned := strings.Trim(line, `",' `)
			if cleaned == "" || strings.HasPrefix(cleaned, "#") {
				continue
			}
			// Handle PEP 621 style: "fastapi>=0.100" — check version specifiers first
			extracted := false
			for _, sep := range []string{">=", "<=", "==", "!=", "~=", ">", "<", ";"} {
				if idx := strings.Index(cleaned, sep); idx > 0 {
					dep := strings.TrimSpace(cleaned[:idx])
					if dep != "" && dep != "python" {
						deps = append(deps, dep)
					}
					extracted = true
					break
				}
			}
			if extracted {
				continue
			}
			// Handle poetry style: name = "version" (only plain " = " without version specifiers)
			if strings.Contains(cleaned, " = ") {
				parts := strings.SplitN(cleaned, " = ", 2)
				dep := strings.TrimSpace(parts[0])
				if dep != "" && dep != "python" {
					deps = append(deps, dep)
				}
				continue
			}
			// Bare package name
			cleaned = strings.TrimSpace(cleaned)
			if cleaned != "" && cleaned != "python" {
				deps = append(deps, cleaned)
			}
		}
	}

	return deps
}

// ─── Utility ────────────────────────────────────────────────

func mergeMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}
