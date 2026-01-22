# Example

Sample environment to try gitsink locally.

## Usage

```bash
# Build gitsink
go build -o gitsink .

# Set up sample environment
./example/setup.sh

# Run gitsink
./gitsink -src ./example/src -repo ./example/repo -local
```

## What setup.sh creates

```
example/
├── src/                  # Source directory
│   ├── dc-tokyo/
│   │   ├── spine-01
│   │   └── leaf-01
│   └── dc-osaka/
│       └── spine-01
└── repo/                 # Git repository (empty)
```

## Try it

```bash
# First run (adds 3 files)
./gitsink -src ./example/src -repo ./example/repo -local

# Check commit history
git -C ./example/repo log --oneline

# Update a file
echo "update" >> ./example/src/dc-tokyo/spine-01

# Run again (updates 1 file)
./gitsink -src ./example/src -repo ./example/repo -local
```
