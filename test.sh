#!/bin/bash

echo "🧪 Running all tests for Poll Service..."
echo ""

# Run core backend tests
echo "📦 Running core backend tests..."
cd core
go test ./... -v -cover
if [ $? -ne 0 ]; then
    echo "❌ Core tests failed!"
    exit 1
fi
cd ..
echo "✅ Core tests passed!"
echo ""

# Run bot tests
echo "📦 Running Telegram bot tests..."
cd bot
go test ./... -v -cover
if [ $? -ne 0 ]; then
    echo "❌ Bot tests failed!"
    exit 1
fi
cd ..
echo "✅ Bot tests passed!"
echo ""

echo "🎉 All tests passed successfully!"
echo ""
echo "📊 Test Coverage Summary:"
echo "  - Core (Go): handler, service, config, middleware"
echo "  - Bot (Go): telegram bot logic"
echo ""
echo "🚀 Ready to build and deploy!"
