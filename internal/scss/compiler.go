package scss

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// compile scss to css using sass binary (managed by go module)
func Compile(inputPath, outputPath string) error {
	// try ~/.local/dart-sass/sass first (user installed)
	sassPath := filepath.Join(os.Getenv("HOME"), ".local/dart-sass/sass")

	// compile scss
	cmd := exec.Command(sassPath, "--no-source-map", inputPath, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("sass compilation failed: %v", err)
		log.Printf("output: %s", string(output))
		return err
	}

	log.Printf("compiled %s -> %s", inputPath, outputPath)
	return nil
}

// watch scss files for changes and recompile
func Watch(scssDir, cssDir string) {
	// initial compilation
	if err := CompileAll(scssDir, cssDir); err != nil {
		log.Printf("initial scss compilation failed: %v", err)
	}

	// track modification times
	modTimes := make(map[string]time.Time)

	// watch for changes
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		changed := false

		// check all scss files
		filepath.Walk(scssDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".scss" {
				modTime := info.ModTime()
				lastModTime, exists := modTimes[path]

				if !exists || modTime.After(lastModTime) {
					modTimes[path] = modTime
					if exists {
						changed = true
					}
				}
			}

			return nil
		})

		if changed {
			if err := CompileAll(scssDir, cssDir); err != nil {
				log.Printf("scss compilation failed: %v", err)
			}
		}
	}
}

// compile all scss files in directory
func CompileAll(scssDir, cssDir string) error {
	// ensure output directory exists
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return err
	}

	// compile main.scss
	inputPath := filepath.Join(scssDir, "main.scss")
	outputPath := filepath.Join(cssDir, "main.css")

	if _, err := os.Stat(inputPath); err == nil {
		if err := Compile(inputPath, outputPath); err != nil {
			return err
		}
	}

	return nil
}

// copy file helper (kept for backwards compatibility if needed)
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
