#!/bin/bash
# Uso: ./start-task.sh [ID-TICKET] (ej. ./start-task.sh TS-14)

TICKET=$1
DOMAIN="turispro.atlassian.net"
EMAIL="moldrek@gmail.com"
API_TOKEN="ATATT3xFfGF01QqMEwZvIZqTWIIflk1xNgueoKbtlIPAXyaHel5-DMuxiHKOjikjQncP0pheupW5SFaZrUSNoMDodv5cX_G87R4DCRSk_D9V2OKgh94fwIlhxY0F7WbZyjHkYfCQeYbGO3jr7vDI9_jnyJTKEA6CFuR_XIipKZDWYU6fSBIZGSQ=EF14D1ED" # Idealmente ponlo en una variable de entorno $JIRA_API_TOKEN

if [ -z "$TICKET" ]; then
  echo "Error: Debes indicar el ID del ticket de Jira."
  exit 1
fi

echo "Descargando tarea $TICKET desde Jira..."

# Llamada a la API de Jira
RESPONSE=$(curl -s -u $EMAIL:$API_TOKEN \
  -X GET \
  -H "Accept: application/json" \
  "https://$DOMAIN/rest/api/2/issue/$TICKET")

# Extraer título y descripción usando jq
SUMMARY=$(echo $RESPONSE | jq -r '.fields.summary')
DESCRIPTION=$(echo $RESPONSE | jq -r '.fields.description')

if [ "$SUMMARY" == "null" ]; then
  echo "Error: No se encontró el ticket o error de autenticación."
  exit 1
fi

echo "Generando contexto para OpenCode..."

# Construir el prompt específico para esta tarea
PROMPT_FILE=".agent/current_task.md"

echo "### TAREA: $TICKET - $SUMMARY ###" > $PROMPT_FILE
echo "" >> $PROMPT_FILE
echo "$DESCRIPTION" >> $PROMPT_FILE

echo "Tarea lista en $PROMPT_FILE. Puedes inyectarla a OpenCode."
