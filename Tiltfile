load('ext://restart_process', 'docker_build_with_restart')


local_resource(
  'go-compile',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/migrations cmd/migrations/main.go',
  labels=["scripts"],
  deps=['cmd/', 'db/', 'pkg/', 'threadGenerator/']
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
    sync('./build', '/app/build'),
    sync('./doc/swagger', '/doc/swagger'),
    restart_container ()
  ])

docker_build(
  'migrations-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-migrations',
  only=[
    './build/migrations',
    './pkg/db/migrations',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./pkg/db/migrations', '/migrations'),
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
  'web-proto-generator': {
    'auto_init': False,
    'trigger_mode': TRIGGER_MODE_MANUAL,
    'labels': ['scripts']
    },
  'api': {'labels': ['services'], 'resource_deps': ['go-compile', 'db', 'migrations']},
  'minio': {'labels': ['database'], 'resource_deps': ['db']},
  'web': {'labels': ['web'], 'resource_deps': ['api']},
}


for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
