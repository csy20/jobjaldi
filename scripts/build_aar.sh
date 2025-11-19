#!/bin/bash
set -e

# Ensure we are in the project root
cd "$(dirname "$0")/.."

export PATH=$PATH:/opt/android-studio/jbr/bin

echo "Building jobagent.aar..."

# Go to scrapers directory
cd scrapers

# Build the AAR with the correct package name
# The package name must match what is used in MainActivity.kt: dev.csy.jobagent.jobagent
gomobile bind -target=android -androidapi=21 -o jobagent.aar -javapkg=dev.csy.jobagent ./jobagent

echo "Copying AAR to Android project..."
mkdir -p ../android/app/libs
cp jobagent.aar ../android/app/libs/

echo "Done! AAR updated."
