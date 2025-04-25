# Thread Art Generator

![Thread Art Generator](https://github.com/Damione1/thread-art-generator/assets/14912510/6b6ef9e1-9bad-4dd7-8579-17fe55ae9c13)

[![Go Report Card](https://goreportcard.com/badge/github.com/Damione1/thread-art-generator)](https://goreportcard.com/report/github.com/Damione1/thread-art-generator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Thread Art Generator transforms your images into unique pieces of circular thread art. Upload images, customize settings, create compositions, and generate physical thread art with visualization and machine instructions (GCode).

## Features

- **Image Transformation**: Convert regular images into thread art designs
- **Composition Creation**: Design and compare multiple thread art compositions
- **Physical Output**: Generate GCode for creating thread art with physical machines
- **Customization Options**:
  - Configurable number of nails around the circular board
  - Adjustable image size affecting detail level
  - Maximum thread lines control
  - Randomized starting positions
  - Brightness and contrast adjustments
- **Command Line Interface**: Manage arts and generate thread designs directly from the terminal

## Architecture

```
[Web UI] <--> [API Server] <--> [Queue] <--> [Worker Service]
              ^    |                                |
              |    v                                v
            [CLI]  [Database] <--------------> [Storage Bucket]
```

- **API Server**: Handles user requests, manages art/composition metadata
- **Queue**: Manages composition processing tasks (RabbitMQ)
- **Worker Service**: Processes compositions using thread_generator
- **Database**: Stores metadata (PostgreSQL)
- **Storage**: Stores images and generation results (Object Storage)
- **Web UI**: Next.js frontend for user interaction
- **CLI**: Command-line interface for managing arts and generating designs

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.22+
- Tilt (for local development)

### Quick Start

1. **Initial Setup**

```bash
# Run the one-time setup script to configure your local environment
make setup
```

This will:

- Check for required tools
- Set up SSL certificates
- Create an `.env` file with generated keys
- Configure local hostnames
- Build the CLI tool

2. **Start Development Environment**

```bash
# Start all services with Tilt
make up

# Or directly
tilt up
```

3. **Access the Application**

- Web UI: https://tag.local
- API: https://tag.local/grpc-api
- MinIO Console: http://localhost:9001 (credentials in .env)

## Development

### Project Structure

- `/cmd` - Application entry points
- `/core` - Core business logic
- `/proto` - Protocol buffer definitions
- `/web` - Next.js frontend
- `/threadGenerator` - Thread art generation algorithm
- `/infra` - Infrastructure configuration
- `/scripts` - Utility scripts

### Working with the CLI

The Thread Art Generator includes a CLI tool for managing arts and generating thread designs directly from the terminal.

#### Running the CLI

```bash
# Using the wrapper script (recommended)
./scripts/cli <command>

# Or build and run directly
go run ./cmd/cli <command>
```

#### Available Commands

- **Authentication**

  ```bash
  # Log in with Auth0
  ./scripts/cli login

  # Log out and clear credentials
  ./scripts/cli logout

  # Check connection status
  ./scripts/cli status
  ```

- **Art Management**

  ```bash
  # List all your arts
  ./scripts/cli arts list

  # Get details for a specific art
  ./scripts/cli arts get <art-id>

  # Create a new art
  ./scripts/cli arts create "My Artwork Title"

  # Delete an art
  ./scripts/cli arts delete <art-id>
  ```

- **Thread Art Generation**
  ```bash
  # Generate a thread art design from an existing art
  ./scripts/cli generate <art-id>
  ```

#### Examples

```bash
# Create a new art
./scripts/cli arts create "Sunset Portrait"

# List your arts
./scripts/cli arts list

# Generate a thread art design
./scripts/cli generate arts/1234567890
```

### Development Commands

```bash
# Restart all services
make restart

# Stop all services
make down

# Access PostgreSQL directly
make psql

# Run manual database migrations
tilt trigger migrations

# Rebuild proto files after changes
tilt trigger proto-rebuild
```

### Database Access

Connect to the database using:

```bash
docker-compose exec db psql local -U local -d local
```

### Local HTTPS Development

For local development with Auth0:

1. Generate local certificates:

   ```bash
   tilt trigger setup-mkcert
   ```

2. Start the development environment:

   ```bash
   tilt up
   ```

3. Access at https://tag.local

4. For Auth0 integration, add `https://tag.local/callback` to your Auth0 application's Allowed Callback URLs.

The Traefik dashboard is available at http://localhost:8080/dashboard/.

## Storage Options

The application supports multiple storage providers:

- **Local MinIO** (development): Configured automatically
- **GCS** (production): Requires GCP credentials
- **S3** (production): Requires AWS credentials

Configure in the `.env` file.

## Production Deployment

Production deployment instructions are available in the `/infra/README.md` file.

## Hardware

This project includes designs for physical thread art machinery. The schematics and designs are sourced from the StringArt project by [Bdring](https://github.com/bdring/StringArt).

### FluidNC Configuration

The project includes configuration files for FluidNC, a high-performance Grbl CNC firmware for ESP32 microcontrollers. More information can be found on their [official GitHub page](https://github.com/bdring/FluidNC).

## Roadmap

- [x] Core thread art algorithm
- [x] Basic web interface
- [x] API server with composition storage
- [x] Worker service for async processing
- [x] UI with real-time previews and visualization
- [x] CLI tool for terminal-based interactions
- [⏳] GCode generator for thread path creation (In Progress)
- [⏳] Enhanced customization settings (In Progress)
- [ ] Connect-RPC Migration
  - [x] API Server Connect handler setup
  - [x] Update interceptors to Connect middleware
  - [x] Update proto generation configuration
  - [x] Update client implementations
  - [ ] Remove Envoy and gRPC Gateway dependencies
  - [ ] Update documentation
- [ ] CLI Enhancements
  - [ ] Support for composition management
  - [ ] Interactive mode for settings configuration
  - [ ] Offline thread generation capability
- [ ] Testing implementation
  - [ ] Unit tests for core services
  - [ ] Integration tests for API endpoints
  - [ ] Performance testing for thread generation
  - [ ] End-to-end testing

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Bdring's StringArt project](https://github.com/bdring/StringArt) for hardware designs
- [FluidNC](https://github.com/bdring/FluidNC) for CNC firmware
