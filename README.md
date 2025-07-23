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

## Architecture

```
[Web UI] <--> [API Server] <--> [Queue] <--> [Worker Service]
                   |                                |
                   v                                v
               [Database] <--------------> [Storage Bucket]
```

- **API Server**: Handles user requests, manages art/composition metadata
- **Queue**: Manages composition processing tasks (RabbitMQ)
- **Worker Service**: Processes compositions using thread_generator
- **Database**: Stores metadata (PostgreSQL)
- **Storage**: Stores images and generation results (Object Storage)
- **Web UI**: Go+HTMX frontend for user interaction

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

- Check for required tools (Docker, Tilt, Node.js, Go, etc.)
- Create an `.env` file with generated keys
- Install protocol buffer tools
- Set up frontend dependencies (Tailwind CSS, Templ)

2. **Start Development Environment**

```bash
# Start all services with Tilt
make up

# Or directly
tilt up
```

3. **Access the Application**

- Web UI: http://localhost:8080
- API: http://localhost:9090
- MinIO Console: http://localhost:9001 (credentials in .env)
- RabbitMQ Management: http://localhost:15672 (guest/guest)
- Firebase Emulator UI: http://localhost:4000

## Development

### Project Structure

- `/cmd` - Application entry points (api, worker, migrations)
- `/core` - Core business logic and shared libraries
  - `/auth` - Firebase authentication
  - `/db` - Database models and migrations
  - `/service` - Business logic services
  - `/storage` - Blob storage abstraction (MinIO/GCS)
  - `/pb` - Generated protocol buffer code
- `/client` - Go+HTMX frontend application
  - `/cmd/frontend` - Frontend server entry point
  - `/internal` - Frontend-specific code (handlers, templates, services)
  - `/public` - Static assets (CSS, JS, images)
- `/proto` - Protocol buffer definitions
- `/threadGenerator` - Thread art generation algorithm
- `/functions` - Firebase Functions (TypeScript)
- `/Infra` - Infrastructure configuration (Dockerfiles, Terraform)
- `/scripts` - Utility scripts and CLI tools


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

# Generate protocol buffer files (recommended)
make proto

# Clean generated protocol buffer files
make proto-clean
```

### Protocol Buffer Generation

The project uses [Connect-RPC](https://connectrpc.com/) for API communication. When you modify `.proto` files, you need to regenerate the Go and Connect-RPC code:

```bash
# Generate all protocol buffer files
make proto
```

This command will:
- Auto-install required tools (protoc-gen-go, protoc-gen-connect-go, protoc-gen-openapiv2)
- Generate Go types in `core/pb/`
- Generate Connect-RPC clients/servers in `core/pb/pbconnect/`
- Generate OpenAPI documentation in `api/openapi/`

**Requirements:**
- [Buf CLI](https://buf.build/docs/installation) - Protocol buffer build tool
- Go 1.22+ - For installing protoc generators

The generation uses Connect-RPC v2 (`connectrpc.com/connect`) for modern, efficient RPC communication.

### Database Access

Connect to the database using:

```bash
docker-compose exec db psql local -U local -d local
```

### Firebase Authentication Setup

The application uses Firebase Authentication for user management:

1. **Development**: Firebase emulator runs automatically with `tilt up`
   - Auth Emulator: http://localhost:9099
   - Functions Emulator: http://localhost:5001  
   - Emulator UI: http://localhost:4000

2. **Production**: Configure Firebase project credentials in `.env`
   - Set `FIREBASE_PROJECT_ID`
   - Set `FIREBASE_WEB_API_KEY` 
   - Set `FIREBASE_AUTH_DOMAIN`

## Storage Options

The application supports multiple storage providers:

- **Local MinIO** (development): Configured automatically with dual-bucket setup
- **Google Cloud Storage (GCS)** (production): Requires GCP credentials and project configuration

Configure storage provider in the `.env` file using `STORAGE_PROVIDER=minio` for development or `STORAGE_PROVIDER=gcs` for production.

## Production Deployment

Production deployment instructions are available in the `/infra/README.md` file.

## Hardware

This project includes designs for physical thread art machinery. The schematics and designs are sourced from the StringArt project by [Bdring](https://github.com/bdring/StringArt).

### FluidNC Configuration

The project includes configuration files for FluidNC, a high-performance Grbl CNC firmware for ESP32 microcontrollers. More information can be found on their [official GitHub page](https://github.com/bdring/FluidNC).

## Roadmap

### ✅ Completed Features
- [x] Core thread art algorithm
- [x] Basic web interface  
- [x] API server with composition storage
- [x] Worker service for async processing
- [x] UI with real-time previews and visualization
- [x] **GCode generator for thread path creation** 
- [x] **Enhanced customization settings**
- [x] **CDN support for image storage and caching**
- [x] **Migration from Next.js to Templ (full Go infrastructure)**
- [x] **Addition of HTMX for frontend interactivity**
- [x] **Compiled JavaScript integration** 
- [x] **BFF (Backend for Frontend) setup**
- [x] **Migration from Auth0 to Firebase Authentication**
- [x] **Firebase Functions for user creation and management**
- [x] Connect-RPC Migration
  - [x] API Server Connect handler setup
  - [x] Update interceptors to Connect middleware  
  - [x] Update proto generation configuration
  - [x] Update client implementations
  - [x] Remove Envoy and gRPC Gateway dependencies

### 🚧 In Progress / Todo
- [ ] **Infrastructure & Deployment**
  - [ ] Terraform infrastructure setup
  - [ ] Automated deployment pipeline (CI/CD)
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
