package main

import (
	"testing"

	"inhumas-em-foco/internal/model"
)

func TestJobUsesTenantContextPolicy(t *testing.T) {
	tenantJobs := []model.JobType{
		model.JobPublishPost,
		model.JobExpirePromotion,
		model.JobExpireBanner,
		model.JobCollectNews,
	}
	for _, jobType := range tenantJobs {
		if !jobUsesTenantContext(jobType) {
			t.Fatalf("%s should use tenant context", jobType)
		}
	}

	globalJobs := []model.JobType{
		model.JobBackupDatabase,
		model.JobVacuumDB,
		model.JobCleanupOldJobs,
		model.JobCompressOldUploads,
		model.JobGenerateSitemap,
	}
	for _, jobType := range globalJobs {
		if jobUsesTenantContext(jobType) {
			t.Fatalf("%s should be global", jobType)
		}
	}
}

func TestGlobalRecurringJobsPolicy(t *testing.T) {
	jobs := globalRecurringJobs()
	want := map[model.JobType]bool{
		model.JobBackupDatabase:     true,
		model.JobGenerateSitemap:    true,
		model.JobVacuumDB:           true,
		model.JobCleanupOldJobs:     true,
		model.JobCompressOldUploads: true,
	}
	if len(jobs) != len(want) {
		t.Fatalf("global recurring job count = %d, want %d", len(jobs), len(want))
	}
	for _, job := range jobs {
		if !want[job.jobType] {
			t.Fatalf("unexpected global recurring job: %s", job.jobType)
		}
		if job.runAt.IsZero() {
			t.Fatalf("global recurring job %s has zero runAt", job.jobType)
		}
	}
}
