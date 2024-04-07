# Load the docker compose configuration
docker_compose('docker-compose.yml')

# Set the 'manual' resources to auto_init=False
resources = {
  'migrations': {
    'auto_init': True,
    'trigger_mode': TRIGGER_MODE_MANUAL
    },
  'go-proto-generator': {
    'auto_init': False,
    },
  'web-proto-generator': {'auto_init': False},
  'generate-models': {'auto_init': False},
}

# docker_build(
#   # Image name - must match the image in the docker-compose file
#   'tilt.dev/express-redis-app',
#   # Docker context
#   '.',
#   live_update = [
#     # Sync local files into the container.
#     sync('.', '/var/www/app'),

#     # Re-run npm install whenever package.json changes.
#     run('npm i', trigger='package.json'),

#     # Restart the process to pick up the changed files.
#     restart_container()
#   ])


dc_resource('api', labels=["services"])
#dc_resource('worker', labels=["worker"])

dc_resource('db', labels=["database"])
dc_resource('adminer', labels=["database"])
dc_resource('migrations', labels=["database", "scripts"])

dc_resource('go-proto-generator', labels=["scripts"])
dc_resource('generate-models', labels=["scripts"])
dc_resource('web-proto-generator', labels=["scripts"])

for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
