# GitHub Authentication Migration Guide

This document explains how to use the new GitHub App authentication method instead of Personal Access Tokens.

## New GitHub App Authentication (Recommended)

### Benefits
- More secure than Personal Access Tokens
- Better audit trail and access control
- Can be scoped to specific organizations
- Supports OAuth 2.0 client credentials flow

### Setup Steps

1. **Create a GitHub App**:
   - Go to GitHub Settings > Developer settings > GitHub Apps
   - Click "New GitHub App"
   - Fill in the app details:
     - App name: "My Integration Connector Agent"
     - Homepage URL: Your application URL
     - Webhook URL: `https://your-domain.com/github/webhook`
   
2. **Set Permissions**:
   - Repository permissions:
     - Contents: Read
     - Metadata: Read
     - Pull requests: Read
     - Issues: Read
     - Actions: Read
   - Organization permissions:
     - Members: Read

3. **Get Credentials**:
   - Note down the Client ID
   - Generate a new Client Secret
   - Store both securely

4. **Configure the Agent**:
   ```json
   {
     "type": "github",
     "clientId": {
       "fromEnv": "GITHUB_CLIENT_ID"
     },
     "clientSecret": {
       "fromEnv": "GITHUB_CLIENT_SECRET"
     },
     "organization": "your-org-name"
   }
   ```

5. **Set Environment Variables**:
   ```bash
   export GITHUB_CLIENT_ID="your_client_id"
   export GITHUB_CLIENT_SECRET="your_client_secret"
   ```

## Legacy Personal Access Token (Deprecated)

If you need to continue using Personal Access Token authentication:

1. **Create a Personal Access Token**:
   - Go to GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
   - Generate new token with scopes:
     - `repo` (for private repositories)
     - `public_repo` (for public repositories)
     - `read:org` (for organization access)

2. **Configure the Agent**:
   ```json
   {
     "type": "github",
     "token": {
       "fromEnv": "GITHUB_API_TOKEN"
     },
     "organization": "your-org-name"
   }
   ```

3. **Set Environment Variable**:
   ```bash
   export GITHUB_API_TOKEN="your_token_here"
   ```

## Migration Path

To migrate from Personal Access Token to GitHub App:

1. Set up the GitHub App as described above
2. Update your configuration to use `clientId` and `clientSecret` instead of `token`
3. Update your environment variables
4. Test the import functionality: `./scripts/github-import.sh --no-auth --verbose`
5. Remove the old `GITHUB_API_TOKEN` environment variable

## Example Configurations

See the example configuration files:
- `examples/github-app-config.json` - GitHub App authentication
- `examples/github-token-config.json` - Legacy token authentication

## Script Usage

The import script works with both authentication methods:

```bash
# Trigger import (no authentication required for the script itself)
./scripts/github-import.sh --no-auth --verbose

# With HMAC authentication for the import endpoint
./scripts/github-import.sh --secret "your-import-secret" --verbose
```

## Troubleshooting

### GitHub App Authentication Issues

1. **OAuth Error**: Verify Client ID and Client Secret are correct
2. **Access Denied**: Check that the GitHub App has proper permissions
3. **Organization Access**: Ensure the GitHub App is installed in your organization

### Legacy Token Issues

1. **403 Errors**: Check token permissions and scopes
2. **Rate Limiting**: GitHub Apps have higher rate limits than Personal Access Tokens
3. **Token Expiration**: Personal Access Tokens can expire, GitHub Apps use refresh tokens

### General Issues

1. **Import Fails**: Check organization name is correctly set
2. **No Data**: Verify the organization has repositories with the expected resources
3. **Network Issues**: Ensure the agent can reach `api.github.com`
