# Load the docker compose configuration
docker_compose('docker-compose.yml')

# Set the 'manual' resources to auto_init=False
resources = {
  # 'go-proto-generator': {'auto_init': False},
  # 'web-proto-generator': {'auto_init': False},
  # 'generate-models': {'auto_init': False},
}

for resource_name, resource_config in resources.items():
    dc_resource(
        resource_name,
        **resource_config,
    )
