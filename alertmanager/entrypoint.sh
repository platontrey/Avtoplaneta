#!/bin/sh
set -e

CONFIG_FILE="/etc/alertmanager/alertmanager.yml"

if [ -f "/etc/alertmanager/alertmanager.template.yml" ]; then
  cp /etc/alertmanager/alertmanager.template.yml /tmp/alertmanager.yml
  
  sed -i "s|__EMAIL_SMARTHOST__|${ALERT_EMAIL_SMARTHOST:-smtp.example.com:587}|g" /tmp/alertmanager.yml
  sed -i "s|__EMAIL_FROM__|${ALERT_EMAIL_FROM:-alerts@example.com}|g" /tmp/alertmanager.yml
  sed -i "s|__EMAIL_AUTH_USERNAME__|${ALERT_EMAIL_AUTH_USERNAME:-}|g" /tmp/alertmanager.yml
  sed -i "s|__EMAIL_AUTH_PASSWORD__|${ALERT_EMAIL_AUTH_PASSWORD:-}|g" /tmp/alertmanager.yml
  sed -i "s|__TELEGRAM_BOT_TOKEN__|${ALERT_TELEGRAM_BOT_TOKEN:-disabled}|g" /tmp/alertmanager.yml
  sed -i "s|__TELEGRAM_CHAT_ID__|${ALERT_TELEGRAM_CHAT_ID:-0}|g" /tmp/alertmanager.yml
  sed -i "s|__EMAIL_TO__|${ALERT_EMAIL_TO:-alerts@example.com}|g" /tmp/alertmanager.yml
  
  CONFIG_FILE="/tmp/alertmanager.yml"
fi

exec /bin/alertmanager --config.file="${CONFIG_FILE}" --storage.path=/alertmanager "$@"
