#!/bin/bash

# LLM Evaluation Service Test Script

BASE_URL="http://localhost:8080"

if [ $# -eq 1 ]; then
    BASE_URL=$1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Test result counters
PASSED=0
FAILED=0

# Function to test HTTP response
test_request() {
    local test_name="$1"
    local expected_status="$2"
    local method="$3"
    local url="$4"
    local data="$5"
    
    echo "Testing: $test_name"
    echo "Expected status: $expected_status"
    
    # Build curl command
    if [ -n "$data" ]; then
        # POST request with data
        response=$(curl -s -w '%{http_code}' -X "$method" -H "Content-Type: application/json" -d "$data" "$url")
    else
        # GET request or other methods without data
        response=$(curl -s -w '%{http_code}' -X "$method" "$url")
    fi
    
    status_code="${response: -3}"
    body="${response%???}"
    
    echo "Actual status: $status_code"
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} - Status code matches"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC} - Expected $expected_status, got $status_code"
        FAILED=$((FAILED + 1))
    fi
    
    # Pretty print JSON if it's valid
    if echo "$body" | jq . >/dev/null 2>&1; then
        echo "Response:"
        echo "$body" | jq .
    else
        echo "Response body:"
        echo "$body"
    fi
    
    echo ""
}

echo "Testing LLM Evaluation Service at $BASE_URL"
echo "=================================================="

# Test 1: Health Check
echo "Test 1: Health Check"
echo "--------------------"
test_request "Health endpoint" "200" "GET" "$BASE_URL/health"

# Test 2: Simple Evaluation
echo "Test 2: Simple Evaluation"
echo "-------------------------"
test_request "Simple evaluation" "200" "POST" "$BASE_URL/evaluate" '{"input": "What is the capital of France?", "output": "The capital of France is Paris.", "criteria": "The response should be factually correct and directly answer the question."}'

# Test 3: Mathematical Evaluation
echo "Test 3: Mathematical Evaluation"
echo "-------------------------------"
test_request "Mathematical evaluation" "200" "POST" "$BASE_URL/evaluate" '{"input": "Calculate 15 * 24", "output": "15 * 24 = 360", "criteria": "The mathematical calculation should be correct."}'

# Test 4: Complex Evaluation
echo "Test 4: Complex Evaluation"
echo "--------------------------"
test_request "Complex evaluation" "200" "POST" "$BASE_URL/evaluate" '{"input": "Explain machine learning", "output": "Machine learning is like teaching a computer to learn patterns", "criteria": "The explanation should be age-appropriate and accurate."}'

# Test 5: Error Cases
echo "Test 5: Error Cases"
echo "-------------------"

echo "5a. Missing required field (should return 400):"
test_request "Missing output field" "400" "POST" "$BASE_URL/evaluate" '{"input": "What is 2+2?", "criteria": "Should be correct"}'

echo "5b. Invalid JSON (should return 400):"
test_request "Invalid JSON" "400" "POST" "$BASE_URL/evaluate" '{"invalid": json}'

echo "5c. Empty payload (should return 400):"
test_request "Empty payload" "400" "POST" "$BASE_URL/evaluate" ""

# Test 6: Method Not Allowed
echo "Test 6: Method Not Allowed"
echo "--------------------------"
echo "6a. GET request to /evaluate (should return 405):"
test_request "GET to /evaluate" "405" "GET" "$BASE_URL/evaluate"

echo "6b. POST request to /health (should return 405):"
test_request "POST to /health" "405" "POST" "$BASE_URL/health"

# Test 7: Load Test (Basic)
echo "Test 7: Basic Concurrency Test"
echo "----------------------"
echo "Sending 10 concurrent requests..."

for i in {1..10}; do
    curl -s -X POST "$BASE_URL/evaluate" \
        -H "Content-Type: application/json" \
        -d '{
            "input": "Test input '$i'",
            "output": "Test output '$i'",
            "criteria": "Should be reasonable"
        }' > /dev/null &
done

wait
echo "Load test completed!"
echo ""

# Summary
echo "===================="
echo "Test Summary"
echo "===================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed${NC}"
    exit 1
fi 
