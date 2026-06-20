# Mailgraph em servidor Postfix

Guia para instalar e executar o Mailgraph (versão Go) diretamente no mesmo servidor que roda o **Postfix**, ou via Docker montando o log local.

O Mailgraph lê o log de e-mail, grava estatísticas em arquivos RRD e exibe gráficos interativos em `http://<servidor>:8080/mailgraph/`.

---

## Requisitos

| Componente | Versão / detalhe |
|------------|------------------|
| Postfix | com logs em formato syslog |
| Go (somente para compilar) | 1.26+ |
| rrdtool | 1.7+ (obrigatório em runtime) |
| SO | Debian / Ubuntu (recomendado), ou outra distro com Postfix + rrdtool |

Gráficos suportados (quando os respectivos serviços geram log):

- enviados / recebidos / rejeitados / bounced
- SPF (`policyd-spf`)
- DMARC (`opendmarc`)
- DKIM (`opendkim`)
- vírus / spam (Amavis, ClamAV, SpamAssassin, etc.)
- logins Dovecot (se usar Dovecot no mesmo host)

---

## 1. Preparar o log do Postfix

O Mailgraph precisa de um arquivo de log legível com entradas do Postfix (e opcionalmente Dovecot, Amavis, OpenDKIM, etc.).

### Debian / Ubuntu com rsyslog

Crie `/etc/rsyslog.d/mailgraph.conf`:

```
# Log dedicado para o Mailgraph
$template MailgraphFormat,"%TIMESTAMP% %HOSTNAME% %syslogtag%%msg%\n"

if $programname startswith 'postfix'
   or $programname == 'policyd-spf'
   or $programname == 'opendmarc'
   or $programname == 'opendkim'
   or $programname == 'dovecot'
   or $programname startswith 'amavis'
then {
    /var/log/mail/mail.log;MailgraphFormat
    stop
}
```

Aplique e verifique:

```bash
sudo mkdir -p /var/log/mail
sudo chown syslog:adm /var/log/mail
sudo chmod 750 /var/log/mail
sudo systemctl restart rsyslog

# Deve aparecer linhas do Postfix após enviar/receber um e-mail de teste
sudo tail -f /var/log/mail/mail.log
```

Exemplo de linha esperada:

```
Jun 20 10:00:01 mail.example.com postfix/smtpd[1234]: ABCD: client=unknown[203.0.113.10]
```

### Caminhos de log alternativos

| Ambiente | Caminho comum |
|----------|---------------|
| rsyslog dedicado | `/var/log/mail/mail.log` |
| syslog geral | `/var/log/syslog` ou `/var/log/messages` |
| journald apenas | configure rsyslog ou redirecione para arquivo (recomendado) |

Se usar outro caminho, ajuste `log.file` em `config.toml` ou passe `--logfile` ao subcomando `server`.

---

## 2. Instalação do binário (nativo)

### 2.1 Dependências

```bash
sudo apt update
sudo apt install -y rrdtool git
```

### 2.2 Compilar

```bash
git clone https://github.com/jniltinho/MailgraphContainer.git
cd MailgraphContainer

make build
# ou: go build -trimpath -ldflags="-s -w" -o mailgraph .

sudo install -m 755 bin/mailgraph /usr/local/bin/mailgraph
```

### 2.3 Diretórios de dados

```bash
sudo mkdir -p /var/lib/mailgraph/rrd
sudo chown mailgraph:mailgraph /var/lib/mailgraph/rrd 2>/dev/null || sudo chown root:root /var/lib/mailgraph/rrd
```

Na primeira execução, se não existir RRD, o histórico atual de `/var/log/mail/mail.log` é processado automaticamente.

### 2.4 Arquivo de configuração (recomendado)

```bash
sudo mkdir -p /etc/mailgraph
sudo mailgraph generate-config
sudo cp config_*.toml /etc/mailgraph/config.toml
sudo nano /etc/mailgraph/config.toml
```

Exemplo para Postfix em produção:

```toml
[log]
file = "/var/log/mail/mail.log"
type = "syslog"
year = 2026

[rrd]
dir = "/var/lib/mailgraph/rrd"
name = "mailgraph"

[server]
listen = "127.0.0.1:8080"
hostname = "mail.example.com"

[filter]
ignore_localhost = true
```

