load('ext://restart_process', 'docker_build_with_restart')


local_resource(
  'go-compile',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/migrations cmd/migrations/main.go',
  labels=["scripts"],
  deps=['cmd/', 'core/', 'threadGenerator/', 'web/**/*.go'],
  ignore=['web/templates/**/*.go']  # Ignore templ-generated Go files
)

# Add templ compiler to generate Go code from .templ files
local_resource(
  'templ-compiler',
  cmd='go run github.com/a-h/templ/cmd/templ@latest generate',
  labels=["scripts"],
  deps=['web/templates/**/*.templ'],
  resource_deps=[],
  ignore=['web/templates/**/*.go'],  # Ignore generated Go files
  auto_init=True
)

# Run Next.js locally for development
local_resource(
  'nextjs-dev',
  serve_cmd='cd web && npm run dev',
  labels=["frontend"],
  links=[
    link('https://tag.local', 'Next.js Frontend')
  ],
  readiness_probe=probe(
    period_secs=2,
    http_get=http_get_action(
      port=3000,
      path='/',
      host='localhost'
    )
  ),
  allow_parallel=True
)

docker_build(
  'api-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-api',
  only=[
    './build/api',
    './doc/swagger',
  ],
  live_update=[
    sync('./build/api', '/app/build/api'),
    sync('./doc/swagger', '/doc/swagger'),
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
  cmd='./setup_mkcert.sh',
  labels=["scripts"],
  trigger_mode=TRIGGER_MODE_MANUAL,
  auto_init=False
)

# Build Traefik image with configuration
docker_build(
  'traefik-image',
  '.',
  dockerfile_contents='''
FROM traefik:v2.10
COPY ./certs/tag.local.crt /certs/tag.local.crt
COPY ./certs/tag.local.key /certs/tag.local.key
''',
  only=['./certs'],
  live_update=[
    sync('./certs', '/certs')
  ]
)

# Set the 'manual' resources to auto_init=False
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
  'go-proto-generator': {
    'auto_init': False,
    'trigger_mode': TRIGGER_MODE_MANUAL,
    'labels': ['scripts'],
    },
  'api': {
    'labels': ['services'],
    'resource_deps': ['go-compile', 'db', 'migrations'],
    'trigger_mode': TRIGGER_MODE_AUTO,  # Explicit trigger mode
  },
  'minio': {'labels': ['database'], 'resource_deps': ['db']},
  'traefik': {
    'labels': ['networking'],
    'resource_deps': ['api'],
    'links': [
      link('https://tag.local', 'Thread Art Generator (HTTPS)'),
      link('http://localhost:8080/dashboard/', 'Traefik Dashboard')
    ]
  }
}

for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
