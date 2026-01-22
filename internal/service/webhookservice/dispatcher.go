package webhookservice

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/internal/repo"
)

type Dispatcher struct {
	repo       repo.WebhookRepo
	httpClient *http.Client
	logger     *slog.Logger
	eventChan  chan *dispatchJob
}

type dispatchJob struct {
	webhook *entity.Webhook
	event   *WebhookEvent
}

func NewDispatcher(repo repo.WebhookRepo, logger *slog.Logger) *Dispatcher {
	d := &Dispatcher{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:    logger,
		eventChan: make(chan *dispatchJob, 1000),
	}

	// inicia workers
	for range 5 {
		go d.worker()
	}

	return d
}

func (d *Dispatcher) worker() {
	for job := range d.eventChan {
		d.sendWebhook(context.Background(), job.webhook, job.event)
	}
}

// envia um evento para todos os webhooks configurados do usuário
func (d *Dispatcher) Dispatch(ctx context.Context, userID uint, event *WebhookEvent) {
	webhooks, err := d.repo.GetActiveByUserAndEvent(ctx, userID, string(event.Type))
	if err != nil {
		d.logger.Error("failed to get webhooks", "error", err, "userID", userID)
		return
	}

	for _, webhook := range webhooks {
		w := webhook // para goroutine
		d.eventChan <- &dispatchJob{
			webhook: &w,
			event:   event,
		}
	}
}

func (d *Dispatcher) sendWebhook(ctx context.Context, webhook *entity.Webhook, event *WebhookEvent) {
	startTime := time.Now()

	payload, err := json.Marshal(event)
	if err != nil {
		d.logger.Error("failed to marshal event", "error", err)
		return
	}

	// assinatura HMAC
	signature := d.sign(payload, webhook.Secret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(payload))
	if err != nil {
		d.logger.Error("failed to create request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", string(event.Type))
	req.Header.Set("X-Webhook-ID", event.ID)
	req.Header.Set("X-Webhook-Timestamp", event.Timestamp.Format(time.RFC3339))

	resp, err := d.httpClient.Do(req)
	duration := time.Since(startTime).Milliseconds()

	log := &entity.WebhookLog{
		WebhookID: webhook.ID,
		EventType: string(event.Type),
		Payload:   string(payload),
		Duration:  duration,
	}

	if err != nil {
		log.Error = err.Error()
		d.repo.IncrementFailCount(ctx, webhook.ID)
		d.logger.Error("webhook request failed", "error", err, "webhookID", webhook.ID, "url", webhook.URL)
	} else {
		defer resp.Body.Close()
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)

		log.ResponseBody = string(buf[:n])

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			d.repo.ResetFailCount(ctx, webhook.ID)
			d.logger.Info("webhook sent successfully", "webhookID", webhook.ID, "status", resp.StatusCode)
		} else {
			d.repo.IncrementFailCount(ctx, webhook.ID)
			d.logger.Warn("webhook returned error", "webhookID", webhook.ID, "status", resp.StatusCode)
		}
	}

	if err := d.repo.UpdateLastUsed(ctx, webhook.ID); err != nil {
		d.logger.Error("failed to update webhook last used", "error", err, "webhookID", webhook.ID)
	}

	if err := d.repo.InsertLog(ctx, log); err != nil {
		d.logger.Error("failed to save webhook log", "error", err)
	}
}

func (d *Dispatcher) sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
