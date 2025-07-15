#!/bin/bash

set -e

GO_VERSION="1.23.11"
GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://dl.google.com/go/${GO_TAR}"

echo "➡️  Downloading Go ${GO_VERSION}..."
wget -q --show-progress "${GO_URL}"

echo "🚮 Removing old Go installations..."
sudo rm -rf /usr/local/go

echo "📦 Extracting Go ${GO_VERSION}..."
sudo tar -C /usr/local -xzf "${GO_TAR}"

echo "🧹 Cleaning up..."
rm -f "${GO_TAR}"

echo "🔧 Setting Go in PATH for this session..."
export PATH=/usr/local/go/bin:$PATH

echo "✅ Installed Go version:"
go version

echo "🎉 Installation complete!"
