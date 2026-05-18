package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"inhumas-em-foco/internal/automation"
	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/sitemap"
	"inhumas-em-foco/internal/storage"
)

func main() {
	cfg, err := config.LoadWithError()
	if err != nil {
		log.Fatal(err)
	}

	repo, err := repository.Open(cfg.DBDriver, cfg.DatabaseURL, cfg.MigrationsDir)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco: ", err)
	}
	defer repo.Close()

	fmt.Println("Worker iniciado")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ensureRecurringJobs(repo)

	// Run once immediately
	processJobs(repo, cfg)

	for {
		select {
		case <-ticker.C:
			ensureRecurringJobs(repo)
			processJobs(repo, cfg)
		case <-signalChan():
			fmt.Println("Worker encerrando...")
			return
		}
	}
}

func signalChan() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}

func processJobs(repo *repository.Repository, cfg *config.Config) {
	ctx := context.Background()
	jobs, err := repo.JobClaimPending(ctx, 10)
	if err != nil {
		log.Println("Erro ao buscar jobs:", err)
		return
	}

	for _, job := range jobs {
		if err := executeJob(repo, cfg, job); err != nil {
			log.Printf("Job %d falhou: %v", job.ID, err)
			movedToDead, updateErr := repo.JobRecordFailure(ctx, job, err.Error())
			if updateErr != nil {
				log.Printf("Erro ao atualizar falha do job %d: %v", job.ID, updateErr)
				continue
			}
			if movedToDead {
				log.Printf("Job %d movido para dead_jobs apos %d tentativas", job.ID, job.Attempts+1)
			}
		} else {
			if err := repo.JobUpdateStatus(ctx, job.ID, model.JobCompleted, ""); err != nil {
				log.Printf("Erro ao marcar job %d como concluido: %v", job.ID, err)
				continue
			}
			scheduleNextRecurring(repo, job.Type)
		}
	}
}

func ensureRecurringJobs(repo *repository.Repository) {
	ctx := context.Background()
	recurring := []struct {
		jobType model.JobType
		runAt   time.Time
	}{
		{model.JobBackupDatabase, nextDaily(3, 0)},
		{model.JobGenerateSitemap, nextDaily(2, 0)},
		{model.JobVacuumDB, nextWeekly(time.Sunday, 4, 0)},
		{model.JobCleanupOldJobs, nextWeekly(time.Sunday, 5, 0)},
		{model.JobCompressOldUploads, nextWeekly(time.Sunday, 5, 30)},
	}
	for _, job := range recurring {
		scheduleIfMissing(repo, job.jobType, job.runAt)
	}
	settings, err := repo.PortalSettingsGet(ctx)
	if err == nil && settings.AutomationEnabled {
		scheduleIfMissing(repo, model.JobCollectNews, time.Now().Add(automationInterval(settings)))
	}
}

func scheduleNextRecurring(repo *repository.Repository, jobType model.JobType) {
	switch jobType {
	case model.JobBackupDatabase:
		scheduleIfMissing(repo, jobType, nextDaily(3, 0))
	case model.JobGenerateSitemap:
		scheduleIfMissing(repo, jobType, nextDaily(2, 0))
	case model.JobVacuumDB:
		scheduleIfMissing(repo, jobType, nextWeekly(time.Sunday, 4, 0))
	case model.JobCleanupOldJobs:
		scheduleIfMissing(repo, jobType, nextWeekly(time.Sunday, 5, 0))
	case model.JobCompressOldUploads:
		scheduleIfMissing(repo, jobType, nextWeekly(time.Sunday, 5, 30))
	case model.JobCollectNews:
		ctx := context.Background()
		settings, err := repo.PortalSettingsGet(ctx)
		if err == nil && settings.AutomationEnabled {
			scheduleIfMissing(repo, jobType, time.Now().Add(automationInterval(settings)))
		}
	}
}

