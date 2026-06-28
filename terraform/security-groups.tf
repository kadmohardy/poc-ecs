############################################
# SECURITY GROUP - ECS
############################################

resource "aws_security_group" "ecs" {
  name   = "${var.project_name}-ecs-sg"
  vpc_id = aws_vpc.this.id

  ingress {
    from_port = 8080
    to_port   = 8080
    protocol  = "tcp"

    security_groups = [
      aws_security_group.alb.id
    ]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0"
    ]
  }
}

############################################
# SECURITY GROUP - ALB
############################################

resource "aws_security_group" "alb" {
  name   = "${var.project_name}-alb-sg"
  vpc_id = aws_vpc.this.id

  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"

    cidr_blocks = [
      "0.0.0.0/0"
    ]
  }

  ingress {
    from_port = 3000
    to_port   = 3000
    protocol  = "tcp"

    cidr_blocks = [
      "0.0.0.0/0"
    ]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = [
      "0.0.0.0/0"
    ]
  }
}

############################################
# SECURITY GROUP - GRAFANA
############################################

resource "aws_security_group" "grafana" {
  name   = "${var.project_name}-grafana-sg"
  vpc_id = aws_vpc.this.id

  ingress {
    from_port = 3000
    to_port   = 3000
    protocol  = "tcp"

    security_groups = [
      aws_security_group.alb.id
    ]
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"

    cidr_blocks = ["0.0.0.0/0"]
  }
}

############################################
# SECURITY GROUP - ADOT
############################################

resource "aws_security_group" "adot" {
  name   = "${var.project_name}-adot-sg"
  vpc_id = aws_vpc.this.id

  ingress {
    from_port = 4317
    to_port   = 4317
    protocol  = "tcp"

    security_groups = [
      aws_security_group.ecs.id
    ]
  }

  ingress {
    from_port = 4318
    to_port   = 4318
    protocol  = "tcp"

    security_groups = [
      aws_security_group.ecs.id
    ]
  }

  ingress {
    from_port = 13133
    to_port   = 13133
    protocol  = "tcp"
    security_groups = [
      aws_security_group.ecs.id
    ]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

