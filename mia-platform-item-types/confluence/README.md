# Confluence Event Type Schemas

This directory contains JSON schemas for different Confluence event types that can be processed by the Confluence source in the integration connector agent.

## Schemas

### Resource Schemas

1. **confluence-space.json** - Confluence space (workspace) events

Contains the schema for Confluence space objects including:
- Space metadata (id, key, name, type, status)
- Description and homepage information
- Creation timestamps and web URLs

2. **confluence-page.json** - Confluence page events

Contains the schema for Confluence page objects including:
- Page metadata (id, type, status, title)
- Space association and parent page relationships
- Version information and author details
- Page content in storage and Atlas document formats
- Creation timestamps and web URLs

## Usage

These schemas are used by the Confluence source to validate and process events from:

### Webhook Events
- Page lifecycle events (created, updated, moved, removed)
- Space lifecycle events (created, updated, removed) 
- Comment events (created, updated, removed)
- Blog post events (created, updated, removed)
- Attachment events (created, updated, removed)
- User and group management events
- Label and like events

### Import Events
- Spaces (workspaces) imported via API
- Pages imported via API

## Event Processing

The schemas ensure that both real-time webhook events and imported data follow a consistent structure for downstream processing by mappers, filters, and sinks in the integration pipeline.