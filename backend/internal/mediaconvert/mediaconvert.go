package mediaconvert

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsmc "github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	awsmctypes "github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
)

const (
	defaultPollInterval              = 5 * time.Second
	defaultMaterializeTimeout        = 30 * time.Minute
	defaultPosterFrameSequence       = ".0000000.jpg"
	defaultH264MaxBitrate      int32 = 5_000_000
	defaultAACBitrate          int32 = 96_000
	defaultAACSampleRate       int32 = 48_000
	defaultQvbrQuality         int32 = 7
)

// Config は MediaConvert client の最小設定を表します。
type Config struct {
	Region string
}

// Validate は MediaConvert 設定の整合性を検証します。
func (c Config) Validate() error {
	if strings.TrimSpace(c.Region) == "" {
		return fmt.Errorf("region is required")
	}

	return nil
}

type api interface {
	ListQueues(ctx context.Context, params *awsmc.ListQueuesInput, optFns ...func(*awsmc.Options)) (*awsmc.ListQueuesOutput, error)
	CreateJob(ctx context.Context, params *awsmc.CreateJobInput, optFns ...func(*awsmc.Options)) (*awsmc.CreateJobOutput, error)
	GetJob(ctx context.Context, params *awsmc.GetJobInput, optFns ...func(*awsmc.Options)) (*awsmc.GetJobOutput, error)
}

// Client は MediaConvert access check と materialization job 実行を包みます。
type Client struct {
	api          api
	pollInterval time.Duration
	maxJobWait   time.Duration
}

// MaterializeRequest は raw video を delivery-ready asset へ変換する入力です。
type MaterializeRequest struct {
	InputBucket    string
	InputKey       string
	OutputBucket   string
	PlaybackKey    string
	PosterBaseKey  string
	ServiceRoleARN string
}

// MaterializeResult は MediaConvert materialization の結果です。
type MaterializeResult struct {
	JobID           string
	PlaybackKey     string
	PosterSourceKey string
	DurationMS      int64
}

// JobError は MediaConvert job の失敗を retryability つきで表します。
type JobError struct {
	Code      string
	Message   string
	Retryable bool
}

