load('ext://restart_process', 'docker_build_with_restart')


local_resource(
  'go-compile',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/migrations cmd/migrations/main.go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/web cmd/web/main.go',
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

docker_build(
  'web-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-web',
  only=[
    './build/web',
    './web/static',
  ],
  live_update=[
    sync('./build/web', '/app/build/web'),
    sync('./web/static', '/app/web/static'),
    restart_container()
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
  'web': {
    'labels': ['services'],
    'resource_deps': ['go-compile'],  # Removed templ-compiler dependency
    'trigger_mode': TRIGGER_MODE_AUTO,
  },
  'minio': {'labels': ['database'], 'resource_deps': ['db']},
}


for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