Prioridade: flags > `MAILGRAPH_*` > `config.toml` > padrões. Ver [README.md](README.md#configuração) para a lista completa.

---

## 3. Executar manualmente (teste)

```bash
sudo mailgraph server \
  --logfile=/var/log/mail/mail.log \
  --daemon-rrd=/var/lib/mailgraph/rrd \
  --hostname=$(hostname -f) \
  --listen=127.0.0.1:8080
```

Com `config.toml` em `/etc/mailgraph/`:

```bash
sudo mailgraph server
```

Abra no navegador (via SSH tunnel ou proxy):

```
http://127.0.0.1:8080/mailgraph/
```

### Importar log histórico sem subir o servidor web

```bash
sudo mailgraph cat \
  --logfile=/var/log/mail/mail.log \
  --daemon-rrd=/var/lib/mailgraph/rrd \
  --year=$(date +%Y) \
  --verbose
```

### Opções úteis em Postfix

| Flag / config | Quando usar |
|---------------|-------------|
| `--ignore-localhost` / `filter.ignore_localhost` | Ignora tráfego de/para `127.0.0.1` (scanners locais, Amavis em loopback) |
| `--ignore-host=HOST` / `filter.ignore_hosts` | Ignora relay de um host específico (regex, repetível) |
| `--rbl-is-spam` / `filter.rbl_is_spam` | Conta rejeições RBL como spam |
| `--virbl-is-virus` / `filter.virbl_is_virus` | Conta rejeições VIRBL como vírus |
| `--host=mail.example.com` / `log.host_filter` | Filtra apenas entradas de um hostname no syslog |
| `--listen=127.0.0.1:8080` / `server.listen` | Escuta só em localhost (mais seguro) |

Exemplo com Amavis em localhost:

```bash
sudo mailgraph server \
  --logfile=/var/log/mail/mail.log \
  --daemon-rrd=/var/lib/mailgraph/rrd \
  --ignore-localhost \
  --hostname=mail.example.com \
  --listen=127.0.0.1:8080
```

---

## 4. Serviço systemd (produção)

Crie `/etc/systemd/system/mailgraph.service`:

```ini
[Unit]
Description=Mailgraph mail statistics
After=network.target rsyslog.service postfix.service
Wants=rsyslog.service

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/mailgraph server \
  --config /etc/mailgraph/config.toml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Alternativa sem arquivo de config (flags inline):

```ini
ExecStart=/usr/local/bin/mailgraph server \
  --logfile=/var/log/mail/mail.log \
  --daemon-rrd=/var/lib/mailgraph/rrd \
  --hostname=mail.example.com \
  --ignore-localhost \
  --listen=127.0.0.1:8080
```

Substitua `mail.example.com` pelo FQDN do seu servidor.

Ative o serviço:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now mailgraph
sudo systemctl status mailgraph
```

Logs do serviço:

```bash
sudo journalctl -u mailgraph -f
```

---

## 5. Instalação via Docker no mesmo servidor Postfix

Se o Postfix já roda no host, monte o log e o diretório RRD:

```bash
sudo mkdir -p /var/lib/mailgraph/rrd

docker run --rm -d \
  --name mailgraph \
  --restart unless-stopped \
  -v /var/log/mail/mail.log:/var/log/mail/mail.log:ro \
  -v /var/lib/mailgraph/rrd:/var/www/mailgraph/rrd \
  -v /etc/localtime:/etc/localtime:ro \
  -p 127.0.0.1:8080:8080 \
  davidullrich/mailgraph:latest
```

O entrypoint do container executa `mailgraph server` automaticamente. Para sobrescrever:

```bash
docker run --rm -d \
  --name mailgraph \
  -v /var/log/mail/mail.log:/var/log/mail/mail.log:ro \
  -v /var/lib/mailgraph/rrd:/var/www/mailgraph/rrd \
  -e MAILGRAPH_SERVER_HOSTNAME=mail.example.com \
  -e MAILGRAPH_FILTER_IGNORE_LOCALHOST=true \
  -p 127.0.0.1:8080:8080 \
  davidullrich/mailgraph:latest
```

Gráficos em `http://127.0.0.1:8080/mailgraph/`.

### Docker Compose

```yaml
services:
  mailgraph:
    image: davidullrich/mailgraph:latest
    hostname: mail.example.com
    restart: unless-stopped
    volumes:
      - /var/log/mail/mail.log:/var/log/mail/mail.log:ro
      - /var/lib/mailgraph/rrd:/var/www/mailgraph/rrd
      - /etc/localtime:/etc/localtime:ro
    ports:
      - "127.0.0.1:8080:8080"
```

---

## 6. Expor com segurança (Nginx + HTTPS)

Recomendado: Mailgraph escuta em `127.0.0.1:8080` e o Nginx faz proxy com autenticação.

`/etc/nginx/sites-available/mailgraph`:

```nginx
server {
    listen 443 ssl http2;
    server_name mail.example.com;

    ssl_certificate     /etc/letsencrypt/live/mail.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mail.example.com/privkey.pem;

    auth_basic "Mailgraph";
    auth_basic_user_file /etc/nginx/.htpasswd-mailgraph;

    location /mailgraph/ {
        proxy_pass http://127.0.0.1:8080/mailgraph/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Crie o usuário de acesso:

```bash
sudo apt install apache2-utils
sudo htpasswd -c /etc/nginx/.htpasswd-mailgraph admin
sudo ln -s /etc/nginx/sites-available/mailgraph /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Acesso: `https://mail.example.com/mailgraph/`

---

## 7. Verificação

```bash
# Serviço ativo
systemctl is-active mailgraph

# Log do Postfix chegando
sudo tail -5 /var/log/mail/mail.log

# RRDs sendo criados/atualizados
ls -la /var/lib/mailgraph/rrd/
# mailgraph.rrd  mailgraph_virus.rrd  mailgraph_dovecot.rrd

# Interface web
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/mailgraph/
# Esperado: 200
```

Envie um e-mail de teste (entrada e saída) e aguarde 1–2 minutos; os gráficos atualizam automaticamente a cada 5 minutos na página.

---

## 8. Solução de problemas

### Gráficos vazios

1. Confirme que `/var/log/mail/mail.log` recebe linhas `postfix/...`
2. Verifique permissão de leitura do usuário que roda o Mailgraph
3. Processe o log manualmente com `mailgraph cat --verbose` e observe erros de `rrdtool`
4. Confirme que `rrdtool` está instalado: `which rrdtool`

### RRD parou de atualizar

- Timestamps no log não podem retroceder (ajuste de relógio ou ano errado → use `--year` ou `log.year`)
- Inspecione o último timestamp: `rrdtool last /var/lib/mailgraph/rrd/mailgraph.rrd`

### Só aparece tráfego enviado, nada recebido

- Fetchmail ou relay local pode usar `127.0.0.1` → use `--ignore-localhost` ou ajuste o `smtphost` no fetchmail

### SPF / DKIM / DMARC sem dados

- O log precisa conter entradas de `policyd-spf`, `opendkim` e `opendmarc`
- Inclua esses programas no filtro do rsyslog (seção 1)

### Porta 8080 exposta na internet

- Prefira `server.listen = "127.0.0.1:8080"` + Nginx/Traefik com TLS e autenticação
- Não exponha estatísticas de e-mail publicamente sem proteção

---

## 9. Referência rápida de comandos

```bash
# Ajuda
mailgraph --help
mailgraph server --help

# Versão
mailgraph version

# Rodar em foreground (debug)
sudo mailgraph server --verbose \
  --logfile=/var/log/mail/mail.log \
  --daemon-rrd=/var/lib/mailgraph/rrd \
  --listen=127.0.0.1:8080

# Reprocessar log inteiro (sem servidor web)
sudo mailgraph cat \
  --logfile=/var/log/mail/mail.log \
  --daemon-rrd=/var/lib/mailgraph/rrd \
  --verbose

# Gerar config.toml
mailgraph generate-config
```

---

## Links

- [Mailgraph original](https://mailgraph.schweikert.ch/)
- [README Docker](README.md) — uso geral do container
- Patch SPF/DMARC/DKIM: [kernel-error.de](https://www.kernel-error.de/2014/04/22/mailgraph-graphen-um-spf-dmarc-und-dkim-erweitern/)