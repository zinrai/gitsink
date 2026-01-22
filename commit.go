package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func formatGitArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		if strings.Contains(arg, " ") {
			quoted[i] = fmt.Sprintf("%q", arg)
		} else {
			quoted[i] = arg
		}
	}
	return strings.Join(quoted, " ")
}

func logGitCommand(args []string) {
	fmt.Printf("git %s\n", formatGitArgs(args))
}

func validateRepository(repoDir string) error {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository: %s", repoDir)
	}
	return nil
}

func gitSSHCommand(sshKey string) string {
	return fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new", sshKey)
}

func pullChanges(repoDir string, cfg *config) error {
	args := []string{"-C", repoDir, "pull", "--ff-only", cfg.remote, cfg.branch}

	if cfg.dryRun {
		fmt.Printf("[dry-run] git %s\n", formatGitArgs(args))
		return nil
	}

	logGitCommand(args)
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+gitSSHCommand(cfg.sshKey))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}

func commitChanges(repoDir string, changes *changeSet, cfg *config) error {
	if len(changes.changes) == 0 {
		fmt.Println("commit: no changes to commit")
		return nil
	}

	for _, change := range changes.changes {
		msg := formatCommitMessage(change)

		// Stage the file
		addArgs := []string{"-C", repoDir, "add", change.path}

		if cfg.dryRun {
			fmt.Printf("[dry-run] git %s\n", formatGitArgs(addArgs))
			commitArgs := []string{"-C", repoDir, "commit", "--author", fmt.Sprintf("%s <%s>", cfg.authorName, cfg.authorEmail), "-m", msg}
			fmt.Printf("[dry-run] git %s\n", formatGitArgs(commitArgs))
			continue
		}

		logGitCommand(addArgs)
		cmd := exec.Command("git", addArgs...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage %s: %w", change.path, err)
		}

		// Check if there are staged changes
		checkCmd := exec.Command("git", "-C", repoDir, "diff", "--cached", "--quiet")
		if err := checkCmd.Run(); err == nil {
			fmt.Printf("commit: skipped %s (no changes)\n", change.path)
			continue
		}

		// Commit
		commitArgs := []string{"-C", repoDir, "commit", "--author", fmt.Sprintf("%s <%s>", cfg.authorName, cfg.authorEmail), "-m", msg}
		logGitCommand(commitArgs)
		cmd = exec.Command("git", commitArgs...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to commit %s: %w", change.path, err)
		}
	}

	return nil
}

func formatCommitMessage(change fileChange) string {
	dir := filepath.Dir(change.path)
	file := filepath.Base(change.path)

	var action string
	switch change.action {
	case "add":
		action = "Add"
	case "update":
		action = "Update"
	}

	if dir == "." {
		return fmt.Sprintf("%s %s", action, file)
	}
	return fmt.Sprintf("%s %s in %s", action, file, dir)
}

func pushChanges(repoDir string, cfg *config) error {
	args := []string{"-C", repoDir, "push", cfg.remote, cfg.branch}

	if cfg.dryRun {
		fmt.Printf("[dry-run] git %s\n", formatGitArgs(args))
		return nil
	}

	logGitCommand(args)
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+gitSSHCommand(cfg.sshKey))

	if err := cmd.Run(); err != nil {
		// Reset to remote state to allow clean retry on next run
		ref := fmt.Sprintf("%s/%s", cfg.remote, cfg.branch)
		resetArgs := []string{"-C", repoDir, "reset", "--hard", ref}
		logGitCommand(resetArgs)
		resetCmd := exec.Command("git", resetArgs...)
		if resetErr := resetCmd.Run(); resetErr != nil {
			return fmt.Errorf("push failed and reset failed: push=%w, reset=%v", err, resetErr)
		}
		return fmt.Errorf("push failed (remote has new commits, local commits discarded)")
	}

	return nil
}
