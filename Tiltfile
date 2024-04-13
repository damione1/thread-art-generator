load('ext://restart_process', 'docker_build_with_restart')


local_resource(
  'api-compile',
  cmd='CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api cmd/api/main.go',
  labels=["services"],
  deps=['cmd/', 'db/', 'pkg/', 'threadGenerator/']
)

docker_build(
  'api-image',
  '.',
  dockerfile='Infra/Dockerfiles/Dockerfile-api',
  only=[
    './build',
  ],
  live_update=[
    sync('./build', '/app/build'),
    restart_container ()
  ])

# Load the docker compose configuration
docker_compose('docker-compose.yml')

# Set the 'manual' resources to auto_init=False
resources = {
  'db': {'labels': ['database']},
  'migrations': {
    'auto_init': True,
    'trigger_mode': TRIGGER_MODE_MANUAL,
    'labels': ['database', 'scripts'],
    'resource_deps': ['db']
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
  'api': {'labels': ['services'], 'resource_deps': ['api-compile', 'db']},
  'adminer': {'labels': ['database'], 'resource_deps': ['db']},
}


for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
