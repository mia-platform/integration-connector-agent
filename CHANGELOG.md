# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

### Chaged

- update go to v1.25.3
- update pubsub to v2.2.1
- update run to v1.12.1
- update storage to v1.57.0
- update azcore to v1.19.1
- update azidentity to v1.13.0
- update azeventhubs to v2.0.1
- update azblob to v1.6.3
- update aws-sdk-go-v2 to v1.39.3
- update config to v1.31.13
- update credentials to v1.18.17
- update lambda to v1.78.1
- update s3 to v1.88.5
- update sqs to v1.42.9
- update confluent-kafka-go to v2.12.0
- update gswagger to v0.10.1
- update oauth2 to v0.32.0
- update api to v0.252.0
- update grpc to v1.76.0

## [0.5.5] - 2025-10-15

### Added

#### New Integration Sources

- **Confluence Integration**: Full support for Confluence pages and spaces with configurable authentication and event tracking
- **Enhanced GitHub Integration**: Improved GitHub integration with app-based authentication support
  and comprehensive event types
- **Enhanced GitLab Integration**: Extended GitLab integration with web hook support and CI/CD monitoring capabilities
- **Azure DevOps Integration**: New integration for Azure DevOps pipeline and project monitoring

#### New Item Types and Schemas

- **Azure Item Types**: Support for Azure resource management with comprehensive schemas
  - `azure-resource.json` for Azure resource tracking
- **Confluence Item Types**: Complete Confluence integration schemas
  - `confluence-page.json` for page tracking
  - `confluence-space.json` for space management
- **GitHub Item Types**: Extended GitHub schema support
  - `github-issue.json` for issue tracking
  - `github-member.json` for team management
  - `github-pull-request.json` for PR workflow
  - `github-release.json` for release management
  - `github-repository.json` for repository metadata
  - `github-star.json` for star tracking
  - `github-workflow-run.json` for CI/CD monitoring
- **GitLab Item Types**: Comprehensive GitLab integration schemas
  - `gitlab-merge-request.json` for MR workflow
  - `gitlab-pipeline.json` for CI/CD tracking
  - `gitlab-project.json` for project management
  - `gitlab-release.json` for release tracking

#### Enhanced Features

- **Cloud Vendor Aggregator**: New Azure processor for cloud resource aggregation
- **Console Catalog Sink**: Enhanced console catalog integration with improved client interface
- **Authentication Migration**: Added GitHub authentication migration documentation
- **Azure Permissions**: New troubleshooting documentation for Azure permissions

#### Configuration Examples

- Complete Azure pipeline configurations (simple, conditional, and complete)
- Confluence integration configuration examples
- GitHub app and token-based authentication examples
- GitLab CI monitoring and web hook configuration examples

### Enhanced

- Improved error handling and logging across all integrations
- Enhanced web hook processing capabilities
- Better configuration validation and schema support
- Improved test coverage for all new integrations

### Documentation

- Updated source documentation for all new integrations
- Added troubleshooting guides for Azure permissions
- Enhanced GitHub authentication migration guide
- Comprehensive examples for all new configurations
