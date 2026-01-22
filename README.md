# gitsink

Sync files to a Git repository while automatically maintaining repository health.

## Overview

gitsink is a tool that makes long-term Git operations with machine commits (automated commits) sustainable.

When using Git libraries like go-git or libgit2, automatic garbage collection does not run. With machine commits running frequently over long periods, loose objects accumulate continuously, causing inode exhaustion and performance degradation.

gitsink solves this problem by syncing files to a Git repository and running preventive GC alongside commits.

## Requirements

`git` command

## Installation

```bash
$ go install github.com/zinrai/gitsink:latest
```

## Usage

### Basic usage (remote sync)

```bash
$ gitsink -src /path/to/source -repo /path/to/repo -ssh-key ~/.ssh/id_ed25519
```

Workflow:

1. `git pull` to update the repository
2. Sync files from source directory to repository
3. Commit changed files individually
4. Run `git gc` if loose objects exceed threshold
5. `git push` to remote

### Local only

```bash
$ gitsink -src /path/to/source -repo /path/to/repo -local
```

Runs only local commits and GC without remote sync (pull/push).

### dry-run

```bash
$ gitsink -src /path/to/source -repo /path/to/repo -ssh-key ~/.ssh/id_ed25519 -dry-run
```

Shows what would be executed without making actual changes.

See [example](example/) to try it locally.

## Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| -src | Yes | - | Source directory |
| -repo | Yes | - | Git repository path |
| -local | - | false | Local only mode (no pull/push) |
| -ssh-key | * | - | SSH private key path (required unless -local) |
| -remote | - | origin | Remote name |
| -branch | - | main | Branch name |
| -gc-threshold | - | 10000 | Loose objects threshold for GC (0 to disable) |
| -author-name | - | gitsink | Author name for commits |
| -author-email | - | gitsink@localhost | Author email for commits |
| -dry-run | - | false | Show plan without executing |

## Output example

```
$ gitsink -src ./configs -repo ./repo -ssh-key ~/.ssh/id_ed25519
git -C ./repo pull --ff-only origin main
sync: scanning source directory ./configs
sync: 2 files added, 1 files updated
git -C ./repo add dc-tokyo/spine-01
git -C ./repo commit --author "gitsink <gitsink@localhost>" -m "Add spine-01 in dc-tokyo"
git -C ./repo add dc-tokyo/leaf-01
git -C ./repo commit --author "gitsink <gitsink@localhost>" -m "Add leaf-01 in dc-tokyo"
git -C ./repo add dc-osaka/spine-02
git -C ./repo commit --author "gitsink <gitsink@localhost>" -m "Update spine-02 in dc-osaka"
gc: checking loose objects count
gc: 15234 loose objects (threshold: 10000)
gc: 15234 loose objects exceeds threshold 10000
git -C ./repo gc --prune=now
gc: completed (15234 -> 3 objects)
git -C ./repo push origin main
```

## Design notes

### File deletion

gitsink does not delete files from the repository that are missing from the source directory.

There are various reasons why a file might be missing from the source:

- The device was decommissioned
- Network failure during retrieval
- Authentication error during retrieval
- Bug in upstream tool

Since gitsink cannot distinguish between these cases, it does not delete files for safety. If deletion is needed, run `git rm` directly in the repository.

### Push failure behavior

If someone else pushes while gitsink is running, the push will fail. In this case, gitsink automatically runs `git reset --hard origin/{branch}` to discard local commits and exits with an error.

On the next run, pull fetches the latest state and changes are re-detected and committed, allowing automatic recovery.

## License

This project is licensed under the [MIT License](./LICENSE).
