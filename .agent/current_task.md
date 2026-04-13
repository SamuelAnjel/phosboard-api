### TAREA: PHOS-40 - CRUD de Fuentes y Control de Límites en Discovery ###

La base de datos ya almacena fuentes, pero necesitamos una API HTTP para administrarlas y evitar que el worker {{discovery}} sature la cola de Pub/Sub extrayendo links de forma descontrolada. Implementaremos un CRUD para la tabla {{sources}} y un límite de extracción por fuente.

*Instrucciones de Ejecución:*

# *Actualización de Base de Datos:*
#* Crea una nueva migración SQL para alterar la tabla {{sources}}: agrega la columna {{config JSONB DEFAULT '{}'::jsonb}}.
#* Esta columna almacenará parámetros como {{{"max_links": 20}}}.
# *CRUD de Sources (API Backend):*
#* Crea {{internal/repository/source_repository.go}} y {{internal/handler/source_handler.go}}.
#* Implementa los endpoints protegidos por el middleware JWT bajo el grupo {{/api/v1/tenants/{tenant_id}/sources}}:
#** {{GET}} (Listar fuentes del tenant).
#** {{POST}} (Crear fuente: recibe {{name}}, {{type}}, {{url}}, {{interval_minutes}} y {{config}}).
#** {{PUT}} y {{DELETE}}.
# *Refactor del Worker Discovery:*
#* Modifica el worker en {{workers/discovery}} para que lea el campo {{config}} de la fuente que está procesando.
#* Extrae el valor numérico de {{max_links}} (si no existe en el JSONB, usa un default seguro, ej. 20).
#* Al iterar sobre los items del RSS ({{gofeed}}), implementa un contador. Si el contador alcanza el límite de {{max_links}}, interrumpe el ciclo ({{break}}) para no encolar más mensajes hacia el tópico {{url-scrape}}.

*Reglas Estrictas:*

* Validar que la URL inyectada en el {{POST}} tenga un formato válido.
* Asegurar que el handler verifique que el {{tenant_id}} de la ruta coincide con el {{tenant_id}} del token JWT.
