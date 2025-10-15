# GitHub Event Type Schemas

This directory contains JSON schemas for different GitHub event types that can be processed by the GitHub source in the integration connector agent.

## Available Schemas

### Core Repository Events

1. **github-repository.json** - Repository events (created, edited, deleted, etc.)
   - Used for: `repository` event type
   - Contains: Repository metadata, settings, owner information, organization details

2. **github-repo.json** - Legacy repository schema (existing)
   - Enhanced repository schema with additional fields for import scenarios

### Code & Development Events

3. **github-push.json** - Push events 
   - Used for: `push` event type
   - Contains: Commit information, branch details, diff data, pusher information

4. **github-pull-request.json** - Pull request events
   - Used for: `pull_request` event type  
   - Contains: PR details, state, reviewers, commits, changes, author information

5. **github-issue.json** - Issue events
   - Used for: `issues` event type
   - Contains: Issue details, state, labels, assignees, milestone information

### Release & Deployment Events

6. **github-release.json** - Release events
   - Used for: `release` event type
   - Contains: Release information, assets, tags, publication details

7. **github-workflow-run.json** - GitHub Actions workflow events
   - Used for: `workflow_run` event type
   - Contains: Workflow execution details, status, conclusion, job information

### Collaboration Events

8. **github-fork.json** - Fork events
   - Used for: `fork` event type
   - Contains: Original and forked repository information, fork relationship

9. **github-star.json** - Star/watch events
   - Used for: `star`, `watch` event types
   - Contains: Repository and user information for starring/watching actions

10. **github-member.json** - Repository member events
    - Used for: `member` event type
    - Contains: Member details, permission changes, collaboration information

### Reference Events

11. **github-create.json** - Branch/tag creation events
    - Used for: `create` event type
    - Contains: Reference information, branch/tag details, creation context

12. **github-delete.json** - Branch/tag deletion events
    - Used for: `delete` event type
    - Contains: Reference information, deletion context

## Common Fields

All schemas include these standard fields:

- `eventType`: The GitHub event type that triggered the data
- `source`: Data source identifier ("github-webhook" or "github-import")
- `repositoryId`, `repositoryName`: Basic repository identification
- `organizationLogin`: Organization context when applicable
- `senderLogin`, `senderType`: User who triggered the event

## Usage

These schemas can be used for:

1. **Data Validation**: Ensure incoming GitHub data matches expected structure
2. **Console Catalog**: Define item type definitions for storing GitHub entities
3. **Documentation**: Understanding the data structure for each event type
4. **Mapping Configuration**: Reference for mapper processor field mappings

## Event Type Mapping

The schemas correspond to GitHub webhook events as defined in `internal/sources/github/events.go`:

- Repository events → `github-repository.json`
- Pull request events → `github-pull-request.json`  
- Issue events → `github-issue.json`
- Release events → `github-release.json`
- Workflow events → `github-workflow-run.json`
- Fork events → `github-fork.json`
- Star/watch events → `github-star.json`
- Member events → `github-member.json`
- Create events → `github-create.json`
- Delete events → `github-delete.json`
- Push events → `github-push.json`

Additional event types can be added as needed following the same pattern.