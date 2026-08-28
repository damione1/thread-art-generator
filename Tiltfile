# Thread Art Generator - Optimized Tiltfile for Local Development
# Load external extensions
load('ext://restart_process', 'docker_build_with_restart')

# ================================================
# CONSTANTS AND CONFIGURATIONS
# ================================================

# Define directories to watch for changes
CODE_DIRS = {
  'api': ['cmd/api', 'core'],
  'worker': ['cmd/worker', 'core', 'threadGenerator'],
  'frontend': ['client', 'core']
}

# Environment configuration helper
def load_env_vars():
    """Load environment variables from .env file"""
    ENV_FILE = '.env'
    if os.path.exists(ENV_FILE):
        env_content = str(read_file(ENV_FILE)).strip()
        if env_content:
            env_lines = env_content.split('\n')
            env_dict = {}
            for line in env_lines:
                if line.strip() and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    env_dict[key.strip()] = value.strip()
            return env_dict
    return {}

# Load environment variables
ENV_VARS = load_env_vars()

# ================================================
# HELPER FUNCTIONS
# ================================================

def watch_templ_changes():
  # Watch templ files for changes to trigger rebuild
  local_resource(
    'templ-generate',
    cmd='make generate-templ',
    labels=["build"],
    deps=[
      'client/internal/templates/**/*.templ',
      'client/internal/components/**/*.templ',
      'Makefile',
    ],
    ignore=[
      'client/internal/templates/**/*_templ.go',  # Ignore generated Go files
      'client/internal/components/**/*_templ.go',  # Ignore generated Go files
      '**/*.templ.go',  # Catch any other generated templ files
    ],
    trigger_mode=TRIGGER_MODE_AUTO,
  )

def watch_frontend_assets():
  # Build CSS and JS assets using npm
  local_resource(
    'frontend-assets-build',
    cmd='cd client && npm install && npm run build',
    labels=["build"],
    deps=[
      'client/tailwind.config.js',
      'client/styles/input.css',
      'client/package.json',
      'client/src/**/*.js',
      'client/src/**/*.ts',
      'client/webpack.config.js',
    ],
    trigger_mode=TRIGGER_MODE_AUTO,
  )

# ================================================
# BUILD CONFIGURATIONS
# ================================================

# Set up file watches for key directories
# Use watch_file for entire directories to track all files within them
watch_file('proto')
watch_file('core/pb')
watch_file('client/internal/pb')

# Run helper functions to set up watches
watch_templ_changes()
watch_frontend_assets()

# Protocol buffer generation using make
local_resource(
  'proto-generate',
  cmd='make proto',
  labels=["build"],
  deps=['proto/**/*.proto'],
  trigger_mode=TRIGGER_MODE_AUTO,
)

# Separate build resources for each service
local_resource(
  'api-build',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go',
  labels=["build"],
  deps=CODE_DIRS['api'],
  resource_deps=['proto-generate'],
  ignore=[
    'proto/**',
    'core/pb/**',
    'build/**',
  ],
  trigger_mode=TRIGGER_MODE_AUTO,
)

local_resource(
  'worker-build',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/worker cmd/worker/main.go',
  labels=["build"],
  deps=CODE_DIRS['worker'],
  resource_deps=['proto-generate'],
  ignore=[
    'proto/**',
    'core/pb/**',
    'build/**',
  ],
  trigger_mode=TRIGGER_MODE_AUTO,
)

local_resource(
  'frontend-build',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/frontend client/cmd/frontend/main.go',
  labels=["build"],
  deps=CODE_DIRS['frontend'],
  resource_deps=['proto-generate', 'templ-generate', 'frontend-assets-build'],
  ignore=[
    'client/internal/templates/**/*.templ',
    'client/src/**/*.js',
    'client/styles/**',
    'client/public/**',
    'proto/**',
    'core/pb/**',
    'client/internal/pb/**',
    'build/**',
  ],
  trigger_mode=TRIGGER_MODE_AUTO,
)


# ================================================
# DOCKER IMAGE BUILDS
# ================================================

# API image build with optimized live updates
docker_build(
  'api-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-api',
  only=['./build/api'],
  live_update=[
    # Minimal run step required by Tilt
    run('echo "Binary updated"', trigger=['./build/api']),
    restart_container()
  ]
)

# Migration and model generation now handled via local tools

# Worker image build with optimized live updates
docker_build(
  'worker-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-worker',
  only=['./build/worker'],
  live_update=[
    # Minimal run step required by Tilt
    run('echo "Binary updated"', trigger=['./build/worker']),
    restart_container()
  ]
)

# Client (Go+HTMX Frontend) image build with optimized live updates
docker_build(
  'frontend-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-frontend',
  only=[
    './build/frontend',
    './client/public',
  ],
  live_update=[
    # Sync public assets without restart
    sync('./client/public', '/app/client/public'),

    # For binary updates, restart container (no sync needed due to volume mount)
    restart_container()
  ]
)

# Note: database models now handled via Make targets instead of Docker services

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



# ================================================
# SERVICE CONFIGURATION
# ================================================

# Database operations using Make targets
local_resource(
  'run-migrations',
  cmd='make run-migrations',
  labels=['database'],
  resource_deps=['db'],
  deps=['core/db/migrations'],
  auto_init=True,
  trigger_mode=TRIGGER_MODE_AUTO,
)

local_resource(
  'generate-db-models',
  cmd='make generate-models',
  labels=['database'],
  deps=['core/db/migrations'],
  resource_deps=['db'],
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
  'worker',
  labels=['worker'],
  resource_deps=['worker-build', 'run-migrations', 'minio'],
  auto_init=True,
)

dc_resource(
  'api',
  labels=['application'],
  resource_deps=['api-build', 'db', 'minio'],
  auto_init=True,
  links=[
    link('http://localhost:9090', 'Connect API'),
    link('http://localhost:9090/health', 'API Health Check'),
  ]
)

dc_resource(
  'client',
  labels=['application'],
  resource_deps=['frontend-build', 'frontend-assets-build', 'api'],
  auto_init=True,
  links=[
    link('http://localhost:8080', 'Go+HTMX Frontend'),
    link('http://localhost:8080/health', 'Frontend Health Check'),
  ]
)

dc_resource(
  'minio',
  labels=['storage'],
  links=[
    link('http://localhost:9001', 'MinIO Console'),
  ]
)

dc_resource(
  'minio-init',
  labels=['storage'],
  resource_deps=['minio'],
  auto_init=True,
)

dc_resource(
  'redis',
  labels=['cache'],
  auto_init=True,
)
