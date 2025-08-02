package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	apiURL             = "https://api.deepseek.com/v1/chat/completions"
	model              = "deepseek-chat"
	defaultMaxTokens   = 2048
	defaultTemperature = 0.7
	envFile            = ".env"
)

var (
	apiKey      string
	verbose     bool
	maxTokens   int
	temperature float64
	rawOutput   bool
	logger      *log.Logger
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type ResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func init() {
	// Configurar logger por defecto
	logger = log.New(os.Stderr, "", log.LstdFlags)
}

func loadEnv() error {
	// Intentar cargar el archivo .env
	err := godotenv.Load(envFile)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Printf("Archivo %s no encontrado, usando variables de entorno del sistema", envFile)
		} else {
			return fmt.Errorf("error cargando %s: %v", envFile, err)
		}
	}
	return nil
}

func printHelp() {
	helpText := `
deepcli - Asistente de desarrollo por terminal en español usando DeepSeek,
programado por nchgroup con <3 para la comunidad.

Uso básico:
  ` + os.Args[0] + ` -i "<consulta>" [opciones]
  ` + os.Args[0] + ` -i "<consulta>" -f <archivo>
  cat <archivo> | ` + os.Args[0] + ` -i "<consulta>"

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
     $ ` + os.Args[0] + ` -i "Cómo invertir un array en Python"

  2. Análisis de archivo:
     $ ` + os.Args[0] + ` -i "Explica este código" -f programa.js

  3. Pipeline Unix:
     $ git diff | ` + os.Args[0] + ` -i "Explica los cambios"

Ejemplos detallados:
  # Análisis de código con salida a archivo
  $ ` + os.Args[0] + ` -i "Detecta errores" -f codigo.py -o errores.txt

  # Refactorización estricta
  $ ` + os.Args[0] + ` -i "Refactoriza este código" -t 0.3 -m 4096 -f legacy.rs

  # Generación de documentación
  $ ` + os.Args[0] + ` -i "Genera documentación Markdown" -f module.go

  # Consulta interactiva (multilínea)
  $ ` + os.Args[0] + ` -i "$(cat <<EOF
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

`
	fmt.Println(helpText)
}