func (e *JobError) Error() string {
	if e == nil {
		return ""
	}

	if e.Message == "" {
		return e.Code
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewClient は AWS SDK を使って MediaConvert client を構築します。
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return newClient(awsmc.NewFromConfig(awsCfg)), nil
}

func newClient(api api) *Client {
	return &Client{
		api:          api,
		pollInterval: defaultPollInterval,
		maxJobWait:   defaultMaterializeTimeout,
	}
}

// CheckAccess は queue 一覧取得で MediaConvert への接続前提を検証し、代表 queue 名を返します。
func (c *Client) CheckAccess(ctx context.Context) (string, error) {
	if c == nil {
		return "", fmt.Errorf("mediaconvert client is nil")
	}

	output, err := c.api.ListQueues(ctx, &awsmc.ListQueuesInput{})
	if err != nil {
		return "", fmt.Errorf("list mediaconvert queues: %w", err)
	}
	if len(output.Queues) == 0 {
		return "", fmt.Errorf("list mediaconvert queues returned no queues")
	}
	if output.Queues[0].Name == nil {
		return "", fmt.Errorf("mediaconvert queue name is empty")
	}

	queueName := strings.TrimSpace(*output.Queues[0].Name)
	if queueName == "" {
		return "", fmt.Errorf("mediaconvert queue name is empty")
	}

	return queueName, nil
}

// MaterializeVideo は raw video を mp4 / poster 出力へ変換し、完了まで待機します。
func (c *Client) MaterializeVideo(ctx context.Context, req MaterializeRequest) (MaterializeResult, error) {
	if c == nil {
		return MaterializeResult{}, fmt.Errorf("mediaconvert client is nil")
	}
	if err := req.validate(); err != nil {
		return MaterializeResult{}, err
	}

	createOutput, err := c.api.CreateJob(ctx, buildCreateJobInput(req))
	if err != nil {
		return MaterializeResult{}, &JobError{
			Code:      "mediaconvert_submit_failed",
			Message:   err.Error(),
			Retryable: true,
		}
	}
	if createOutput.Job == nil || createOutput.Job.Id == nil {
		return MaterializeResult{}, fmt.Errorf("mediaconvert create job returned no job id")
	}

	jobID := strings.TrimSpace(*createOutput.Job.Id)
	if jobID == "" {
		return MaterializeResult{}, fmt.Errorf("mediaconvert create job returned empty job id")
	}

	waitCtx, cancel, deadlineAdded := materializeContext(ctx, c.maxJobWait)
	defer cancel()

	job, err := c.waitForJob(waitCtx, jobID)
	if err != nil {
		if deadlineAdded && errors.Is(err, context.DeadlineExceeded) {
			return MaterializeResult{}, &JobError{
				Code:      "mediaconvert_wait_timeout",
				Message:   fmt.Sprintf("job %s did not complete within %s", jobID, c.maxJobWait),
				Retryable: true,
			}
		}
		return MaterializeResult{}, err
	}

	durationMS, err := extractDurationMS(job)
	if err != nil {
		return MaterializeResult{}, err
	}

	return MaterializeResult{
		JobID:           jobID,
		PlaybackKey:     strings.TrimSpace(req.PlaybackKey),
		PosterSourceKey: strings.TrimSpace(req.PosterBaseKey) + defaultPosterFrameSequence,
		DurationMS:      durationMS,
	}, nil
}

func (c *Client) waitForJob(ctx context.Context, jobID string) (*awsmctypes.Job, error) {
	for {
		output, err := c.api.GetJob(ctx, &awsmc.GetJobInput{Id: aws.String(jobID)})
		if err != nil {
			return nil, &JobError{
				Code:      "mediaconvert_get_job_failed",
				Message:   err.Error(),
				Retryable: true,
			}
		}
		if output.Job == nil {
			return nil, fmt.Errorf("mediaconvert get job returned no job")
		}

		switch output.Job.Status {
		case awsmctypes.JobStatusSubmitted, awsmctypes.JobStatusProgressing:
			if err := sleepWithContext(ctx, c.pollInterval); err != nil {
				return nil, err
			}
		case awsmctypes.JobStatusComplete:
			return output.Job, nil
		case awsmctypes.JobStatusCanceled:
			return nil, jobFailure(output.Job, "mediaconvert_job_canceled", false)
		case awsmctypes.JobStatusError:
			return nil, jobFailure(output.Job, "mediaconvert_job_error", true)
		default:
			return nil, fmt.Errorf("unexpected mediaconvert job status: %s", output.Job.Status)
		}
	}
}

func buildCreateJobInput(req MaterializeRequest) *awsmc.CreateJobInput {
	return &awsmc.CreateJobInput{
		Role: aws.String(strings.TrimSpace(req.ServiceRoleARN)),
		Settings: &awsmctypes.JobSettings{
			Inputs: []awsmctypes.Input{
				{
					FileInput: aws.String(s3URL(req.InputBucket, req.InputKey)),
					AudioSelectors: map[string]awsmctypes.AudioSelector{
						"Audio Selector 1": {
							DefaultSelection: awsmctypes.AudioDefaultSelectionDefault,
						},
					},
				},
			},
			OutputGroups: []awsmctypes.OutputGroup{
				buildMP4OutputGroup(req),
				buildPosterOutputGroup(req),
			},
		},
	}
}

func materializeContext(ctx context.Context, maxWait time.Duration) (context.Context, context.CancelFunc, bool) {
	if maxWait <= 0 {
		return ctx, func() {}, false
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}, false
	}

	derived, cancel := context.WithTimeout(ctx, maxWait)
	return derived, cancel, true
}

