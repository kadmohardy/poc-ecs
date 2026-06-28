docker buildx build --platform linux/amd64 -t poc-ecs:latest .
docker tag poc-ecs:latest 755209495094.dkr.ecr.us-east-1.amazonaws.com/poc-ecs:latest
docker push 755209495094.dkr.ecr.us-east-1.amazonaws.com/poc-ecs:latest

aws ecr get-login-password \
  --region us-east-1 \
| docker login \
  --username AWS \
  --password-stdin 755209495094.dkr.ecr.us-east-1.amazonaws.com