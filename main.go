package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

const GroqServerURL = "https://api.groq.com/openai/v1/chat/completions"

func main() {
	groqAPIKey := os.Getenv("GROQ_API_KEY")

	// Interfaz Web Customizada (Temática Campo/Pampa con botón corregido)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8">
<title>🌾 Proyecto Pampa IA</title>
<style>
body { font-family: 'Segoe UI', sans-serif; background: #111827; color: #f3f4f6; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
.card { background: #1f2937; padding: 2.5rem; border-radius: 16px; border: 1px solid #059669; text-align: center; max-width: 500px; width: 90%%; box-shadow: 0 10px 30px rgba(0,0,0,0.5); }
h1 { color: #34d399; margin-top: 0; font-size: 2.4rem; font-weight: 800; letter-spacing: -0.05em; }
p { color: #9ca3af; font-size: 1.1rem; }
.input-box { width: 100%%; padding: 12px; border-radius: 8px; border: 1px solid #4b5563; background: #111827; color: white; margin-bottom: 15px; box-sizing: border-box; font-size: 1rem; }
.input-box:focus { border-color: #34d399; outline: none; }
.btn { background: #10b981; color: #ffffff; border: none; padding: 12px 20px; border-radius: 8px; font-weight: bold; font-size: 1rem; cursor: pointer; width: 100%%; transition: background 0.2s; }
.btn:hover { background: #059669; }
#response { margin-top: 20px; padding: 15px; background: #111827; border-radius: 8px; text-align: left; font-size: 1rem; max-height: 250px; overflow-y: auto; color: #e5e7eb; border-left: 4px solid #10b981; display: none; line-height: 1.5; }
</style>
</head>
<body>
<div class="card">
<h1>🌾 Proyecto Pampa</h1>
<p>Arrimate al fogón y preguntá lo que quieras:</p>
<input type="text" id="prompt" class="input-box" value="Che Pampa, contame qué es Go y para qué sirve posta.">
<button class="btn" onclick="askPampa()">Cebar Pregunta 🧉</button>
<div id="response"></div>
</div>
<script>
function askPampa() {
var promptText = document.getElementById('prompt').value;
var resDiv = document.getElementById('response');
resDiv.style.display = 'block';
resDiv.innerText = 'Preparando el mate (pensando)...';
fetch('/ask?q=' + encodeURIComponent(promptText))
.then(response => response.json())
.then(data => { resDiv.innerText = data.reply; })
.catch(err => { resDiv.innerText = "Se enfrió el agua (Error de conexión)."; });
}
</script>
</body>
</html>`)
	})

	// Ruta de consulta a Groq con la nueva personalidad campera
	http.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("q")

		if groqAPIKey == "" {
			json.NewEncoder(w).Encode(map[string]string{"reply": "[ERROR] No se detecta la variable GROQ_API_KEY."})
			return
		}

		reqGroq := ChatRequest{
			Model: "llama-3.1-8b-instant",
			Messages: []Message{
				{
					Role: "system", 
					Content: "Sos Pampa, un asistente de IA gaucho, moderno, tecnológico, amigable y muy argentino (usás palabras como 'che', 'viste', 'totalmente', 'un caño'). Sos extremadamente inteligente y preciso con los datos técnicos de programación. Explicás las cosas de forma clara, con metáforas criollas pero con total rigurosidad técnica, sin inventar nada. Respondés de forma concisa.",
				},
				{Role: "user", Content: query},
			},
		}

		jsonData, _ := json.Marshal(reqGroq)
		req, _ := http.NewRequest("POST", GroqServerURL, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+groqAPIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"reply": "[ERROR DE RED] " + err.Error()})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			json.NewEncoder(w).Encode(map[string]string{"reply": fmt.Sprintf("[ERROR GROQ %d] %s", resp.StatusCode, string(body))})
			return
		}

		var chatResp ChatResponse
		json.Unmarshal(body, &chatResp)

		replyText := "No se recibió respuesta del modelo."
		if len(chatResp.Choices) > 0 {
			replyText = chatResp.Choices[0].Message.Content
		}

		json.NewEncoder(w).Encode(map[string]string{"reply": replyText})
	})

	fmt.Println("Servidor de Pampa iniciado en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
