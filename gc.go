package main

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
)

func checkGitCommand() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git command not found: gitsink requires git")
	}
	return nil
}

func countLooseObjects(repoDir string) (int, error) {
	objectsDir := filepath.Join(repoDir, ".git", "objects")
	count := 0

	err := filepath.WalkDir(objectsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == objectsDir {
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if name == "pack" || name == "info" {
				return filepath.SkipDir
			}
			return nil
		}

		dir := filepath.Base(filepath.Dir(path))
		if len(dir) == 2 {
			count++
		}

		return nil
	})

	return count, err
}

func runGC(repoDir string) error {
	args := []string{"-C", repoDir, "gc", "--prune=now"}
	logGitCommand(args)
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func runGCIfNeeded(repoDir string, threshold int, dryRun bool) error {
	fmt.Println("gc: checking loose objects count")

	count, err := countLooseObjects(repoDir)
	if err != nil {
		return err
	}

	if count <= threshold {
		fmt.Printf("gc: %d loose objects (threshold: %d)\n", count, threshold)
		fmt.Println("gc: skipped")
		return nil
	}

	args := []string{"-C", repoDir, "gc", "--prune=now"}

	if dryRun {
		fmt.Printf("[dry-run] git %s\n", formatGitArgs(args))
		return nil
	}

	fmt.Printf("gc: %d loose objects exceeds threshold %d\n", count, threshold)

	if err := runGC(repoDir); err != nil {
		return err
	}

	newCount, err := countLooseObjects(repoDir)
	if err != nil {
		return err
	}

	fmt.Printf("gc: completed (%d -> %d objects)\n", count, newCount)
	return nil
}
