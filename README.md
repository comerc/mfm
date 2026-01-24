# nsfw-mod
Moderation NSFW

## TODO

- [ ] Добавить Grafana MCP

```json
    "grafana": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "GRAFANA_URL=http://host.docker.internal:3000",
        "-e", "GRAFANA_API_KEY=ваш_токен_здесь",
        "grafana/mcp-grafana:latest"
      ]
    }
```