load('ext://restart_process', 'docker_build_with_restart')

#compile and set executable. But for the cli one, we need to build it for the current platform and make it executable
local_resource(
  'go-compile',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/migrations cmd/migrations/main.go && go build -o build/cli cmd/cli/main.go',
  labels=["scripts"],
  deps=['cmd/', 'core/', 'threadGenerator/', 'web/**/*.go'],
)


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
  ])

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
  ])


# Load the docker compose configuration
docker_compose('docker-compose.yml')

# Add proto file watching for automatic rebuilds
local_resource(
  'proto-watch',
  cmd='echo "Proto files changed - rebuilding..."',
  deps=['proto/'],
  resource_deps=['proto-rebuild'],
  labels=["auto-proto"]
)

# Combined proto rebuild resource
local_resource(
  'proto-rebuild',
  cmd='docker-compose run --rm proto-build',
  trigger_mode=TRIGGER_MODE_MANUAL,
  auto_init=False,
  labels=["auto-proto"]
)

# Function to ensure mkcert is installed
def ensure_mkcert():
    """Ensures mkcert is installed and generates certificates for tag.local"""
    local('which mkcert || brew install mkcert')
    local('mkcert -install')

    # Create certs directory if it doesn't exist
    local('mkdir -p ./certs')

    # Generate certificates for tag.local
    local('mkcert -cert-file ./certs/tag.local.crt -key-file ./certs/tag.local.key tag.local "*.tag.local"')

    # Add tag.local to /etc/hosts if not already present
    local('grep -q "tag.local" /etc/hosts || sudo sh -c \'echo "127.0.0.1 tag.local" >> /etc/hosts\'')

# Run mkcert setup
local_resource(
  'setup-mkcert',
  cmd='./scripts/setup_mkcert.sh',
  labels=["scripts"],
  trigger_mode=TRIGGER_MODE_MANUAL,
  auto_init=False
)

# Set resources
resources = {
  'db': {'labels': ['database']},
  'migrations': {
    'auto_init': True,
    'trigger_mode': TRIGGER_MODE_MANUAL,
    'labels': ['database'],
    'resource_deps': ['go-compile', 'db']
    },
  'generate-models': {
    'auto_init': False,
    'trigger_mode': TRIGGER_MODE_MANUAL,
    'labels': ['scripts'],
    'resource_deps': ['db']
    },
  'proto-build': {
    'auto_init': False,
    'trigger_mode': TRIGGER_MODE_MANUAL,
    'labels': ['scripts'],
    },
  'api': {
    'labels': ['services'],
    'resource_deps': ['go-compile', 'db', 'migrations'],
    'trigger_mode': TRIGGER_MODE_AUTO,
  },
  'frontend': {
    'labels': ['frontend'],
    'resource_deps': ['api'],
    'trigger_mode': TRIGGER_MODE_AUTO,
  },
  'envoy': {
    'labels': ['networking'],
    'resource_deps': ['api', 'frontend'],
    'links': [
      link('https://tag.local', 'Thread Art Generator (HTTPS)'),
    ]
  },
  'minio': {'labels': ['database'], 'resource_deps': ['db']},
}

for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
