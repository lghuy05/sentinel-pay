package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type Handler struct {
	store       *Store
	redis       *redis.Client
	brokers     []string
	serviceURLs map[string]string
	mlBaseURL   string
	client      *http.Client
}

func NewHandler(store *Store, redisClient *redis.Client, brokers []string, serviceURLs map[string]string, mlBaseURL string) *Handler {
	return &Handler{
		store:       store,
		redis:       redisClient,
		brokers:     brokers,
		serviceURLs: serviceURLs,
		mlBaseURL:   strings.TrimRight(mlBaseURL, "/"),
		client:      &http.Client{Timeout: 3 * time.Second},
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health/"+ServiceName, h.health)
	mux.HandleFunc("/actuator/health", h.health)
	mux.HandleFunc("/decisions", h.listDecisions)
	mux.HandleFunc("/decisions/", h.getDecision)
	mux.HandleFunc("/feedback", h.submitFeedback)
	mux.HandleFunc("/ml/status", h.mlStatus)
	mux.HandleFunc("/ml/retrain", h.mlRetrain)
	mux.HandleFunc("/ml/reload", h.mlReload)
	mux.HandleFunc("/system/kafka", h.kafka)
	mux.HandleFunc("/system/redis", h.redisStatus)
	mux.HandleFunc("/system/services", h.services)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]string{"status": "UP", "service": ServiceName})
}

func (h *Handler) listDecisions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	var reviewed *bool
	if raw := r.URL.Query().Get("reviewed"); raw == "false" {
		value := false
		reviewed = &value
	}
	records, err := h.store.List(r.Context(), limit, reviewed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, records)
}

func (h *Handler) getDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	transactionID := r.URL.Path[len("/decisions/"):]
	record, ok, err := h.store.Get(r.Context(), transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Decision not found", http.StatusNotFound)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, record)
}

type feedbackRequest struct {
	TransactionID string `json:"transactionId"`
	Label         *int   `json:"label"`
}

func (h *Handler) submitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid feedback payload", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(request.TransactionID) == "" || request.Label == nil {
		http.Error(w, "transactionId and label are required", http.StatusBadRequest)
		return
	}
	found, err := h.store.SubmitFeedback(r.Context(), request.TransactionID, *request.Label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Decision not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) mlStatus(w http.ResponseWriter, r *http.Request) {
	h.proxyML(w, r, http.MethodGet, "/ml/status")
}

func (h *Handler) mlRetrain(w http.ResponseWriter, r *http.Request) {
	h.proxyML(w, r, http.MethodPost, "/ml/retrain")
}

func (h *Handler) mlReload(w http.ResponseWriter, r *http.Request) {
	h.proxyML(w, r, http.MethodPost, "/ml/reload")
}

func (h *Handler) proxyML(w http.ResponseWriter, r *http.Request, method, path string) {
	if r.Method != method {
		w.Header().Set("Allow", method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body io.Reader
	if r.Body != nil {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read request body", http.StatusBadRequest)
			return
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(r.Context(), method, h.mlBaseURL+path, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()
	resp, err := h.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Handler) kafka(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload := map[string]any{"status": "DOWN"}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	conn, err := kafka.DialContext(ctx, "tcp", h.brokers[0])
	if err == nil {
		defer conn.Close()
		partitions, err := conn.ReadPartitions()
		if err == nil {
			topics := map[string]bool{}
			for _, partition := range partitions {
				topics[partition.Topic] = true
			}
			names := make([]string, 0, len(topics))
			for topic := range topics {
				names = append(names, topic)
			}
			payload = map[string]any{"status": "UP", "topics": names}
		}
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	httpserver.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) redisStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload := map[string]any{"status": "UP"}
	if err := h.redis.Ping(r.Context()).Err(); err != nil {
		payload["status"] = "DOWN"
		payload["error"] = err.Error()
	}
	httpserver.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) services(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload := make(map[string]any, len(h.serviceURLs))
	for name, url := range h.serviceURLs {
		payload[name] = h.probe(r.Context(), url)
	}
	httpserver.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) probe(ctx context.Context, url string) map[string]any {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return map[string]any{"status": "DOWN", "error": err.Error()}
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return map[string]any{"status": "DOWN", "error": err.Error()}
	}
	defer resp.Body.Close()
	result := map[string]any{"status": "UP"}
	var body any
	if err := json.NewDecoder(resp.Body).Decode(&body); err == nil {
		result["response"] = body
	}
	if resp.StatusCode >= 400 {
		result["status"] = "DOWN"
		result["code"] = resp.StatusCode
	}
	return result
}
