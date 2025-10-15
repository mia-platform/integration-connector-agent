#!/bin/bash

# GitHub Import Script for integration-connector-agent
# This script triggers a full import of GitHub resources

set -e

# Configuration
DEFAULT_HOST="localhost"
DEFAULT_PORT="8080"
DEFAULT_PATH="/github/import"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Triggers a GitHub full import via the integration-connector-agent"
    echo ""
    echo "Options:"
    echo "  -h, --host HOST           Server host (default: $DEFAULT_HOST)"
    echo "  -p, --port PORT           Server port (default: $DEFAULT_PORT)"
    echo "  --path PATH               Import endpoint path (default: $DEFAULT_PATH)"
    echo "  -s, --secret SECRET       HMAC secret for authentication"
    echo "  --secret-env VAR          Environment variable containing the HMAC secret"
    echo "  --no-auth                 Skip authentication (for development)"
    echo "  -v, --verbose             Verbose output"
    echo "  --help                    Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                          # Simple call without auth"
    echo "  $0 --secret mysecret                       # With secret"
    echo "  $0 --secret-env GITHUB_IMPORT_SECRET       # Using env variable"
    echo "  $0 -h api.example.com -p 443 --secret xyz  # Custom host/port"
    echo ""
}

# Function to generate HMAC SHA-256 signature
generate_signature() {
    local secret="$1"
    local payload="$2"

    # Try different methods to generate HMAC signature, using various paths

    # Method 1: Try OpenSSL in common locations
    for openssl_path in "/opt/homebrew/bin/openssl" "/usr/bin/openssl" "openssl"; do
        if $openssl_path version >/dev/null 2>&1; then
            # Try cut first, then awk
            if command -v cut >/dev/null 2>&1; then
                echo -n "$payload" | $openssl_path dgst -sha256 -hmac "$secret" | cut -d' ' -f2
                return 0
            elif command -v awk >/dev/null 2>&1; then
                echo -n "$payload" | $openssl_path dgst -sha256 -hmac "$secret" | awk '{print $2}'
                return 0
            fi
        fi
    done

    # Method 2: Try Python 3 in common locations
    for python3_path in "/Users/giulioroggero/.pyenv/shims/python3" "/usr/bin/python3" "python3"; do
        if $python3_path --version >/dev/null 2>&1; then
            $python3_path -c "
import hmac
import hashlib
secret = '$secret'.encode('utf-8')
payload = '$payload'.encode('utf-8')
signature = hmac.new(secret, payload, hashlib.sha256).hexdigest()
print(signature)
"
            return 0
        fi
    done

    # Method 3: Try Python in common locations
    for python_path in "/Users/giulioroggero/.pyenv/shims/python" "/usr/bin/python" "python"; do
        if $python_path --version >/dev/null 2>&1; then
            $python_path -c "
import hmac
import hashlib
secret = '$secret'.encode('utf-8')
payload = '$payload'.encode('utf-8')
signature = hmac.new(secret, payload, hashlib.sha256).hexdigest()
print(signature)
"
            return 0
        fi
    done

    # Method 4: Try Node.js in common locations
    for node_path in "/Users/giulioroggero/.nvm/versions/node/v24.2.0/bin/node" "/usr/bin/node" "/usr/local/bin/node" "node"; do
        if $node_path --version >/dev/null 2>&1; then
            $node_path -e "
const crypto = require('crypto');
const secret = '$secret';
const payload = '$payload';
const signature = crypto.createHmac('sha256', secret).update(payload).digest('hex');
console.log(signature);
"
            return 0
        fi
    done

    # If all methods fail
    error "No suitable tool found for HMAC generation."
    error "Tried: openssl, python3, python, node"
    error "Please ensure one of these tools is available or use --no-auth flag"
    return 1
}

# Function to log messages
log() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${YELLOW}[INFO]${NC} $1"
    fi
}

