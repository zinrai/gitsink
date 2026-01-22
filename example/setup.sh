#!/bin/bash
set -e

cd "$(dirname "$0")"

# Clean up
rm -rf src repo

# Create source directory with sample files
mkdir -p src/dc-tokyo src/dc-osaka

cat > src/dc-tokyo/spine-01 << 'EOF'
hostname spine-01
!
interface Ethernet1
  description uplink
  ip address 10.0.1.1/30
!
EOF

cat > src/dc-tokyo/leaf-01 << 'EOF'
hostname leaf-01
!
interface Ethernet1
  description to-spine-01
  ip address 10.0.1.2/30
!
EOF

cat > src/dc-osaka/spine-01 << 'EOF'
hostname spine-01
!
interface Ethernet1
  description uplink
  ip address 10.1.1.1/30
!
EOF

# Create git repository
mkdir repo
cd repo
git init
git config user.email "test@example.com"
git config user.name "Test User"
cd ..

echo ""
echo "Setup complete."
echo ""
echo "Run gitsink:"
echo "  gitsink -src ./example/src -repo ./example/repo -local"
echo ""
echo "Or from example directory:"
echo "  cd example"
echo "  ../gitsink -src ./src -repo ./repo -local"
