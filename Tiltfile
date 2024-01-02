# Load the docker compose configuration
docker_compose('docker-compose.yml')

# Set the 'manual' resources to auto_init=False
resources = {
  'proto-generator': {'auto_init': False},
}

for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
