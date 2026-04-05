#!/usr/bin/env bash
# ado-api.sh — CLI helper for testing Azure DevOps Release REST API calls
#
# Usage:
#   ./scripts/ado-api.sh GET "release/definitions"
#   ./scripts/ado-api.sh GET "release/definitions/1?expand=environments"
#   ./scripts/ado-api.sh POST "release/definitions" --body @payload.json
#   ./scripts/ado-api.sh PUT "release/definitions/1" --body @payload.json
#   ./scripts/ado-api.sh DELETE "release/definitions/1"
#
# Environment variables:
#   AZDO_ORG_SERVICE_URL        — e.g., https://dev.azure.com/myorg
#   AZDO_PERSONAL_ACCESS_TOKEN  — PAT with release management scope
#   AZDO_TEST_PROJECT           — Default project (override with --project)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Validate environment
check_env() {
    local missing=()
    [[ -z "${AZDO_ORG_SERVICE_URL:-}" ]] && missing+=("AZDO_ORG_SERVICE_URL")
    [[ -z "${AZDO_PERSONAL_ACCESS_TOKEN:-}" ]] && missing+=("AZDO_PERSONAL_ACCESS_TOKEN")

    if [[ ${#missing[@]} -gt 0 ]]; then
        echo -e "${RED}Error: Missing environment variables: ${missing[*]}${NC}" >&2
        echo "Set them in your shell or in a .env file" >&2
        exit 1
    fi
}

# Parse org name from URL
get_org() {
    echo "$AZDO_ORG_SERVICE_URL" | sed 's|https://dev.azure.com/||' | sed 's|/$||'
}

# Build the release API base URL
# Release API uses vsrm.dev.azure.com, not dev.azure.com
get_base_url() {
    local org
    org=$(get_org)
    local project="${1:-${AZDO_TEST_PROJECT:-}}"

    if [[ -z "$project" ]]; then
        echo -e "${RED}Error: No project specified. Use --project or set AZDO_TEST_PROJECT${NC}" >&2
        exit 1
    fi

    echo "https://vsrm.dev.azure.com/${org}/${project}/_apis"
}

# Build auth header
get_auth() {
    echo -n ":${AZDO_PERSONAL_ACCESS_TOKEN}" | base64
}

# Main API call function
api_call() {
    local method="$1"
    local path="$2"
    shift 2

    local project="${AZDO_TEST_PROJECT:-}"
    local body=""
    local verbose=false
    local raw=false

    # Parse optional args
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --project)
                project="$2"
                shift 2
                ;;
            --body)
                body="$2"
                shift 2
                ;;
            --verbose|-v)
                verbose=true
                shift
                ;;
            --raw)
                raw=true
                shift
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}" >&2
                exit 1
                ;;
        esac
    done

    local base_url
    base_url=$(get_base_url "$project")

    # Add api-version if not already in path
    local url="${base_url}/${path}"
    if [[ "$url" != *"api-version="* ]]; then
        if [[ "$url" == *"?"* ]]; then
            url="${url}&api-version=7.1"
        else
            url="${url}?api-version=7.1"
        fi
    fi

    local auth
    auth=$(get_auth)

    if $verbose; then
        echo -e "${BLUE}${method} ${url}${NC}" >&2
        [[ -n "$body" ]] && echo -e "${YELLOW}Body: ${body}${NC}" >&2
    fi

    # Build curl command
    local curl_args=(
        -s
        -w "\n%{http_code}"
        -X "$method"
        -H "Authorization: Basic ${auth}"
        -H "Content-Type: application/json"
        -H "Accept: application/json"
    )

    if [[ -n "$body" ]]; then
        if [[ "$body" == @* ]]; then
            curl_args+=(-d "$body")
        elif [[ "$body" == "-" ]]; then
            curl_args+=(-d @-)
        else
            curl_args+=(-d "$body")
        fi
    fi

    # Execute
    local response
    response=$(curl "${curl_args[@]}" "$url")

    # Split response body and status code
    local status_code
    status_code=$(echo "$response" | tail -1)
    local response_body
    response_body=$(echo "$response" | sed '$d')

    # Output
    if $raw; then
        echo "$response_body"
    else
        # Status indicator
        if [[ "$status_code" -ge 200 && "$status_code" -lt 300 ]]; then
            echo -e "${GREEN}✓ ${status_code}${NC}" >&2
        elif [[ "$status_code" -ge 400 ]]; then
            echo -e "${RED}✗ ${status_code}${NC}" >&2
        else
            echo -e "${YELLOW}→ ${status_code}${NC}" >&2
        fi

        # Pretty print JSON if jq is available
        if command -v jq &>/dev/null; then
            echo "$response_body" | jq . 2>/dev/null || echo "$response_body"
        else
            echo "$response_body"
        fi
    fi

    # Return appropriate exit code
    if [[ "$status_code" -ge 400 ]]; then
        return 1
    fi
}

