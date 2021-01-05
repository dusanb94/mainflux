# SMTP Notifications

SMTP Notifications provides a service for sending notification emails.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                         | Description                                               | Default                |
| -------------------------------- | --------------------------------------------------------- | ---------------------- |
| MF_NATS_URL                      | NATS instance URL                                         | nats://localhost:4222  |
| MF_SMTP_NOTIFICATIONS_LOG_LEVEL    | Log level for Cassandra writer (debug, info, warn, error) | error                  |
| MF_SMTP_NOTIFICATIONS_PORT         | Service HTTP port                                         | 8180                   |
| MF_SMTP_NOTIFICATIONS_CONFIG_PATH  | Configuration file path with NATS subjects list           | /config.toml           |

## Deployment

```yaml
  version: "3.7"
  cassandra-writer:
    image: mainflux/cassandra-writer:[version]
    container_name: [instance name]
    expose:
      - [Service HTTP port]
    restart: on-failure
    environment:
      MF_NATS_URL: [NATS instance URL]
      MF_SMTP_NOTIFICATIONS_LOG_LEVEL: [Cassandra writer log level]
      MF_SMTP_NOTIFICATIONS_PORT: [Service HTTP port]
      MF_SMTP_NOTIFICATIONS_CONFIG_PATH: [Configuration file path with NATS subjects list]
    ports:
      - [host machine port]:[configured HTTP port]
    volume:
      - ./config.toml:/config.toml
```

To start the service, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/mainflux/mainflux

cd mainflux

# compile the cassandra writer
make cassandra-writer

# copy binary to bin
make install

# Set the environment variables and run the service
MF_NATS_URL=[NATS instance URL] \
MF_SMTP_NOTIFICATIONS_LOG_LEVEL=[Cassandra writer log level] \
MF_SMTP_NOTIFICATIONS_PORT=[Service HTTP port] \
MF_SMTP_NOTIFICATIONS_CONFIG_PATH=[Configuration file path with NATS subjects list] \
$GOBIN/mainflux-cassandra-writer
```

### Using docker-compose

This service can be deployed using docker containers. Docker compose file is
available in `<project_root>/docker/addons/smtp-notifications/docker-compose.yml`.
In order to run all Mainflux core services, as well as mentioned optional ones,
execute following command:

## Usage

Starting service will start consuming messages.

[doc]: http://mainflux.readthedocs.io