func buildMP4OutputGroup(req MaterializeRequest) awsmctypes.OutputGroup {
	return awsmctypes.OutputGroup{
		OutputGroupSettings: &awsmctypes.OutputGroupSettings{
			Type: awsmctypes.OutputGroupTypeFileGroupSettings,
			FileGroupSettings: &awsmctypes.FileGroupSettings{
				Destination: aws.String(s3URL(req.OutputBucket, strings.TrimSuffix(strings.TrimSpace(req.PlaybackKey), ".mp4"))),
			},
		},
		Outputs: []awsmctypes.Output{
			{
				ContainerSettings: &awsmctypes.ContainerSettings{
					Container: awsmctypes.ContainerTypeMp4,
					Mp4Settings: &awsmctypes.Mp4Settings{
						MoovPlacement: awsmctypes.Mp4MoovPlacementProgressiveDownload,
					},
				},
				AudioDescriptions: []awsmctypes.AudioDescription{
					{
						AudioSourceName: aws.String("Audio Selector 1"),
						CodecSettings: &awsmctypes.AudioCodecSettings{
							Codec: awsmctypes.AudioCodecAac,
							AacSettings: &awsmctypes.AacSettings{
								Bitrate:    aws.Int32(defaultAACBitrate),
								CodingMode: awsmctypes.AacCodingModeCodingMode20,
								SampleRate: aws.Int32(defaultAACSampleRate),
							},
						},
					},
				},
				VideoDescription: &awsmctypes.VideoDescription{
					CodecSettings: &awsmctypes.VideoCodecSettings{
						Codec: awsmctypes.VideoCodecH264,
						H264Settings: &awsmctypes.H264Settings{
							MaxBitrate:         aws.Int32(defaultH264MaxBitrate),
							QualityTuningLevel: awsmctypes.H264QualityTuningLevelSinglePass,
							RateControlMode:    awsmctypes.H264RateControlModeQvbr,
							SceneChangeDetect:  awsmctypes.H264SceneChangeDetectTransitionDetection,
							QvbrSettings: &awsmctypes.H264QvbrSettings{
								QvbrQualityLevel: aws.Int32(defaultQvbrQuality),
							},
						},
					},
				},
			},
		},
	}
}

func buildPosterOutputGroup(req MaterializeRequest) awsmctypes.OutputGroup {
	return awsmctypes.OutputGroup{
		OutputGroupSettings: &awsmctypes.OutputGroupSettings{
			Type: awsmctypes.OutputGroupTypeFileGroupSettings,
			FileGroupSettings: &awsmctypes.FileGroupSettings{
				Destination: aws.String(s3URL(req.OutputBucket, strings.TrimSpace(req.PosterBaseKey))),
			},
		},
		Outputs: []awsmctypes.Output{
			{
				ContainerSettings: &awsmctypes.ContainerSettings{
					Container: awsmctypes.ContainerTypeRaw,
				},
				VideoDescription: &awsmctypes.VideoDescription{
					CodecSettings: &awsmctypes.VideoCodecSettings{
						Codec: awsmctypes.VideoCodecFrameCapture,
						FrameCaptureSettings: &awsmctypes.FrameCaptureSettings{
							FramerateNumerator:   aws.Int32(1),
							FramerateDenominator: aws.Int32(1),
							MaxCaptures:          aws.Int32(1),
							Quality:              aws.Int32(80),
						},
					},
				},
			},
		},
	}
}

func extractDurationMS(job *awsmctypes.Job) (int64, error) {
	if job == nil {
		return 0, fmt.Errorf("mediaconvert job is nil")
	}

	for _, outputGroup := range job.OutputGroupDetails {
		for _, outputDetail := range outputGroup.OutputDetails {
			if outputDetail.DurationInMs == nil {
				continue
			}
			return int64(*outputDetail.DurationInMs), nil
		}
	}

	return 0, fmt.Errorf("mediaconvert job returned no duration")
}

func jobFailure(job *awsmctypes.Job, fallbackCode string, retryable bool) error {
	if job == nil {
		return &JobError{
			Code:      fallbackCode,
			Message:   "job is nil",
			Retryable: retryable,
		}
	}

	code := fallbackCode
	if job.ErrorCode != nil {
		code = strconv.Itoa(int(*job.ErrorCode))
	}

	return &JobError{
		Code:      code,
		Message:   strings.TrimSpace(aws.ToString(job.ErrorMessage)),
		Retryable: retryable,
	}
}

func sleepWithContext(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (r MaterializeRequest) validate() error {
	switch {
	case strings.TrimSpace(r.InputBucket) == "":
		return fmt.Errorf("input bucket is required")
	case strings.TrimSpace(r.InputKey) == "":
		return fmt.Errorf("input key is required")
	case strings.TrimSpace(r.OutputBucket) == "":
		return fmt.Errorf("output bucket is required")
	case strings.TrimSpace(r.PlaybackKey) == "":
		return fmt.Errorf("playback key is required")
	case strings.TrimSpace(r.PosterBaseKey) == "":
		return fmt.Errorf("poster base key is required")
	case strings.TrimSpace(r.ServiceRoleARN) == "":
		return fmt.Errorf("mediaconvert service role arn is required")
	default:
		return nil
	}
}

func s3URL(bucket string, key string) string {
	return fmt.Sprintf("s3://%s/%s", strings.TrimSpace(bucket), strings.TrimLeft(strings.TrimSpace(key), "/"))
}
