import json
import base64
import os
from google.cloud import run_v2
from google.cloud import sql_v1
from google.cloud import redis_v1
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def budget_enforcer(cloud_event):
    """
    Cloud Function triggered by Pub/Sub messages from Cloud Billing budgets.
    Shuts down services when budget threshold is exceeded.
    """
    
    try:
        # Decode the Pub/Sub message
        pubsub_message = base64.b64decode(cloud_event.data['message']['data']).decode('utf-8')
        budget_data = json.loads(pubsub_message)
        
        logger.info(f"Received budget alert: {budget_data}")
        
        # Check if this is a budget alert
        if 'costAmount' not in budget_data or 'budgetAmount' not in budget_data:
            logger.info("Not a valid budget alert message")
            return
        
        cost_amount = float(budget_data['costAmount'])
        budget_amount = float(budget_data['budgetAmount'])
        threshold_percent = cost_amount / budget_amount
        
        logger.info(f"Current spend: $${cost_amount:.2f}, Budget: $${budget_amount:.2f}, Threshold: {threshold_percent:.2%}")
        
        project_id = os.environ.get('PROJECT_ID')
        environment = os.environ.get('ENVIRONMENT')
        
        # If we've exceeded 90% of budget, shut down non-essential services
        if threshold_percent >= 0.9:
            logger.warning(f"Budget threshold exceeded ({threshold_percent:.2%}). Shutting down services.")
            
            # Shut down Cloud Run services
            shutdown_cloud_run_services(project_id, environment)
            
            # If we've exceeded 100%, shut down everything including database
            if threshold_percent >= 1.0:
                logger.critical(f"Budget completely exceeded ({threshold_percent:.2%}). Shutting down all services including database.")
                shutdown_cloud_sql_instances(project_id, environment)
                shutdown_redis_instances(project_id, environment)
        
        return {'status': 'success', 'threshold': threshold_percent}
        
    except Exception as e:
        logger.error(f"Error processing budget alert: {str(e)}")
        raise

def shutdown_cloud_run_services(project_id, environment):
    """Shut down Cloud Run services"""
    try:
        client = run_v2.ServicesClient()
        parent = f"projects/{project_id}/locations/us-central1"
        
        # List all services
        services = client.list_services(parent=parent)
        
        for service in services:
            service_name = service.name
            if environment in service_name:
                logger.info(f"Shutting down Cloud Run service: {service_name}")
                
                # Update service to scale to zero
                service.spec.template.spec.scaling.max_instance_count = 0
                
                update_mask = {"paths": ["spec.template.spec.scaling.max_instance_count"]}
                operation = client.update_service(
                    service=service,
                    update_mask=update_mask
                )
                
                logger.info(f"Cloud Run service {service_name} scaled to zero")
                
    except Exception as e:
        logger.error(f"Error shutting down Cloud Run services: {str(e)}")

def shutdown_cloud_sql_instances(project_id, environment):
    """Shut down Cloud SQL instances"""
    try:
        client = sql_v1.SqlInstancesServiceClient()
        
        # List all SQL instances
        request = sql_v1.SqlInstancesListRequest(project=project_id)
        instances = client.list(request=request)
        
        for instance in instances.items:
            if environment in instance.name:
                logger.info(f"Shutting down Cloud SQL instance: {instance.name}")
                
                # Stop the instance
                request = sql_v1.SqlInstancesPatchRequest(
                    project=project_id,
                    instance=instance.name,
                    body=sql_v1.DatabaseInstance(
                        settings=sql_v1.Settings(
                            activation_policy="NEVER"
                        )
                    )
                )
                
                operation = client.patch(request=request)
                logger.info(f"Cloud SQL instance {instance.name} shutdown initiated")
                
    except Exception as e:
        logger.error(f"Error shutting down Cloud SQL instances: {str(e)}")

def shutdown_redis_instances(project_id, environment):
    """Shut down Redis instances"""
    try:
        client = redis_v1.CloudRedisClient()
        parent = f"projects/{project_id}/locations/us-central1"
        
        # List all Redis instances
        instances = client.list_instances(parent=parent)
        
        for instance in instances:
            if environment in instance.name:
                logger.info(f"Deleting Redis instance: {instance.name}")
                
                # Delete the instance (Redis doesn't have a stop operation)
                operation = client.delete_instance(name=instance.name)
                logger.info(f"Redis instance {instance.name} deletion initiated")
                
    except Exception as e:
        logger.error(f"Error shutting down Redis instances: {str(e)}")