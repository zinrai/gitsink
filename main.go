package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	srcDir      string
	repoDir     string
	local       bool
	remote      string
	branch      string
	sshKey      string
	gcThreshold int
	authorName  string
	authorEmail string
	dryRun      bool
}

func main() {
	cfg := parseFlags()

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.srcDir, "src", "", "Source directory (required)")
	flag.StringVar(&cfg.repoDir, "repo", "", "Git repository path (required)")
	flag.BoolVar(&cfg.local, "local", false, "Local only mode (no pull/push)")
	flag.StringVar(&cfg.remote, "remote", "origin", "Remote name")
	flag.StringVar(&cfg.branch, "branch", "main", "Branch name")
	flag.StringVar(&cfg.sshKey, "ssh-key", "", "SSH private key path (required unless -local)")
	flag.IntVar(&cfg.gcThreshold, "gc-threshold", 10000, "Loose objects threshold for GC (0 to disable)")
	flag.StringVar(&cfg.authorName, "author-name", "gitsink", "Author name for commits")
	flag.StringVar(&cfg.authorEmail, "author-email", "gitsink@localhost", "Author email for commits")
	flag.BoolVar(&cfg.dryRun, "dry-run", false, "Show plan without executing")

	flag.Parse()

	if cfg.srcDir == "" || cfg.repoDir == "" {
		fmt.Fprintln(os.Stderr, "error: -src and -repo are required")
		flag.Usage()
		os.Exit(1)
	}

	if !cfg.local && cfg.sshKey == "" {
		fmt.Fprintln(os.Stderr, "error: -ssh-key is required (use -local for local only mode)")
		flag.Usage()
		os.Exit(1)
	}

	return cfg
}

func run(cfg *config) error {
	if err := checkGitCommand(); err != nil {
		return err
	}

	if _, err := os.Stat(cfg.srcDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", cfg.srcDir)
	}

	if err := validateRepository(cfg.repoDir); err != nil {
		return err
	}

	// Pull (remote mode only)
	if !cfg.local {
		if err := pullChanges(cfg.repoDir, cfg); err != nil {
			return fmt.Errorf("failed to pull: %w", err)
		}
	}

	// Sync
	fmt.Printf("sync: scanning source directory %s\n", cfg.srcDir)

	changes, err := syncFiles(cfg.srcDir, cfg.repoDir, cfg.dryRun)
	if err != nil {
		return fmt.Errorf("failed to sync files: %w", err)
	}

	fmt.Printf("sync: %d files added, %d files updated\n",
		changes.added, changes.updated)

	// Commit
	if err := commitChanges(cfg.repoDir, changes, cfg); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// GC
	if cfg.gcThreshold > 0 {
		if err := runGCIfNeeded(cfg.repoDir, cfg.gcThreshold, cfg.dryRun); err != nil {
			return fmt.Errorf("failed to run gc: %w", err)
		}
	}

	// Push (remote mode only)
	if !cfg.local {
		if err := pushChanges(cfg.repoDir, cfg); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
	}

	return nil
}
