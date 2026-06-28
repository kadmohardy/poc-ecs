
############################################
# APPLICATION LOAD BALANCER
############################################

resource "aws_lb" "this" {
  name = "${var.project_name}-alb"

  internal           = false
  load_balancer_type = "application"

  security_groups = [
    aws_security_group.alb.id
  ]

  subnets = [
    aws_subnet.public_a.id,
    aws_subnet.public_b.id
  ]
}

############################################
# TARGET GROUP
############################################
resource "aws_lb_target_group" "tg2" {
  name        = "${var.project_name}-tg2"
  port        = 8080
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.this.id

  lifecycle {
    create_before_destroy = true
  }

  health_check {
    enabled             = true
    path                = "/"
    matcher             = "200"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 30
    timeout             = 5
  }
}

############################################
# LISTENER
############################################

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.tg2.arn
  }

  depends_on = [
    aws_lb_target_group.tg2
  ]
}
