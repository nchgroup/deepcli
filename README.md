# deepcli
Asistente de desarrollo por terminal en español usando deepseek

# Help

```
$ ./deepcli -h

deepcli - Asistente de desarrollo por terminal en español usando DeepSeek,
programado por NCH con <3 para la comunidad.

Uso básico:
  ./deepcli -i "<consulta>" [opciones]
  ./deepcli -i "<consulta>" -f <archivo>
  cat <archivo> | ./deepcli -i "<consulta>"

Descripción:
  Herramienta para interactuar con DeepSeek Chat desde terminal,
  ideal para análisis de código, asistencia técnica y generación de contenido.

Argumentos principales:
  -i, --instruction <texto>   Consulta/prompt para la IA (requerido)
  -f, --file <archivo>        Archivo a analizar (opcional)
  -o, --output <archivo>      Guardar respuesta en archivo (opcional)

Opciones de modelo:
  -t, --temperature <0.0-2.0> Controla aleatoriedad:
                              0.0 = preciso/factico
                              0.7 = balanceado (default)
                              1.5+ = creativo/arriesgado
  -m, --maxtokens <número>    Longitud máxima de respuesta (default: 2048)

Modos de entrada:
  1. Consulta directa:
     $ ./deepcli -i "Cómo invertir un array en Python"

  2. Análisis de archivo:
     $ ./deepcli -i "Explica este código" -f programa.js

  3. Pipeline Unix:
     $ git diff | ./deepcli -i "Explica los cambios"

Ejemplos detallados:
  # Análisis de código con salida a archivo
  $ ./deepcli -i "Detecta errores" -f codigo.py -o errores.txt

  # Refactorización estricta
  $ ./deepcli -i "Refactoriza este código" -t 0.3 -m 4096 -f legacy.rs

  # Generación de documentación
  $ ./deepcli -i "Genera documentación Markdown" -f module.go

  # Consulta interactiva (multilínea)
  $ ./deepcli -i "$(cat <<EOF
Explica los patrones de diseño usados en este
código y sugiere mejoras de rendimiento
EOF
)" -f sistema.py

Configuración:
  La API key se configura mediante:
  • Variable de entorno: $ export DEEPSEEK_API_KEY="tu_key"
  • Archivo .env: $ echo 'DEEPSEEK_API_KEY=tu_key' > .env

Opciones avanzadas:
  -raw                  Salida sin formato (para procesamiento pipeline)
  -v, --verbose         Mostrar logs detallados
  -h, --help            Mostrar esta ayuda

Sugerencias:
  • Para código complejo, usa --maxtokens 4096
  • Combina con jq para procesar JSON: -raw | jq '.choices[0]...'
  • Usa --temperature 1.2 para brainstorming creativo
  • El contexto máximo es 128K tokens


```
