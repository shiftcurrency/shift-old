#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
shfdir="$workspace/src/github.com/shiftcurrency"
if [ ! -L "$shfdir/shift" ]; then
    mkdir -p "$shfdir"
    cd "$shfdir"
    ln -s ../../../../../. shift
    cd "$root"
fi

# Set up the environment to use the workspace.
# Also add Godeps workspace so we build using canned dependencies.
GOPATH="$shfdir/shift/Godeps/_workspace:$workspace"
GOBIN="$PWD/build/bin"
export GOPATH GOBIN

# Run the command inside the workspace.
cd "$shfdir/shift"
PWD="$shfdir/shift"

# Launch the arguments with the configured environment.
exec "$@"