# Convenience commands
cmd_list_definitions() {
    api_call GET "release/definitions" "$@"
}

cmd_get_definition() {
    local id="$1"
    shift
    api_call GET "release/definitions/${id}?expand=environments,artifacts,triggers,variables" "$@"
}

cmd_create_definition() {
    local body_file="$1"
    shift
    api_call POST "release/definitions" --body "@${body_file}" "$@"
}

cmd_delete_definition() {
    local id="$1"
    shift
    api_call DELETE "release/definitions/${id}" "$@"
}

# Help
show_help() {
    cat <<'HELP'
ado-api.sh — Azure DevOps Release API CLI Helper

USAGE:
    ./scripts/ado-api.sh <METHOD> <PATH> [OPTIONS]
    ./scripts/ado-api.sh <COMMAND> [ARGS] [OPTIONS]

METHODS:
    GET, POST, PUT, PATCH, DELETE

OPTIONS:
    --project <name>    Override project (default: $AZDO_TEST_PROJECT)
    --body <data>       Request body (use @file.json for file, - for stdin)
    --verbose, -v       Show request details
    --raw               Output raw response (no formatting)

SHORTCUT COMMANDS:
    list-definitions                  List all release definitions
    get-definition <id>               Get definition with full expansion
    create-definition <file.json>     Create from JSON file
    delete-definition <id>            Delete a definition

EXAMPLES:
    # List all release definitions
    ./scripts/ado-api.sh GET "release/definitions"

    # Get definition with environments expanded
    ./scripts/ado-api.sh GET "release/definitions/1?expand=environments" -v

    # Create from file
    ./scripts/ado-api.sh POST "release/definitions" --body @examples/minimal.json

    # Pipe body from stdin
    echo '{"name":"test"}' | ./scripts/ado-api.sh POST "release/definitions" --body -

    # Use shortcut
    ./scripts/ado-api.sh list-definitions --project MyProject

ENVIRONMENT:
    AZDO_ORG_SERVICE_URL         https://dev.azure.com/myorg
    AZDO_PERSONAL_ACCESS_TOKEN   Your PAT
    AZDO_TEST_PROJECT            Default project name
HELP
}

# Main
main() {
    if [[ $# -eq 0 ]] || [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
        show_help
        exit 0
    fi

    check_env

    local cmd="$1"
    shift

    case "$cmd" in
        GET|POST|PUT|PATCH|DELETE)
            api_call "$cmd" "$@"
            ;;
        list-definitions)
            cmd_list_definitions "$@"
            ;;
        get-definition)
            cmd_get_definition "$@"
            ;;
        create-definition)
            cmd_create_definition "$@"
            ;;
        delete-definition)
            cmd_delete_definition "$@"
            ;;
        *)
            echo -e "${RED}Unknown command: ${cmd}${NC}" >&2
            show_help
            exit 1
            ;;
    esac
}

main "$@"
