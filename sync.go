package main

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type fileChange struct {
	path   string
	action string // "add", "update"
}

type changeSet struct {
	changes []fileChange
	added   int
	updated int
}

func syncFiles(srcDir, repoDir string, dryRun bool) (*changeSet, error) {
	srcFiles, err := listFiles(srcDir)
	if err != nil {
		return nil, err
	}

	repoFiles, err := listFiles(repoDir)
	if err != nil {
		return nil, err
	}

	changes := &changeSet{}

	for relPath := range srcFiles {
		change, err := processFile(srcDir, repoDir, relPath, repoFiles, dryRun)
		if err != nil {
			return nil, err
		}
		if change == nil {
			continue
		}
		changes.changes = append(changes.changes, *change)
		if change.action == "add" {
			changes.added++
		} else {
			changes.updated++
		}
	}

	return changes, nil
}

func processFile(srcDir, repoDir, relPath string, repoFiles map[string]struct{}, dryRun bool) (*fileChange, error) {
	srcPath := filepath.Join(srcDir, relPath)
	repoPath := filepath.Join(repoDir, relPath)

	_, exists := repoFiles[relPath]

	if !exists {
		if !dryRun {
			if err := copyFile(srcPath, repoPath); err != nil {
				return nil, err
			}
		}
		return &fileChange{path: relPath, action: "add"}, nil
	}

	differs, err := filesDiffer(srcPath, repoPath)
	if err != nil {
		return nil, err
	}

	if !differs {
		return nil, nil
	}

	if !dryRun {
		if err := copyFile(srcPath, repoPath); err != nil {
			return nil, err
		}
	}
	return &fileChange{path: relPath, action: "update"}, nil
}

func listFiles(root string) (map[string]struct{}, error) {
	files := make(map[string]struct{})

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		if !d.IsDir() {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files[relPath] = struct{}{}
		}

		return nil
	})

	return files, err
}

func filesDiffer(path1, path2 string) (bool, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return false, err
	}

	content2, err := os.ReadFile(path2)
	if err != nil {
		return false, err
	}

	return !bytes.Equal(content1, content2), nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
