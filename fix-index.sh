#!/bin/bash

echo "Fixing index issue..."

# Clear local cache
echo "Clearing local cache..."
rm -rf ~/.cache/prompt-vault/prompts/*

# Run sync to rebuild from GitHub
echo "Syncing from GitHub..."
./pv sync -v

echo "Done! Your index should now be clean."
echo ""
echo "Next steps:"
echo "1. Check 'pv list' to see if the empty entry is gone"
echo "2. Try uploading your test prompts again"