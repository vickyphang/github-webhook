# Github Webhook 📬
[![GitHub license](https://img.shields.io/github/license/vickyphang/archimedes)](https://github.com/vickyphang/github-webhook/blob/main/LICENSE)
![GitHub stars](https://img.shields.io/github/stars/vickyphang/archimedes)

<p align="center"> <img src="images/logo.png"> </p>

The purpose of this project is to setup `Github` to `push notifications` to a `monitoring server` whenever a developer pushes changes to a repository using `Github's webhooks`. Webhooks allow you to send `HTTP POST` requests to a specified URL in response to certain events, such as code pushes.

## Deployment: quick start
### 1. Prepare RabbitMQ
```bash
docker run -d -p 15672:15672 -p 5672:5672 --name rabbitmq rabbitmq:3-management
```

### 2. Prepare Monitoring Server
This server first verifies the payload's signature to ensure the request is from GitHub. Then handles incoming webhook requests. It reads the request body, verifies the signature, parses the JSON payload, and sends it to RabbitMQ.

- Build image
```bash
docker build -t github-webhook github-webhook/
```

- Run container
```bash
docker run -d -p 5000:5000 -e WEBHOOK_SECRET='super-secret' \
	-e AMQP_URL='amqp://guest:guest@rabbitmq_ip:5672/' \
	-e QUEUE_NAME='github_webhook' \
	-e LISTEN_PORT=':5000' \
	--name github-webhook github-webhook
```

### 3. Prepare Worker
This server connects to RabbitMQ and listens for messages from the specified queue. Then parses the JSON payload and processes the push event data by printing commit messages and author names.

- Build image
```bash
docker build -t worker worker/
```

- Run container
```bash
docker run -d -e QUEUE_NAME='github_webhook' \
	-e AMQP_URL='amqp://guest:guest@rabbitmq_ip:5672/' \
	--name worker worker
```

### 4. Create a Webhook in GitHub
- Go to your repository on GitHub.
- Navigate to `Settings` > `Webhooks`.
- Click the `Add webhook` button.
- In the `Payload URL` field, enter the URL of the `endpoint` on your `monitoring server` where you want to receive the notifications.
- Set the `Content type` to `application/json`.
- Optionally, enter a `Secret`. This secret will be used to ensure that the incoming webhook is indeed from GitHub.
- In the Which events would you like to trigger this webhook? section, select `Just the push event`.
- Click the `Add webhook` button to save the webhook configuration.

<p align="center"> <img src="images/edit-webhook.png"> </p>

<p align="center"> <img src="images/add-success.png"> </p>

### 5. Verify
Create any changes in the repository and then run `docker logs --follow worker` to check the logs. It should look like this:

<p align="center"> <img src="images/push-notification.png"> </p>