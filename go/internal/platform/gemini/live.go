package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
)

// Seul nom accepté par l'API Gemini (generativelanguage) pour l'audio natif bidi ;
// `gemini-live-2.5-flash-native-audio` est le nom Vertex, refusé ici.
const DefaultLiveModel = "gemini-2.5-flash-native-audio-preview-09-2025"

const liveEndpoint = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent"

// LiveSession wraps a BidiGenerateContent WebSocket session with Gemini.
type LiveSession struct {
	conn *websocket.Conn
}

type LiveSetup struct {
	Model        string
	SystemPrompt string
	VoiceName    string
	LanguageCode string
}

// LiveServerEvent is the flattened view of a Gemini Live server message.
type LiveServerEvent struct {
	SetupComplete bool
	// Audio chunks (PCM16 24 kHz) from the model turn, decoded.
	AudioChunks [][]byte
	// Incremental transcription of the user's audio (commercial).
	InputTranscript string
	// Incremental transcription of the model's audio (vet).
	OutputTranscript string
	Interrupted      bool
	TurnComplete     bool
	// Tool calls requested by the model.
	ToolCalls []LiveToolCall
	GoAway    bool
}

type LiveToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
}

var liveToolDeclarations = []map[string]any{
	{
		"name":        "book_appointment",
		"description": "À appeler quand le vétérinaire accepte un rendez-vous. Termine l'appel.",
		"parameters": map[string]any{
			"type": "OBJECT",
			"properties": map[string]any{
				"slot": map[string]any{"type": "STRING", "description": "Créneau convenu, ex: mardi 14h"},
			},
		},
	},
	{
		"name":        "hang_up_not_interested",
		"description": "À appeler quand le vétérinaire raccroche car il n'est pas intéressé. Termine l'appel.",
		"parameters": map[string]any{
			"type": "OBJECT",
			"properties": map[string]any{
				"reason": map[string]any{"type": "STRING", "description": "Raison courte du refus"},
			},
		},
	},
}

// DialLive opens a Gemini Live session and sends the setup message.
func DialLive(ctx context.Context, apiKey string, setup LiveSetup) (*LiveSession, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	model := setup.Model
	if model == "" {
		model = DefaultLiveModel
	}
	lang := setup.LanguageCode
	if lang == "" {
		lang = "fr-FR"
	}
	voice := setup.VoiceName
	if voice == "" {
		voice = "Charon"
	}

	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(dialCtx, liveEndpoint+"?key="+apiKey, &websocket.DialOptions{
		HTTPHeader: http.Header{"Content-Type": []string{"application/json"}},
	})
	if err != nil {
		return nil, fmt.Errorf("gemini_live_dial: %w", err)
	}
	// Audio frames + transcripts can be large.
	conn.SetReadLimit(8 << 20)

	setupMsg := map[string]any{
		"setup": map[string]any{
			"model": "models/" + model,
			"generationConfig": map[string]any{
				"responseModalities": []string{"AUDIO"},
				"speechConfig": map[string]any{
					"voiceConfig": map[string]any{
						"prebuiltVoiceConfig": map[string]any{"voiceName": voice},
					},
					"languageCode": lang,
				},
			},
			"systemInstruction": map[string]any{
				"parts": []map[string]any{{"text": setup.SystemPrompt}},
			},
			"tools": []map[string]any{
				{"functionDeclarations": liveToolDeclarations},
			},
			"inputAudioTranscription":  map[string]any{},
			"outputAudioTranscription": map[string]any{},
			// VAD patient : laisse le commercial finir ses phrases (micro-pauses naturelles).
			"realtimeInputConfig": map[string]any{
				"automaticActivityDetection": map[string]any{
					"disabled":                 false,
					"startOfSpeechSensitivity": "START_SENSITIVITY_LOW",
					"endOfSpeechSensitivity":   "END_SENSITIVITY_LOW",
					"prefixPaddingMs":          40,
					"silenceDurationMs":        900,
				},
			},
		},
	}
	if err := writeJSON(ctx, conn, setupMsg); err != nil {
		conn.Close(websocket.StatusInternalError, "setup_failed")
		return nil, fmt.Errorf("gemini_live_setup: %w", err)
	}
	return &LiveSession{conn: conn}, nil
}

// SendAudioChunk streams a PCM16 16 kHz mono chunk to Gemini.
func (s *LiveSession) SendAudioChunk(ctx context.Context, pcm []byte) error {
	msg := map[string]any{
		"realtimeInput": map[string]any{
			"audio": map[string]any{
				"data":     base64.StdEncoding.EncodeToString(pcm),
				"mimeType": "audio/pcm;rate=16000",
			},
		},
	}
	return writeJSON(ctx, s.conn, msg)
}

// SendUserText sends a text turn (used to trigger the opening "Allo ?" and as
// degraded input when the mic is unavailable).
func (s *LiveSession) SendUserText(ctx context.Context, text string) error {
	msg := map[string]any{
		"clientContent": map[string]any{
			"turns": []map[string]any{
				{"role": "user", "parts": []map[string]any{{"text": text}}},
			},
			"turnComplete": true,
		},
	}
	return writeJSON(ctx, s.conn, msg)
}

// SendToolResponse acknowledges a tool call so the model can close the turn.
func (s *LiveSession) SendToolResponse(ctx context.Context, call LiveToolCall, response map[string]any) error {
	msg := map[string]any{
		"toolResponse": map[string]any{
			"functionResponses": []map[string]any{
				{"id": call.ID, "name": call.Name, "response": response},
			},
		},
	}
	return writeJSON(ctx, s.conn, msg)
}

