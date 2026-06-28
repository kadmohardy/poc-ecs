############################################
# OUTPUTS
############################################

output "alb_url" {
  value = "http://${aws_lb.this.dns_name}"
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.this.name
}

output "ecs_service_name" {
  value = aws_ecs_service.app.name
}

############################################
# AMAZON MANAGED PROMETHEUS
############################################

resource "aws_prometheus_workspace" "this" {
  alias = "${var.project_name}-amp"
}

############################################
# AMP POLICY
############################################

resource "aws_iam_policy" "amp" {
  name = "${var.project_name}-amp-policy"

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"

        Action = [
          "aps:RemoteWrite",
          "aps:GetSeries",
          "aps:GetLabels",
          "aps:GetMetricMetadata"
        ]

        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "amp_attach" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.amp.arn
}


resource "aws_iam_role" "grafana_task" {
  name = "${var.project_name}-grafana-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"

        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }

        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_lb_target_group" "grafana" {
  name        = "${var.project_name}-grafana"
  port        = 3000
  protocol    = "HTTP"
  target_type = "ip"

  vpc_id = aws_vpc.this.id

  health_check {
    path = "/api/health"

    matcher = "200"
  }
}

resource "aws_ecs_service" "grafana" {
  name            = "grafana-service"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.grafana.arn

  desired_count = 1
  launch_type   = "FARGATE"

  network_configuration {
    subnets = [
      aws_subnet.public_a.id,
      aws_subnet.public_b.id
    ]

    security_groups = [
      aws_security_group.grafana.id
    ]

    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.grafana.arn

    container_name = "grafana"
    container_port = 3000
  }

  depends_on = [
    aws_lb_listener.http
  ]
}

resource "aws_lb_listener" "grafana" {
  load_balancer_arn = aws_lb.this.arn

  port     = 3000
  protocol = "HTTP"

  default_action {
    type = "forward"

    target_group_arn = aws_lb_target_group.grafana.arn
  }
}

resource "aws_iam_policy" "grafana_amp" {
  name = "${var.project_name}-grafana-amp"

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"

        Action = [
          "aps:QueryMetrics",
          "aps:GetSeries",
          "aps:GetLabels",
          "aps:GetMetricMetadata",
          "aps:QueryRange",
          "aps:ListWorkspaces",
          "aps:DescribeWorkspace"
        ]

        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_policy" "grafana_xray" {
  name = "${var.project_name}-grafana-xray"

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"

        Action = [
          "xray:BatchGetTraces",
          "xray:GetTraceSummaries",
          "xray:GetTraceGraph",
          "xray:GetGroups",
          "xray:GetGroup",
          "xray:GetTimeSeriesServiceStatistics",
          "xray:GetInsightSummaries",
          "xray:GetInsight",
          "ec2:DescribeRegions"
        ]

        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "grafana_xray_attach" {
  role       = aws_iam_role.grafana_task.name
  policy_arn = aws_iam_policy.grafana_xray.arn
}

resource "aws_iam_role_policy_attachment" "grafana_amp_attach" {
  role       = aws_iam_role.grafana_task.name
  policy_arn = aws_iam_policy.grafana_amp.arn
}

### SETUP ADOT ###
resource "aws_ssm_parameter" "adot_config" {
  name = "/${var.project_name}/adot-config"

  type = "String"

  value = <<EOF
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  memory_limiter: 
    check_interval: 5s 
    limit_mib: 512 
    spike_limit_mib: 128
  batch:
    timeout: 10s
    send_batch_size: 1024

extensions:
  health_check:
    endpoint: 0.0.0.0:13133
  sigv4auth: 
    region: ${var.aws_region}

exporters:
  awsxray:
    region: us-east-1

  prometheusremotewrite:
    endpoint: "https://aps-workspaces.${var.aws_region}.amazonaws.com/workspaces/${aws_prometheus_workspace.this.id}/api/v1/remote_write"
    auth: 
      authenticator: sigv4auth

service:
  telemetry:
    logs:
      level: "debug"

  extensions: 
    - health_check
    - sigv4auth

  pipelines:
    traces:
      receivers: 
        - otlp
     
      processors: 
        - memory_limiter 
        - batch

      exporters: 
        - awsxray

    metrics:
      receivers: 
        - otlp
      processors: 
        - memory_limiter 
        - batch 
      exporters: 
        - prometheusremotewrite
EOF
}

resource "aws_ecs_task_definition" "adot" {
  family                   = "adot"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 512
  memory                   = 1024
  execution_role_arn       = aws_iam_role.ecs_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "adot"
      essential = true

      image = "public.ecr.aws/aws-observability/aws-otel-collector:v0.48.0"

      command = [
        "--config=env:ADOT_CONFIG"
      ]

      environment = [
        {
          name  = "ADOT_CONFIG"
          value = aws_ssm_parameter.adot_config.value
        }
      ]

      portMappings = [
        {
          containerPort = 4317
          protocol      = "tcp"
        },
        {
          containerPort = 4318
          protocol      = "tcp"
        },
        {
          containerPort = 13133
          protocol      = "tcp"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"

        options = {
          awslogs-group         = aws_cloudwatch_log_group.ecs.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "adot"
        }
      }
    }
  ])
}

resource "aws_ecs_service" "adot" {
  name            = "adot-service"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.adot.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets = [
      aws_subnet.public_a.id,
      aws_subnet.public_b.id
    ]

    security_groups = [
      aws_security_group.adot.id
    ]

    assign_public_ip = true
  }

  service_registries {
    registry_arn = aws_service_discovery_service.adot.arn
  }
}

resource "aws_service_discovery_private_dns_namespace" "this" {
  name = "internal"
  vpc  = aws_vpc.this.id
}

resource "aws_service_discovery_service" "adot" {
  name = "adot"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.this.id

    dns_records {
      type = "A"
      ttl  = 10
    }
  }
}
