package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	tunerReady = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_ready",
		Help: "tuner server accepting connections",
	})

	tunerEpoch = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_epoch",
		Help: "tuning epoch of the last successful save, 0 before the first epoch",
	})
	tunerMSE = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_mse",
		Help: "mean squared error of the last successful save",
	})
	tunerEpochsCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tuner_epochs_completed_total",
		Help: "successful epoch saves",
	})
	tunerEpochDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_epoch_duration_seconds",
		Help: "wall-clock duration of the last successful epoch",
	})
	tunerBatchDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_batch_duration_seconds",
		Help: "wall-clock duration of the last completed batch",
	})
	tunerLearningRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_learning_rate",
		Help: "learning rate used for the next epoch",
	})
	tunerK = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_k",
		Help: "sigmoid scaling constant",
	})
	tunerPositionsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tuner_positions_processed_total",
		Help: "EPD positions processed in completed batches",
	})
	tunerChunksCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tuner_chunks_completed_total",
		Help: "chunks completed with a result",
	})
	tunerJobsScheduled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tuner_jobs_scheduled_total",
		Help: "jobs scheduled for clients",
	})
	tunerSchedulerJobTimeAvg = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_scheduler_job_time_avg_seconds",
		Help: "running average job time used to calculate the job TTL",
	})
	tunerBatchesPerEpoch = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_batches_per_epoch",
		Help: "batches in one epoch",
	})
	tunerEpochProgressRatio = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_epoch_progress_ratio",
		Help: "fraction of the current epoch's batches completed, 0..1",
	})
	tunerCoeffsChanged = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_coeffs_changed",
		Help: "tuned coefficients whose rounded value changed in the last successful epoch",
	})
	tunerCoeffsChangedRatio = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_coeffs_changed_ratio",
		Help: "fraction of tuned coefficients whose rounded value changed in the last successful epoch, 0..1",
	})
	tunerCoeffsStableEpochs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_coeffs_stable_epochs",
		Help: "consecutive successful epochs with no tuned coefficient change",
	})
	tunerTunedCoeffs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_tuned_coeffs",
		Help: "number of tuned coefficients",
	})

	tunerMSEInitial = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_mse_initial",
		Help: "mean squared error at the initial coefficients",
	})
	tunerMSEBest = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_mse_best",
		Help: "best mean squared error seen so far",
	})
	tunerMSEDelta = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_mse_delta",
		Help: "mean squared error change from the previous epoch, negative is improvement",
	})
	tunerMSERelativeDelta = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_mse_relative_delta",
		Help: "relative mean squared error change from the previous epoch, negative is improvement",
	})
	tunerEpochsSinceImprovement = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tuner_epochs_since_improvement",
		Help: "epochs since the best mean squared error was observed",
	})
)
