package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stolostron/acm-mcp-go/config"
)

func main() {
	repoPath := flag.String("path", "", "repo path)")
	repoType := flag.String("type", "", "which kind of repo you want to use)")
	branch := flag.String("branch", "", "which branch of repo you want to use)")

	flag.Parse()
	cfg := config.NewConfig(*repoPath, *repoType)
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}
	// Create a new MCP server
	s := server.NewMCPServer(
		"Learn acm knownledge tool",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	repoRootPath := cfg.RepoPath
	if cfg.RepoType == "git" || cfg.RepoType == "github" {
		repoRootPath = cloneRepo(cfg.RepoPath, *branch)
	}

	// Add a learn dir tool
	learnDir := mcp.NewTool("use-acm",
		mcp.WithDescription("learn the acm knownledge, and use the acm env to do the things"),
		mcp.WithString("env",
			mcp.Description("The env directory you want to load"),
		),
	)

	// Add the learn dir handler
	s.AddTool(learnDir, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Using helper functions for type-safe argument access
		dir, err := request.RequireString("dir")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		targetPath := repoRootPath + "/" + dir

		targetDirContexts := readAllFileInDir(targetPath)

		var result string
		systemPrompt := readFile(repoRootPath + "/system_prompt.md")
		if len(systemPrompt) > 10 {
			result = systemPrompt + " "
		}
		formatedTargetDirContexts := formatMapOutPut(targetDirContexts)
		if len(formatedTargetDirContexts) > 10 {
			result = result + "\n" + "You should learn the following knownledges context which get from files." + "\n" + formatedTargetDirContexts
		}
		return mcp.NewToolResultText(result + "\n" + formatedTargetDirContexts), nil
	})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// clone the github repo from repopath
func cloneRepo(repoPath string, branch string) string {
	gitLocalPath := "tmp/"

	// Clean up existing directory
	os.RemoveAll(gitLocalPath)

	// Create the target directory
	if err := os.MkdirAll(gitLocalPath, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", gitLocalPath, err)
		return ""
	}
	var cmd *exec.Cmd
	if len(branch) == 0 {
		cmd = exec.Command("git", "clone", repoPath, gitLocalPath)
	} else {
		cmd = exec.Command("git", "clone", "-b", branch, repoPath, gitLocalPath)
	}
	// Execute git clone command
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error cloning repository %s: %v\n", repoPath, err)
		return ""
	}

	return gitLocalPath
}

// format the map output
func formatMapOutPut(contexts map[string]string) string {
	var result string
	for path, content := range contexts {
		result = result + "File: " + path + "\n" + "Context: \n" + content + "\n" + "\n-------------------------------------------\n"
	}
	return result
}

// read all file in dir
func readAllFileInDir(path string) map[string]string {
	files := listDir(path)
	contexts := make(map[string]string)
	for _, file := range files {
		if strings.HasSuffix(file, ".md") {
			allPath := path + "/" + file
			content := readFile(allPath)
			contexts[allPath] = content
		}
	}
	return contexts
}

func readFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func listDir(path string) []string {
	var paths []string
	err := filepath.Walk(path, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", fullPath, err)
			return nil // 继续遍历其他文件
		}

		// 将绝对路径转换为相对于根路径的相对路径
		relPath, err := filepath.Rel(path, fullPath)
		if err != nil {
			relPath = fullPath
		}

		// 跳过根目录本身
		if relPath != "." {
			paths = append(paths, relPath)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory %s: %v\n", path, err)
		return []string{}
	}

	return paths
}
