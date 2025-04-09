load('ext://restart_process', 'docker_build_with_restart')

# ================================================
# BUILD CONFIGURATIONS
# ================================================

# Compile Go binaries
local_resource(
  'go-compile',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/migrations cmd/migrations/main.go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/worker cmd/worker/main.go && go build -o build/cli cmd/cli/main.go',
  labels=["build"],
  deps=['cmd/', 'core/', 'threadGenerator/', 'web/**/*.go'],
)

# API image build
docker_build(
  'api-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-api',
  only=[
    './build/api',
  ],
  live_update=[
    sync('./build/api', '/app/build/api'),
    restart_container()
  ]
)

# Migrations image build
docker_build(
  'migrations-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-migrations',
  only=[
    './build/migrations',
    './core/db/migrations',
  ],
  live_update=[
    sync('./build/migrations', '/app/build/migrations'),
    sync('./core/db/migrations', '/migrations'),
  ]
)

# Worker image build
docker_build(
  'worker-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-worker',
  only=[
    './build/worker',
  ],
  live_update=[
    sync('./build/worker', '/app/worker'),
    restart_container()
  ]
)

# ================================================
# DOCKER COMPOSE CONFIGURATION
# ================================================

# Load Docker Compose configuration
docker_compose('docker-compose.yml')

# ================================================
# LOCAL DEVELOPMENT RESOURCES
# ================================================

# Proto file watcher - detects changes in proto files
local_resource(
  'proto-watch',
  cmd='echo "Proto files changed"',
  deps=['proto/'],
  labels=["proto"],
)

# Local development setup
local_resource(
  'setup-local-dev',
  cmd='./scripts/local_setup.sh',
  labels=["setup"],
  trigger_mode=TRIGGER_MODE_MANUAL,
  auto_init=False
)

# MinIO setup after container starts
local_resource(
  'setup-minio',
  cmd='./scripts/dev/tilt-minio-setup.sh',
  labels=["storage"],
  resource_deps=['minio'],
  auto_init=True
)

# ================================================
# SERVICE CONFIGURATION
# ================================================

# Define all resources with consistent configuration format
resources = {
  # Database services
  'db': {
    'labels': ['database'],
  },

  'migrations': {
    'labels': ['database'],
    'resource_deps': ['go-compile', 'db'],
    'auto_init': True,
    'trigger_mode': TRIGGER_MODE_MANUAL,
  },

  'generate-models': {
    'labels': ['database'],
    'resource_deps': ['db'],
    'auto_init': False,
    'trigger_mode': TRIGGER_MODE_MANUAL,
  },

  # Queue services
  'rabbitmq': {
    'labels': ['queue'],
    'auto_init': True,
    'trigger_mode': TRIGGER_MODE_AUTO,
    'links': [
      link('http://localhost:15672', 'RabbitMQ Management (guest/guest)'),
    ]
  },

  # Worker service
  'worker': {
    'labels': ['worker'],
    'resource_deps': ['go-compile', 'db', 'rabbitmq'],
    'auto_init': True,
    'trigger_mode': TRIGGER_MODE_AUTO,
  },

  # Proto handling
  'proto-build': {
    'labels': ['proto'],
    'resource_deps': ['proto-watch'],
    'auto_init': True,
    'trigger_mode': TRIGGER_MODE_AUTO,
  },

  # Application services
  'api': {
    'labels': ['application'],
    'resource_deps': ['go-compile', 'db', 'migrations', 'rabbitmq'],
    'trigger_mode': TRIGGER_MODE_AUTO,
  },

  'frontend': {
    'labels': ['application'],
    'resource_deps': ['api'],
    'trigger_mode': TRIGGER_MODE_AUTO,
  },

  # Networking
  'envoy': {
    'labels': ['networking'],
    'resource_deps': ['api', 'frontend'],
    'links': [
      link('https://tag.local', 'Thread Art Generator (HTTPS)'),
    ]
  },

  # Storage
  'minio': {
    'labels': ['storage'],
    'resource_deps': ['db']
  },
}

# Create resources from configuration
for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
