## Purpose

Provide a container image that builds the ddg binary and serves it as an MCP server over StreamableHTTP on port 8080.

## Requirements

### Requirement: Container serves MCP over HTTP on port 8080
The container image SHALL start the ddg MCP server in StreamableHTTP mode and listen on port 8080 inside the container.

#### Scenario: Server listens on 8080
- **WHEN** the container is started with its default entrypoint
- **THEN** the ddg server listens on `:8080` and accepts MCP StreamableHTTP connections

#### Scenario: Startup log identifies http transport
- **WHEN** the container starts
- **THEN** stderr receives the ddg startup line naming the `http` transport and the resolved listen address `:8080`

### Requirement: Static binary in scratch runtime
The image SHALL run on the `scratch` base image with a statically linked ddg binary that has no runtime dependency on a C library or CGO.

#### Scenario: Binary is statically linked
- **WHEN** the image is built
- **THEN** the contained ddg binary is a statically linked executable with no dynamic library dependencies

### Requirement: CA certificates bundled
The image SHALL include a system CA certificate bundle so the server's HTTPS requests complete TLS verification successfully.

#### Scenario: HTTPS fetch succeeds
- **WHEN** the server's `web_read` tool fetches an HTTPS URL
- **THEN** TLS certificate verification succeeds using the bundled CA bundle

### Requirement: Non-root execution
The container SHALL run its process as a non-root user.

#### Scenario: Process is non-root
- **WHEN** the container is started
- **THEN** the ddg process runs with a non-zero user ID

### Requirement: Version injection
The image build SHALL support injecting the ddg version so the running server reports the intended version.

#### Scenario: Build-time version reported
- **WHEN** the image is built with a version argument
- **THEN** the ddg server startup log and version output report that version
