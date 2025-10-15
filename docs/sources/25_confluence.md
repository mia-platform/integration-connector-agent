# Confluence Source

The Confluence source allows you to integrate with Atlassian Confluence to receive real-time events and import existing data.

## Overview

The Confluence source provides two main functionalities:
1. **Webhook endpoint** (`/confluence/webhook`) - Receives real-time events from Confluence
2. **Import endpoint** (`/confluence/import`) - Imports existing Confluence data (spaces and pages)

## Configuration

The Confluence source can be configured with the following properties:

### Basic Webhook Configuration

```json
{
  "source": {
    "type": "confluence",
    "webhookPath": "/confluence/webhook",
    "authentication": {
      "secret": "your-webhook-secret",
      "headerName": "X-Hub-Signature-256"
    }
  }
}
```

### Import Configuration

To enable import functionality, add the import webhook configuration:

```json
{
  "source": {
    "type": "confluence",
    "webhookPath": "/confluence/webhook",
    "authentication": {
      "secret": "your-webhook-secret",
      "headerName": "X-Hub-Signature-256"
    },
    "importWebhookPath": "/confluence/import",
    "importAuthentication": {
      "secret": "your-import-secret",
      "headerName": "X-Hub-Signature-256"
    },
    "username": "your-confluence-username",
    "apiToken": "your-confluence-api-token",
    "baseUrl": "https://your-domain.atlassian.net"
  }
}
```

### Configuration Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `type` | string | ✅ | Must be `"confluence"` |
| `webhookPath` | string | ❌ | Webhook endpoint path (default: `/confluence/webhook`) |
| `authentication` | object | ❌ | Webhook authentication configuration |
| `authentication.secret` | string | ❌ | Secret for webhook signature verification |
| `authentication.headerName` | string | ❌ | Header name for signature (default: `X-Hub-Signature-256`) |
| `importWebhookPath` | string | ❌ | Import endpoint path |
| `importAuthentication` | object | ❌ | Import authentication configuration |
| `importAuthentication.secret` | string | ❌ | Secret for import webhook signature verification |
| `importAuthentication.headerName` | string | ❌ | Header name for import signature |
| `username` | string | ❌ | Confluence username (required for import) |
| `apiToken` | string | ❌ | Confluence API token (required for import) |
| `baseUrl` | string | ❌ | Confluence base URL (required for import) |
| `itemTypes` | string[] | ❌ | Item types to import (default: all). Supported values: `space`, `page` |

## Authentication

### Webhook Authentication

The webhook endpoint supports HMAC signature verification using the configured secret. Confluence should be configured to send the signature in the specified header.

### API Authentication

For import functionality, the source uses basic authentication with:
- **Username**: Your Confluence username or email
- **API Token**: Generate from your Atlassian account settings

## Supported Events

The Confluence source supports the following webhook events:

### Page Events
- `page_created` - When a page is created
- `page_updated` - When a page is updated
- `page_moved` - When a page is moved
- `page_removed` - When a page is deleted

### Blog Events
- `blog_created` - When a blog post is created
- `blog_updated` - When a blog post is updated
- `blog_removed` - When a blog post is deleted

### Space Events
- `space_created` - When a space is created
- `space_updated` - When a space is updated
- `space_removed` - When a space is deleted

### Comment Events
- `comment_created` - When a comment is created
- `comment_updated` - When a comment is updated
- `comment_removed` - When a comment is deleted

### Other Events
- `attachment_created`, `attachment_updated`, `attachment_removed` - Attachment events
- `user_created`, `user_updated`, `user_removed`, `user_deactivated` - User events
- `label_created`, `label_removed` - Label events
- `like_created`, `like_removed` - Like events
- `template_created`, `template_updated`, `template_removed` - Template events
- `group_created`, `group_removed` - Group events

## Import Functionality

The import functionality allows you to fetch existing Confluence data:

### Resources Imported

1. **Spaces (Workspaces)** - All spaces in your Confluence instance
2. **Pages** - All pages within each space

### Filtering Import Types

You can control which types of items are imported by specifying the `itemTypes` configuration:

- **Default behavior**: If `itemTypes` is not specified, all supported item types are imported (spaces and pages)
- **Selective import**: Specify only the item types you need to reduce import time and data volume

Example configurations:
- `"itemTypes": ["space", "page"]` - Import both spaces and pages (same as default)
- `"itemTypes": ["space"]` - Import only spaces
- `"itemTypes": ["page"]` - Import only pages

### Triggering Import

Send a POST request to the import endpoint to trigger a full import:

```bash
curl -X POST "https://your-agent.com/confluence/import" \
  -H "X-Hub-Signature-256: sha256=your-signature"
```

### Import Process

1. **List Spaces**: Fetches all spaces using the Confluence API
2. **Import Spaces**: Creates import events for each space
3. **List Pages**: For each space, fetches all pages
4. **Import Pages**: Creates import events for each page

## Error Handling

The source handles various error conditions:
- Invalid authentication credentials
- Network connectivity issues
- Confluence API rate limits
- Invalid webhook signatures

Errors are logged with appropriate context for debugging.

## Example Configuration

Complete example configuration:

```json
{
  "integrations": [
    {
      "source": {
        "type": "confluence",
        "webhookPath": "/confluence/webhook",
        "authentication": {
          "secret": "my-webhook-secret",
          "headerName": "X-Hub-Signature-256"
        },
        "importWebhookPath": "/confluence/import",
        "importAuthentication": {
          "secret": "my-import-secret",
          "headerName": "X-Hub-Signature-256"
        },
        "username": "john.doe@company.com",
        "apiToken": "ATATT3xFfGF0...",
        "baseUrl": "https://company.atlassian.net",
        "itemTypes": ["space", "page"]
      },
      "pipelines": [
        {
          "processors": [
            {
              "type": "mapper",
              "outputEvent": {
                "type": "confluence-event",
                "data": "{{ .body }}"
              }
            }
          ],
          "sinks": [
            {
              "type": "console-catalog"
            }
          ]
        }
      ]
    }
  ]
}
```

## Confluence API Version

This source uses Confluence Cloud REST API v2. Make sure your Confluence instance supports this API version.

## Rate Limits

Be aware of Confluence API rate limits:
- Standard rate limit: 10 requests per second
- Import operations may take time for large instances

The source implements appropriate error handling and retry logic for rate-limited requests.