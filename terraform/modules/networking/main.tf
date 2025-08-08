# VPC Network
resource "google_compute_network" "vpc" {
  name                    = "thread-art-vpc-${var.environment}"
  auto_create_subnetworks = false
  routing_mode           = "REGIONAL"
  project                = var.project_id
  description            = "VPC network for Thread Art Generator ${var.environment}"
}

# Private subnet for services
resource "google_compute_subnetwork" "private_subnet" {
  name          = "thread-art-private-${var.environment}"
  ip_cidr_range = var.private_subnet_cidr
  region        = var.region
  network       = google_compute_network.vpc.id
  project       = var.project_id
  description   = "Private subnet for Thread Art services"

  # Enable private Google access
  private_ip_google_access = true

  # Secondary IP ranges for additional services if needed
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = var.pods_cidr_range
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = var.services_cidr_range
  }
}

# Cloud Router for NAT
resource "google_compute_router" "router" {
  name    = "thread-art-router-${var.environment}"
  region  = var.region
  network = google_compute_network.vpc.id
  project = var.project_id

  bgp {
    asn = 64514
  }
}

# NAT Gateway for outbound internet access
resource "google_compute_router_nat" "nat" {
  name                               = "thread-art-nat-${var.environment}"
  router                            = google_compute_router.router.name
  region                            = var.region
  project                           = var.project_id
  nat_ip_allocate_option            = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

# VPC connector for Cloud Run
resource "google_vpc_access_connector" "connector" {
  name          = "thread-art-conn-${var.environment}"
  region        = var.region
  project       = var.project_id
  network       = google_compute_network.vpc.name
  ip_cidr_range = var.connector_cidr_range
  
  min_instances = 2
  max_instances = 3
  
  machine_type = "e2-micro"
}

# Firewall rule to allow internal communication
resource "google_compute_firewall" "allow_internal" {
  name    = "thread-art-allow-internal-${var.environment}"
  network = google_compute_network.vpc.id
  project = var.project_id

  allow {
    protocol = "tcp"
    ports    = ["1-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["1-65535"]
  }

  allow {
    protocol = "icmp"
  }

  source_ranges = [
    var.private_subnet_cidr,
    var.pods_cidr_range,
    var.services_cidr_range,
    var.connector_cidr_range
  ]

  target_tags = ["thread-art-internal"]
}

# Firewall rule to allow health checks
resource "google_compute_firewall" "allow_health_checks" {
  name    = "thread-art-allow-health-checks-${var.environment}"
  network = google_compute_network.vpc.id
  project = var.project_id

  allow {
    protocol = "tcp"
    ports    = ["8080", "9090"]
  }

  source_ranges = [
    "130.211.0.0/22",  # Google Load Balancer health check ranges
    "35.191.0.0/16",
  ]

  target_tags = ["thread-art-service"]
}

# Firewall rule to allow Cloud SQL proxy
resource "google_compute_firewall" "allow_cloud_sql_proxy" {
  name    = "thread-art-allow-cloud-sql-proxy-${var.environment}"
  network = google_compute_network.vpc.id
  project = var.project_id

  allow {
    protocol = "tcp"
    ports    = ["5432"]
  }

  source_ranges = [var.private_subnet_cidr]
  target_tags   = ["thread-art-database"]
}

# Private Service Connect configuration for Cloud SQL
resource "google_compute_global_address" "private_ip_address" {
  name          = "thread-art-private-ip-${var.environment}"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc.id
  project       = var.project_id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# Network Security Policy (basic DDoS protection)
resource "google_compute_security_policy" "default" {
  name        = "thread-art-security-policy-${var.environment}"
  description = "Security policy for Thread Art Generator"
  project     = var.project_id

  # Default rule to allow traffic
  rule {
    action   = "allow"
    priority = "2147483647"
    
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
  }

  # Rate limiting rule
  rule {
    action   = "rate_based_ban"
    priority = "1000"
    
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }

    rate_limit_options {
      conform_action = "allow"
      exceed_action  = "deny(429)"
      enforce_on_key = "IP"
      
      rate_limit_threshold {
        count        = 100
        interval_sec = 60
      }
      
      ban_duration_sec = 300
    }
  }
}