# Function to log errors
error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Function to log success
success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Parse command line arguments
HOST="$DEFAULT_HOST"
PORT="$DEFAULT_PORT"
PATH="$DEFAULT_PATH"
SECRET=""
SECRET_ENV=""
NO_AUTH="false"
VERBOSE="false"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            HOST="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        --path)
            PATH="$2"
            shift 2
            ;;
        -s|--secret)
            SECRET="$2"
            shift 2
            ;;
        --secret-env)
            SECRET_ENV="$2"
            shift 2
            ;;
        --no-auth)
            NO_AUTH="true"
            shift
            ;;
        -v|--verbose)
            VERBOSE="true"
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Build URL
URL="http://${HOST}:${PORT}${PATH}"

log "Target URL: $URL"

# Determine secret to use
FINAL_SECRET=""
if [[ "$NO_AUTH" == "true" ]]; then
    log "Authentication disabled"
elif [[ -n "$SECRET" ]]; then
    FINAL_SECRET="$SECRET"
    log "Using provided secret"
elif [[ -n "$SECRET_ENV" ]]; then
    FINAL_SECRET="${!SECRET_ENV}"
    if [[ -z "$FINAL_SECRET" ]]; then
        error "Environment variable $SECRET_ENV is not set or empty"
        exit 1
    fi
    log "Using secret from environment variable: $SECRET_ENV"
else
    log "No authentication configured - proceeding without HMAC signature"
fi

# Prepare curl command
CURL_ARGS=()
CURL_ARGS+=("-X" "POST")
CURL_ARGS+=("-H" "Content-Type: application/json")

# Add authentication if secret is provided
if [[ -n "$FINAL_SECRET" ]]; then
    PAYLOAD=""
    SIGNATURE=$(generate_signature "$FINAL_SECRET" "$PAYLOAD")
    CURL_ARGS+=("-H" "X-Hub-Signature-256: sha256=$SIGNATURE")
    log "Generated HMAC signature"
fi

# Add verbose flag if requested
if [[ "$VERBOSE" == "true" ]]; then
    CURL_ARGS+=("-v")
fi

# Add URL
CURL_ARGS+=("$URL")

log "Executing curl command..."

# Execute the request
if /usr/bin/curl --version >/dev/null 2>&1 || curl --version >/dev/null 2>&1; then
    # Use full path if available, otherwise try curl from PATH
    CURL_CMD="/usr/bin/curl"
    if ! $CURL_CMD --version >/dev/null 2>&1; then
        CURL_CMD="curl"
    fi

    if $CURL_CMD "${CURL_ARGS[@]}"; then
        echo ""
        success "GitHub import triggered successfully!"
        success "Check the server logs for import progress and results."
    else
        echo ""
        error "Failed to trigger GitHub import. Check the server logs for details."
        exit 1
    fi
elif command -v wget >/dev/null 2>&1; then
    # Fallback to wget if curl is not available
    WGET_ARGS=()
    WGET_ARGS+=("--method=POST")
    WGET_ARGS+=("--header=Content-Type: application/json")

    # Add authentication if secret is provided
    if [[ -n "$FINAL_SECRET" ]]; then
        WGET_ARGS+=("--header=X-Hub-Signature-256: sha256=$SIGNATURE")
    fi

    WGET_ARGS+=("--output-document=-")
    WGET_ARGS+=("$URL")

    if wget "${WGET_ARGS[@]}"; then
        echo ""
        success "GitHub import triggered successfully!"
        success "Check the server logs for import progress and results."
    else
        echo ""
        error "Failed to trigger GitHub import. Check the server logs for details."
        exit 1
    fi
else
    error "Neither curl nor wget is available."
    echo ""
    echo "Manual request details:"
    echo "URL: $URL"
    echo "Method: POST"
    echo "Headers:"
    echo "  Content-Type: application/json"
    if [[ -n "$FINAL_SECRET" ]]; then
        echo "  X-Hub-Signature-256: sha256=$SIGNATURE"
    fi
    echo ""
    echo "You can use any HTTP client to make this request."
    exit 1
fi