func main() {
	// Configuración de flags
	instruction := flag.String("i", "", "Instrucción para DeepSeek")
	outputFile := flag.String("o", "", "Archivo de salida para escribir la respuesta")
	inputFile := flag.String("f", "", "Archivo de entrada con el código a analizar")
	flag.Float64Var(&temperature, "t", defaultTemperature, "Temperatura para la generación (0.0-2.0)")
	flag.Float64Var(&temperature, "temperature", defaultTemperature, "Temperatura para la generación (0.0-2.0)")
	flag.IntVar(&maxTokens, "m", defaultMaxTokens, "Máximo número de tokens a generar")
	flag.IntVar(&maxTokens, "maxtokens", defaultMaxTokens, "Máximo número de tokens a generar")
	flag.BoolVar(&verbose, "v", false, "Mostrar mensajes detallados de ejecución")
	flag.BoolVar(&verbose, "verbose", false, "Mostrar mensajes detallados de ejecución")
	flag.BoolVar(&rawOutput, "raw", false, "Mostrar salida cruda en JSON (sin formatear)")
	showHelp := flag.Bool("h", false, "Mostrar ayuda")
	flag.BoolVar(showHelp, "help", false, "Mostrar ayuda")

	// Aliases para flags
	flag.StringVar(instruction, "instruction", "", "Instrucción para DeepSeek")
	flag.StringVar(outputFile, "output", "", "Archivo de salida para escribir la respuesta")
	flag.StringVar(inputFile, "file", "", "Archivo de entrada con el código a analizar")

	flag.Usage = func() {
		printHelp()
		os.Exit(0)
	}

	flag.Parse()

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Configurar logger según modo verboso
	if !verbose {
		logger.SetOutput(io.Discard)
	} else {
		logger.Println("Modo verboso activado")
	}

	// Validar temperatura
	if temperature < 0.0 || temperature > 2.0 {
		fmt.Fprintf(os.Stderr, "Error: La temperatura debe estar entre 0.0 y 2.0\n")
		os.Exit(1)
	}

	// Validar maxTokens
	if maxTokens <= 0 {
		fmt.Fprintf(os.Stderr, "Error: maxTokens debe ser mayor que 0\n")
		os.Exit(1)
	}

	// Cargar variables de entorno desde .env
	if err := loadEnv(); err != nil {
		logger.Printf("Advertencia: %v", err)
	}

	apiKey = os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: No se ha configurado la API key de DeepSeek. Por favor, establece la variable de entorno DEEPSEEK_API_KEY o crea un archivo .env con la clave.\n")
		os.Exit(1)
	}

	// Leer la entrada (puede ser de pipe, archivo o argumentos)
	var input string
	var err error

	// Verificar si hay datos en stdin (pipe)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		logger.Println("Leyendo datos de stdin...")
		var buf bytes.Buffer
		_, err = io.Copy(&buf, os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error al leer de stdin: %v\n", err)
			os.Exit(1)
		}
		input = buf.String()
		logger.Printf("Leídos %d bytes de stdin\n", len(input))
	}

	// Si se especificó un archivo de entrada, leerlo
	if *inputFile != "" {
		logger.Printf("Leyendo archivo de entrada: %s\n", *inputFile)
		fileContent, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error al leer el archivo de entrada: %v\n", err)
			os.Exit(1)
		}
		// Combinar con input de stdin si hubiera
		if input != "" {
			logger.Println("Combinando entrada de stdin con archivo de entrada")
		}
		input = strings.TrimSpace(input) + "\n" + string(fileContent)
		logger.Printf("Total de %d bytes de entrada\n", len(input))
	}

	// Obtener la instrucción
	var prompt string
	if *instruction != "" {
		prompt = *instruction
	} else if len(flag.Args()) > 0 {
		prompt = strings.Join(flag.Args(), " ")
	} else {
		logger.Println("Error: No se proporcionó instrucción")
		printHelp()
		os.Exit(1)
	}

	logger.Printf("Preparando solicitud con prompt: %s\n", prompt)
	logger.Printf("Configuración - Temperatura: %.2f, MaxTokens: %d\n", temperature, maxTokens)

	// Construir el mensaje para la API
	var messages []Message

	// Si hay input (de pipe o archivo), agregarlo como contexto
	if input != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: "Eres un asistente de programación experto. Ayudarás con código proporcionado por el usuario.",
		})

		messages = append(messages, Message{
			Role:    "user",
			Content: "Este es el código con el que necesito ayuda:\n" + input,
		})
	}

	// Agregar la instrucción del usuario
	messages = append(messages, Message{
		Role:    "user",
		Content: prompt,
	})

	logger.Printf("Preparando solicitud con %d mensajes de contexto\n", len(messages))

	// Crear el cuerpo de la solicitud
	requestBody := RequestBody{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      false,
	}

	// Convertir a JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Fatalf("Error al crear el cuerpo JSON: %v", err)
	}

	if verbose {
		logger.Printf("Cuerpo de la solicitud:\n%s\n", jsonBody)
	}

	// Crear la solicitud HTTP
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al crear la solicitud HTTP: %v\n", err)
		os.Exit(1)
	}

	// Configurar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	logger.Println("Enviando solicitud a la API...")

	// Realizar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al realizar la solicitud HTTP: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	logger.Printf("Respuesta recibida, código de estado: %d\n", resp.StatusCode)

	// Leer la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al leer la respuesta HTTP: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		logger.Printf("Respuesta cruda:\n%s\n", body)
	}

	// Si se solicita salida cruda, imprimir y salir
	if rawOutput {
		fmt.Println(string(body))
		return
	}

	// Parsear la respuesta
	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al parsear la respuesta JSON: %v\n", err)
		os.Exit(1)
	}

	// Manejar errores de la API
	if response.Error.Message != "" {
		fmt.Fprintf(os.Stderr, "Error de la API: %s\n", response.Error.Message)
		os.Exit(1)
	}

	// Mostrar la respuesta
	if len(response.Choices) > 0 {
		output := response.Choices[0].Message.Content

		// Si se especificó un archivo de salida, escribir en él
		if *outputFile != "" {
			logger.Printf("Escribiendo respuesta en archivo: %s\n", *outputFile)
			err := os.WriteFile(*outputFile, []byte(output), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error al escribir en el archivo de salida: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Respuesta escrita en %s\n", *outputFile)
		} else {
			// Mostrar en consola si no hay archivo de salida
			fmt.Println(output)
		}
	} else {
		fmt.Fprintf(os.Stderr, "No se recibió ninguna respuesta válida de la API")
		os.Exit(1)
	}
}
