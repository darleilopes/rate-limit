O arquivo `config.yaml` possui as configurações de limitação de taxa para endereços IP e tokens de API:

```yaml
rate_limit:
  ips:
    - ip: "192.168.1.1"
      limit: 10
    - ip: "192.168.1.2"
      limit: 2
      expiration: 5
      block: 10
  tokens:
    - token: "tokenzinho"
      limit: 2
      expiration: 5
      block: 10
```

Onde: 
- **limit**: é número máximo de requisições permitido antes de o sistema bloquear novas requisições do token/ip
- **expiration**: o tempo em segundos quando contagem de requisições do token/ip é resetada
- **block**: o tempo em segundos quando o token/ip é bloqueado após exceder o limite

Quando uma dessas variaveis nao é setada, o sistema seta valores padrão já definidos no arquivo de configuração do `docker-compose.yaml`

## Arquivo docker-compose.yaml

- No serviço `rate_limiter` é possível configurar variáveis de ambiente como `DEFAULT_IP_EXPIRATION_TIME`, `DEFAULT_IP_BLOCK_DURATION`, etc., bem como definir limites padrão e comportamentos de limitação. 
  - **DEFAULT_IP_EXPIRATION_TIME**: é tempo em segundos quando a contagem de requisições do IP é resetada
  - **DEFAULT_TOKEN_EXPIRATION_TIME**: é o tempo em segundos quando a contagem de requisições do token é resetada
  - **DEFAULT_IP_REQUEST_LIMIT**: é número máximo de requisições permitidas antes de o sistema bloquear novas requisições do IP
  - **DEFAULT_TOKEN_REQUEST_LIMIT**: é número máximo de requisições permitidas antes de o sistema bloquear novas requisições do token
  - **DEFAULT_IP_BLOCK_DURATION**: é o tempo em segundos, quando o IP é bloqueado após exceder o limite
  - **DEFAULT_TOKEN_BLOCK_DURATION**: é o tempo em segundos quando o token é bloqueado após exceder o limite

## Levantando o Projeto

```bash
docker-compose up --build
```

```bash
curl -X POST http://localhost:8080/ \
     -H "Content-Type: application/json" \
     -H "API_KEY: tokenzinho"
```