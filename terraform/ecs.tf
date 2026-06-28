############################################
# IAM ROLE - ECS EXECUTION ROLE
############################################

resource "aws_iam_role" "ecs_execution_role" {
  name = "${var.project_name}-ecs-execution-role"

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

############################################
# ECS EXECUTION POLICY
############################################

resource "aws_iam_role_policy_attachment" "ecs_execution_policy" {
  role = aws_iam_role.ecs_execution_role.name

  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

############################################
# IAM ROLE - ECS TASK ROLE
############################################

resource "aws_iam_role" "ecs_task_role" {
  name = "${var.project_name}-ecs-task-role"

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

############################################
# X-RAY POLICY
############################################

resource "aws_iam_role_policy_attachment" "xray" {
  role = aws_iam_role.ecs_task_role.name

  policy_arn = "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
}

############################################
# CLOUDWATCH METRICS POLICY
############################################

resource "aws_iam_policy" "adot_metrics" {
  name = "${var.project_name}-adot-metrics-policy"

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"

        Action = [
          "cloudwatch:PutMetricData"
        ]

        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "metrics_attach" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.adot_metrics.arn
}

############################################
# ECS CLUSTER
############################################

resource "aws_ecs_cluster" "this" {
  name = "${var.project_name}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

############################################
# CLOUDWATCH LOG GROUP
############################################

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${var.project_name}"
  retention_in_days = 7
}


############################################
# ECS TASK DEFINITION
############################################

resource "aws_ecs_task_definition" "app" {
  family = "${var.project_name}-task"

  requires_compatibilities = [
    "FARGATE"
  ]

  network_mode       = "awsvpc"
  cpu                = 512
  memory             = 1024
  execution_role_arn = aws_iam_role.ecs_execution_role.arn
  task_role_arn      = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "app"
      image     = "755209495094.dkr.ecr.us-east-1.amazonaws.com/poc-ecs:latest"
      essential = true

      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "OTEL_EXPORTER_OTLP_ENDPOINT"
          value = "adot.internal:4317"
        },
        {
          name  = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
          value = "adot.internal:4317"
        },
        {
          name  = "OTEL_SERVICE_NAME"
          value = "ecs-demo"
        },
        {
          name  = "OTEL_EXPORTER_OTLP_PROTOCOL"
          value = "grpc"
        },
        {
          name  = "OTEL_EXPORTER_OTLP_INSECURE"
          value = "true"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"

        options = {
          awslogs-group         = aws_cloudwatch_log_group.ecs.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "app"
        }
      }
    }
  ])
}

############################################
# ECS SERVICE
############################################

resource "aws_ecs_service" "app" {
  name            = "${var.project_name}-service"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets = [
      aws_subnet.public_a.id,
      aws_subnet.public_b.id
    ]

    security_groups = [
      aws_security_group.ecs.id
    ]

    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.tg2.arn
    container_name   = "app"
    container_port   = 8080
  }

  depends_on = [
    aws_lb_listener.http
  ]
}


############################################
# ECS GRAFANA
############################################

resource "aws_ecs_task_definition" "grafana" {
  family                   = "grafana"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  network_mode             = "awsvpc"
  execution_role_arn       = aws_iam_role.ecs_execution_role.arn
  task_role_arn            = aws_iam_role.grafana_task.arn

  container_definitions = jsonencode([
    {
      name  = "grafana"
      image = "grafana/grafana:12.2.0"

      portMappings = [
        {
          containerPort = 3000
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "GF_SECURITY_ADMIN_USER"
          value = "admin"
        },
        {
          name  = "GF_SECURITY_ADMIN_PASSWORD"
          value = "admin123"
        },
        {
          name  = "GF_PLUGINS_PREINSTALL"
          value = "grafana-x-ray-datasource"
        },
        {
          name  = "GF_INSTALL_PLUGINS"
          value = "grafana-x-ray-datasource"
        },
        {
          name  = "GF_LOG_LEVEL"
          value = "debug"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"

        options = {
          awslogs-group         = aws_cloudwatch_log_group.ecs.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "grafana"
        }
      }
    }
  ])
}

