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

# Add NextJS dependencies and build
local_resource(
  'nextjs-deps',
  cmd='cd web && npm install',
  labels=["scripts"],
  deps=['web/package.json'],
  resource_deps=[],
  auto_init=True
)

# Run Next.js locally for development
local_resource(
  'nextjs-dev',
  serve_cmd='cd web && npm run dev',
  labels=["frontend"],
  resource_deps=['nextjs-deps'],
  serve_env={
    'NEXT_PUBLIC_API_URL': 'http://localhost:9090',
    'NEXT_PUBLIC_AUTH0_DOMAIN': '${AUTH0_DOMAIN}',
    'NEXT_PUBLIC_AUTH0_CLIENT_ID': '${AUTH0_CLIENT_ID}',
    'NEXT_PUBLIC_AUTH0_AUDIENCE': '${AUTH0_AUDIENCE}',
    'NEXT_PUBLIC_FRONTEND_URL': 'http://localhost:3000',
    'AUTH0_DOMAIN': '${AUTH0_DOMAIN}',
    'AUTH0_CLIENT_ID': '${AUTH0_CLIENT_ID}',
    'AUTH0_CLIENT_SECRET': '${AUTH0_CLIENT_SECRET}',
    'AUTH0_SECRET': '${AUTH0_SECRET}',
    'APP_BASE_URL': 'http://localhost:3000'
  },
  links=[
    link('http://localhost:3000', 'Next.js Frontend')
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
}


for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