func automationInterval(settings model.PortalSettings) time.Duration {
	minutes := settings.AutomationIntervalMinutes
	if minutes < 5 {
		minutes = 5
	}
	return time.Duration(minutes) * time.Minute
}

func scheduleIfMissing(repo *repository.Repository, jobType model.JobType, runAt time.Time) {
	ctx := context.Background()
	exists, err := repo.JobHasActiveType(ctx, jobType)
	if err != nil {
		log.Printf("Erro ao verificar job recorrente %s: %v", jobType, err)
		return
	}
	if exists {
		return
	}
	if err := repo.JobCreate(ctx, &model.Job{
		Type:        jobType,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       runAt,
		MaxAttempts: 3,
	}); err != nil {
		log.Printf("Erro ao agendar job recorrente %s: %v", jobType, err)
	}
}

func nextDaily(hour, minute int) time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func nextWeekly(weekday time.Weekday, hour, minute int) time.Time {
	next := nextDaily(hour, minute)
	for next.Weekday() != weekday {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func executeJob(repo *repository.Repository, cfg *config.Config, job model.Job) error {
	ctx := context.Background()

	switch job.Type {
	case model.JobPublishPost:
		var payload struct {
			PostID int64 `json:"post_id"`
		}
		if err := parsePayload(job.Payload, &payload); err != nil {
			return err
		}
		return repo.PostSetPublished(ctx, payload.PostID, time.Now())

	case model.JobExpirePromotion:
		var payload struct {
			PromotionID int64 `json:"promotion_id"`
		}
		if err := parsePayload(job.Payload, &payload); err != nil {
			return err
		}
		return repo.PromotionUpdateStatus(ctx, payload.PromotionID, "expired")

	case model.JobExpireBanner:
		var payload struct {
			BannerID int64 `json:"banner_id"`
		}
		if err := parsePayload(job.Payload, &payload); err != nil {
			return err
		}
		banner, err := repo.BannerGetByID(ctx, payload.BannerID)
		if err != nil || banner == nil {
			return err
		}
		banner.Active = false
		return repo.BannerUpdate(ctx, banner)

	case model.JobGenerateSitemap:
		entries, err := repo.SitemapEntries(ctx)
		if err != nil {
			return err
		}
		return sitemap.WriteFile(cfg.SiteURL, cfg.StaticDir, entries)

	case model.JobCleanupOldJobs:
		return repo.JobCleanupCompleted(ctx, 30)

	case model.JobVacuumDB:
		_, err := repo.DB().ExecContext(ctx, "VACUUM")
		return err

	case model.JobCompressOldUploads:
		provider := storage.NewLocalProvider(cfg.UploadDir, "")
		_, err := provider.CleanupOriginals(ctx, time.Duration(cfg.OriginalRetentionDays)*24*time.Hour)
		return err

	case model.JobCollectNews:
		settings, err := repo.PortalSettingsGet(ctx)
		if err != nil {
			return err
		}
		if !settings.AutomationEnabled {
			return nil
		}
		_, err = automation.NewService(repo).RunAllActive(ctx)
		return err

	case model.JobBackupDatabase:
		return runBackup(ctx, cfg)

	default:
		return fmt.Errorf("tipo de job desconhecido: %s", job.Type)
	}
}

func runBackup(ctx context.Context, cfg *config.Config) error {
	if cfg.BackupScript == "" {
		return fmt.Errorf("BACKUP_SCRIPT nao configurado")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, cfg.BackupScript)
	cmd.Env = append(os.Environ(),
		"DATABASE_URL="+cfg.DatabaseURL,
		"UPLOAD_DIR="+cfg.UploadDir,
	)
	if cfg.BackupDir != "" {
		cmd.Env = append(cmd.Env, "BACKUP_DIR="+cfg.BackupDir)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("backup falhou: %w: %s", err, string(out))
	}
	return nil
}

func parsePayload(payload string, v interface{}) error {
	return json.Unmarshal([]byte(payload), v)
}