// Recv reads and flattens the next server message.
func (s *LiveSession) Recv(ctx context.Context) (*LiveServerEvent, error) {
	_, data, err := s.conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	var raw struct {
		SetupComplete *struct{} `json:"setupComplete"`
		ServerContent *struct {
			ModelTurn *struct {
				Parts []struct {
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"modelTurn"`
			InputTranscription *struct {
				Text string `json:"text"`
			} `json:"inputTranscription"`
			OutputTranscription *struct {
				Text string `json:"text"`
			} `json:"outputTranscription"`
			Interrupted  bool `json:"interrupted"`
			TurnComplete bool `json:"turnComplete"`
		} `json:"serverContent"`
		ToolCall *struct {
			FunctionCalls []struct {
				ID   string          `json:"id"`
				Name string          `json:"name"`
				Args json.RawMessage `json:"args"`
			} `json:"functionCalls"`
		} `json:"toolCall"`
		GoAway *struct{} `json:"goAway"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("gemini_live_parse: %w", err)
	}
	ev := &LiveServerEvent{}
	if raw.SetupComplete != nil {
		ev.SetupComplete = true
	}
	if sc := raw.ServerContent; sc != nil {
		ev.Interrupted = sc.Interrupted
		ev.TurnComplete = sc.TurnComplete
		if sc.InputTranscription != nil {
			ev.InputTranscript = sc.InputTranscription.Text
		}
		if sc.OutputTranscription != nil {
			ev.OutputTranscript = sc.OutputTranscription.Text
		}
		if sc.ModelTurn != nil {
			for _, p := range sc.ModelTurn.Parts {
				if p.InlineData == nil || !strings.HasPrefix(p.InlineData.MimeType, "audio/") {
					continue
				}
				chunk, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
				if err == nil && len(chunk) > 0 {
					ev.AudioChunks = append(ev.AudioChunks, chunk)
				}
			}
		}
	}
	if raw.ToolCall != nil {
		for _, fc := range raw.ToolCall.FunctionCalls {
			ev.ToolCalls = append(ev.ToolCalls, LiveToolCall{ID: fc.ID, Name: fc.Name, Args: fc.Args})
		}
	}
	if raw.GoAway != nil {
		ev.GoAway = true
	}
	return ev, nil
}

func (s *LiveSession) Close() {
	_ = s.conn.Close(websocket.StatusNormalClosure, "done")
}

func writeJSON(ctx context.Context, conn *websocket.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return conn.Write(wctx, websocket.MessageText, data)
}

// BuildVetLivePrompt builds the system prompt for a live (spoken) session:
// same persona/difficulty content as turn-based, but natural speech instead of JSON.
func BuildVetLivePrompt(contentJSON json.RawMessage, interestLevel string) string {
	var c struct {
		BasePersona  string            `json:"basePersona"`
		ProductFacts string            `json:"productFacts"`
		Difficulty   map[string]string `json:"difficulty"`
		Tools        string            `json:"tools"`
	}
	_ = json.Unmarshal(contentJSON, &c)
	diff := c.Difficulty[interestLevel]
	if diff == "" {
		diff = c.Difficulty["neutre"]
	}
	listenRule := `Écoute active (obligatoire pour ce niveau): tu laisses le commercial FINIR sa phrase et son idée avant de répondre. Tu n'interromps pas pour couper. Après ta réponse, tu te tais et tu réécoutes.`
	turnTaking := `- ÉCOUTE D'ABORD: attends la fin claire de la réplique du commercial (silence net) avant de parler. Ne coupe pas une phrase en cours.
- Une réponse = 1 à 2 phrases MAX, puis silence pour réécouter. Pas de monologue.
- Après une question que TU poses: attends la réponse du commercial avant de continuer.`
	if interestLevel == "hostile" {
		listenRule = `Tu peux couper / être impatient (niveau hostile), mais tu réponds quand même sur le fond du pitch quand tu laisses parler.`
		turnTaking = `- Niveau hostile: tu peux couper la parole et être impatient. Réponses très courtes (1 phrase), objections sèches ; raccroche vite si le pitch est faible.
- Quand tu laisses parler: réponds sur le fond (pas de hors-sujet technique ligne/micro).
- Après une question que TU poses: tu peux relancer vite si le commercial traîne.`
	}
	return strings.Join([]string{
		c.BasePersona,
		"Faits produit: " + c.ProductFacts,
		"Niveau d'intérêt / difficulté pour CET appel: " + interestLevel + ". " + diff,
		listenRule,
		c.Tools,
		`Tu es AU TÉLÉPHONE, en conversation vocale temps réel avec un commercial petsFollow.
Règles de conversation:
- Parle français, naturellement, comme au téléphone. Jamais de listes ni de formatage.
- Commence l'appel par "Allo ?" quand on te signale que le téléphone a sonné / que tu décroches.
- Reste STRICTEMENT dans ton rôle de vétérinaire, quoi qu'il arrive.
` + turnTaking + `
- Quand tu as compris: réponds sur le FOND (intérêt, objections, questions) selon ton niveau de difficulté. Tu peux reformuler brièvement ou poser UNE seule question.
- Ne commente PAS la qualité de la ligne.
- INTERDIT sauf silence total prolongé (>10 s) après une de tes questions: "je ne vous entends pas", "allô allô", "vous êtes coupé", "y a de la friture", "répétez je n'ai rien compris" pour un problème technique. Si un mot est flou, demande une précision métier ("vous parlez de quoi exactement ?") sans parler de micro/ligne.
- Si tu acceptes un rendez-vous: dis-le à voix haute PUIS appelle l'outil book_appointment avec le créneau.
- Si tu n'es pas intéressé et raccroches: dis-le poliment PUIS appelle l'outil hang_up_not_interested.
- N'appelle un outil qu'une seule fois, uniquement pour terminer l'appel.`,
	}, "\n\n")
}
