# Thread Art Generator - Optimized Tiltfile for Local Development
# Load external extensions
load('ext://restart_process', 'docker_build_with_restart')

# ================================================
# CONSTANTS AND CONFIGURATIONS
# ================================================

# Define directories to watch for changes
CODE_DIRS = [
  'cmd',
  'core',
  'client',
  'threadGenerator'
]

# ================================================
# HELPER FUNCTIONS
# ================================================

def watch_templ_changes():
  # Watch templ files for changes to trigger rebuild
  watch_file('client/internal/templates')

  # Rebuild templ files when they change
  local_resource(
    'templ-generate',
    cmd='cd client && GOBIN=$(go env GOPATH)/bin $(go env GOPATH)/bin/templ generate ./internal/templates',
    labels=["build"],
    deps=['client/internal/templates/**/*.templ'],
    ignore=['client/internal/templates/**/*.templ.go'],
  )

def watch_tailwind_changes():
  # Watch and build Tailwind CSS initially
  local_resource(
    'tailwind-build',
    cmd='cd client && mkdir -p public/css && rm -rf node_modules package-lock.json && npm install && npx tailwindcss -i ./styles/input.css -o ./public/css/tailwind.css --minify',
    labels=["build"],
    deps=[
      'client/tailwind.config.js',
      'client/styles/input.css',
      'client/package.json',
    ],
  )

# ================================================
# BUILD CONFIGURATIONS
# ================================================

# Watch key generated files
watch_file('proto')
watch_file('core/pb')
watch_file('client/internal/pb')

# Run helper functions to set up watches
watch_templ_changes()
watch_tailwind_changes()

# Build and run proto when files change
docker_build(
  'proto-build-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-proto-build',
  only=[
    './proto',
  ],
  live_update=[
    sync('./proto', '/app/proto'),
  ]
)

# Compile Go binaries for all services in one build
local_resource(
  'go-compile',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go && ' +
      'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/migrations cmd/migrations/main.go && ' +
      'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/worker cmd/worker/main.go && ' +
      'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/frontend client/cmd/frontend/main.go && ' +
      'go build -o build/cli cmd/cli/main.go',
  labels=["build"],
  deps=CODE_DIRS,
  ignore=[
    'client/internal/templates/**/*.templ',  # Handled by templ-generate
    'proto',                                # Handled by proto-build
    'core/pb',                              # Generated files
    'client/internal/pb',                   # Generated files
  ],
  resource_deps=['templ-generate'],
)

# ================================================
# DOCKER IMAGE BUILDS
# ================================================

# API image build
docker_build(
  'api-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-api',
  only=['./build/api'],
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
  only=['./build/worker'],
  live_update=[
    sync('./build/worker', '/app/build/worker'),
    restart_container()
  ]
)

# Client (Go+HTMX Frontend) image build with improved live updates
docker_build(
  'frontend-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-frontend',
  only=[
    './client',
    './go.mod',
    './go.sum',
  ],
  live_update=[
    # Sync frontend source code changes
    sync('./client/internal', '/app/client/internal'),
    sync('./client/public', '/app/client/public'),
    sync('./client/cmd', '/app/client/cmd'),

    # Copy the compiled binary
    sync('./build/frontend', '/app/frontend'),

    # Restart container when binary changes
    restart_container()
  ]
)

# DB Models image build
docker_build(
  'db-models-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-db-models',
  only=[
    './go.mod',
    './go.sum',
  ]
)

# ================================================
# DOCKER COMPOSE CONFIGURATION
# ================================================

# Load docker-compose
docker_compose('./docker-compose.yml')

# ================================================
# MANUAL ACTIONS AND UTILITIES
# ================================================

# Setup script for local development
local_resource(
  'setup-local-dev',
  cmd='./scripts/local_setup.sh',
  labels=["setup"],
  trigger_mode=TRIGGER_MODE_MANUAL,
  auto_init=False
)

# Minio setup
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

# Add missing services
dc_resource(
  'migrations',
  labels=['database'],
  auto_init=True,
  trigger_mode=TRIGGER_MODE_MANUAL,
)

dc_resource(
  'generate-models',
  labels=['build'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
)

dc_resource(
  'proto-build',
  labels=['build'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
)

# Configure resources with consistent format
dc_resource(
  'db',
  labels=['database'],
  auto_init=True,
)

dc_resource(
  'rabbitmq',
  labels=['queue'],
  auto_init=True,
  links=[
    link('http://localhost:15672', 'RabbitMQ Management (guest/guest)'),
  ]
)

dc_resource(
  'worker',
  labels=['worker'],
  resource_deps=['go-compile', 'db', 'rabbitmq'],
  auto_init=True,
)

dc_resource(
  'api',
  labels=['application'],
  resource_deps=['go-compile', 'db', 'rabbitmq'],
  links=[
    link('http://localhost:9090', 'Connect API'),
    link('http://localhost:9090/health', 'API Health Check'),
  ]
)

dc_resource(
  'client',
  labels=['application'],
  resource_deps=['go-compile', 'tailwind-build'],
  links=[
    link('http://localhost:8080', 'Go+HTMX Frontend'),
    link('http://localhost:8080/health', 'Frontend Health Check'),
  ]
)

dc_resource(
  'envoy',
  labels=['proxy'],
  resource_deps=['api', 'client'],
  links=[
    link('https://front.tag.local', 'Go+HTMX Frontend (via Envoy)'),
    link('https://tag.local/health', 'API Health Check (via Envoy)'),
    link('http://localhost:9901', 'Envoy Admin'),
  ]
)

dc_resource(
  'minio',
  labels=['storage'],
  resource_deps=['db'],
  links=[
    link('http://localhost:9001', 'MinIO Console'),
  ]
)
