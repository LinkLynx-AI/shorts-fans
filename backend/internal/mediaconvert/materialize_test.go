package mediaconvert

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmc "github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	awsmctypes "github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
)

type materializeAPIStub struct {
	createInput  *awsmc.CreateJobInput
	createOutput *awsmc.CreateJobOutput
	createErr    error
	getOutputs   []*awsmc.GetJobOutput
	getErr       error
	getCalls     int
}

func (s *materializeAPIStub) ListQueues(context.Context, *awsmc.ListQueuesInput, ...func(*awsmc.Options)) (*awsmc.ListQueuesOutput, error) {
	return &awsmc.ListQueuesOutput{}, nil
}

func (s *materializeAPIStub) CreateJob(_ context.Context, params *awsmc.CreateJobInput, _ ...func(*awsmc.Options)) (*awsmc.CreateJobOutput, error) {
	s.createInput = params
	if s.createErr != nil {
		return nil, s.createErr
	}
	if s.createOutput == nil {
		return &awsmc.CreateJobOutput{}, nil
	}
	return s.createOutput, nil
}

func (s *materializeAPIStub) GetJob(_ context.Context, params *awsmc.GetJobInput, _ ...func(*awsmc.Options)) (*awsmc.GetJobOutput, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if len(s.getOutputs) == 0 {
		return &awsmc.GetJobOutput{}, nil
	}
	index := s.getCalls
	if index >= len(s.getOutputs) {
		index = len(s.getOutputs) - 1
	}
	s.getCalls++
	if got, want := aws.ToString(params.Id), "job-123"; got != want {
		return nil, errors.New("unexpected job id")
	}
	return s.getOutputs[index], nil
}

func TestMaterializeVideoSuccess(t *testing.T) {
	t.Parallel()

	api := &materializeAPIStub{
		createOutput: &awsmc.CreateJobOutput{
			Job: &awsmctypes.Job{Id: aws.String("job-123")},
		},
		getOutputs: []*awsmc.GetJobOutput{
			{Job: &awsmctypes.Job{Status: awsmctypes.JobStatusSubmitted}},
			{Job: &awsmctypes.Job{
				Status: awsmctypes.JobStatusComplete,
				OutputGroupDetails: []awsmctypes.OutputGroupDetail{{
					OutputDetails: []awsmctypes.OutputDetail{{
						DurationInMs: aws.Int32(42000),
					}},
				}},
			}},
		},
	}
	client := newClient(api)
	client.pollInterval = time.Nanosecond

	result, err := client.MaterializeVideo(context.Background(), MaterializeRequest{
		InputBucket:    "raw-bucket",
		InputKey:       "raw/input.mp4",
		OutputBucket:   "delivery-bucket",
		PlaybackKey:    "shorts/abc/playback.mp4",
		PosterBaseKey:  "shorts/abc/poster-temp",
		ServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	})
	if err != nil {
		t.Fatalf("MaterializeVideo() error = %v, want nil", err)
	}
	if got, want := result.JobID, "job-123"; got != want {
		t.Fatalf("MaterializeVideo() job id got %q want %q", got, want)
	}
	if got, want := result.PlaybackKey, "shorts/abc/playback.mp4"; got != want {
		t.Fatalf("MaterializeVideo() playback key got %q want %q", got, want)
	}
	if got, want := result.PosterSourceKey, "shorts/abc/poster-temp.0000000.jpg"; got != want {
		t.Fatalf("MaterializeVideo() poster key got %q want %q", got, want)
	}
	if got, want := result.DurationMS, int64(42000); got != want {
		t.Fatalf("MaterializeVideo() duration got %d want %d", got, want)
	}
	if api.createInput == nil {
		t.Fatal("MaterializeVideo() create job input = nil")
	}
	if got, want := aws.ToString(api.createInput.Role), "arn:aws:iam::123456789012:role/media-role"; got != want {
		t.Fatalf("MaterializeVideo() role got %q want %q", got, want)
	}
	if got, want := aws.ToString(api.createInput.Settings.Inputs[0].FileInput), "s3://raw-bucket/raw/input.mp4"; got != want {
		t.Fatalf("MaterializeVideo() input file got %q want %q", got, want)
	}
	if got, want := aws.ToString(api.createInput.Settings.OutputGroups[0].OutputGroupSettings.FileGroupSettings.Destination), "s3://delivery-bucket/shorts/abc/playback"; got != want {
		t.Fatalf("MaterializeVideo() playback destination got %q want %q", got, want)
	}
	if got, want := aws.ToString(api.createInput.Settings.OutputGroups[1].OutputGroupSettings.FileGroupSettings.Destination), "s3://delivery-bucket/shorts/abc/poster-temp"; got != want {
		t.Fatalf("MaterializeVideo() poster destination got %q want %q", got, want)
	}
}

func TestMaterializeVideoSubmitFailure(t *testing.T) {
	t.Parallel()

	client := newClient(&materializeAPIStub{createErr: errors.New("submit failed")})
	_, err := client.MaterializeVideo(context.Background(), MaterializeRequest{
		InputBucket:    "raw-bucket",
		InputKey:       "raw/input.mp4",
		OutputBucket:   "delivery-bucket",
		PlaybackKey:    "shorts/abc/playback.mp4",
		PosterBaseKey:  "shorts/abc/poster-temp",
		ServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	})

	var jobErr *JobError
	if !errors.As(err, &jobErr) {
		t.Fatalf("MaterializeVideo() error got %T want *JobError", err)
	}
	if got, want := jobErr.Code, "mediaconvert_submit_failed"; got != want {
		t.Fatalf("MaterializeVideo() code got %q want %q", got, want)
	}
	if !jobErr.Retryable {
		t.Fatal("MaterializeVideo() retryable = false, want true")
	}
}

func TestMaterializeVideoTimesOutWithoutCallerDeadline(t *testing.T) {
	t.Parallel()

	client := newClient(&materializeAPIStub{
		createOutput: &awsmc.CreateJobOutput{
			Job: &awsmctypes.Job{Id: aws.String("job-123")},
		},
		getOutputs: []*awsmc.GetJobOutput{{
			Job: &awsmctypes.Job{Status: awsmctypes.JobStatusProgressing},
		}},
	})
	client.pollInterval = time.Hour
	client.maxJobWait = time.Nanosecond

	_, err := client.MaterializeVideo(context.Background(), MaterializeRequest{
		InputBucket:    "raw-bucket",
		InputKey:       "raw/input.mp4",
		OutputBucket:   "delivery-bucket",
		PlaybackKey:    "shorts/abc/playback.mp4",
		PosterBaseKey:  "shorts/abc/poster-temp",
		ServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	})

	var jobErr *JobError
	if !errors.As(err, &jobErr) {
		t.Fatalf("MaterializeVideo() error got %T want *JobError", err)
	}
	if got, want := jobErr.Code, "mediaconvert_wait_timeout"; got != want {
		t.Fatalf("MaterializeVideo() code got %q want %q", got, want)
	}
	if !jobErr.Retryable {
		t.Fatal("MaterializeVideo() retryable = false, want true")
	}
}

func TestWaitForJobErrorStatus(t *testing.T) {
	t.Parallel()

	client := newClient(&materializeAPIStub{
		getOutputs: []*awsmc.GetJobOutput{{
			Job: &awsmctypes.Job{
				Status:       awsmctypes.JobStatusError,
				ErrorCode:    aws.Int32(1234),
				ErrorMessage: aws.String("transcode failed"),
			},
		}},
	})

	_, err := client.waitForJob(context.Background(), "job-123")
	var jobErr *JobError
	if !errors.As(err, &jobErr) {
		t.Fatalf("waitForJob() error got %T want *JobError", err)
	}
	if got, want := jobErr.Code, "1234"; got != want {
		t.Fatalf("waitForJob() code got %q want %q", got, want)
	}
	if got, want := jobErr.Message, "transcode failed"; got != want {
		t.Fatalf("waitForJob() message got %q want %q", got, want)
	}
	if !jobErr.Retryable {
		t.Fatal("waitForJob() retryable = false, want true")
	}
}

func TestWaitForJobGetFailure(t *testing.T) {
	t.Parallel()

	client := newClient(&materializeAPIStub{getErr: errors.New("get failed")})
	_, err := client.waitForJob(context.Background(), "job-123")

	var jobErr *JobError
	if !errors.As(err, &jobErr) {
		t.Fatalf("waitForJob() error got %T want *JobError", err)
	}
	if got, want := jobErr.Code, "mediaconvert_get_job_failed"; got != want {
		t.Fatalf("waitForJob() code got %q want %q", got, want)
	}
}

func TestHelpers(t *testing.T) {
	t.Parallel()

	if got, want := (&JobError{Code: "code", Message: "detail"}).Error(), "code: detail"; got != want {
		t.Fatalf("JobError.Error() got %q want %q", got, want)
	}
	if got, want := (&JobError{Code: "code"}).Error(), "code"; got != want {
		t.Fatalf("JobError.Error() got %q want %q", got, want)
	}

	if _, err := extractDurationMS(&awsmctypes.Job{}); err == nil {
		t.Fatal("extractDurationMS() error = nil, want error")
	}
	if _, err := extractDurationMS(nil); err == nil {
		t.Fatal("extractDurationMS(nil) error = nil, want error")
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleepWithContext(cancelled, time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatalf("sleepWithContext() error got %v want context.Canceled", err)
	}

	if err := (MaterializeRequest{}).validate(); err == nil {
		t.Fatal("MaterializeRequest.validate() error = nil, want error")
	}
	if got, want := s3URL("bucket", "/key"), "s3://bucket/key"; got != want {
		t.Fatalf("s3URL() got %q want %q", got, want)
	}

	derived, cancel, added := materializeContext(context.Background(), time.Minute)
	defer cancel()
	if !added {
		t.Fatal("materializeContext() added got false want true")
	}
	if _, ok := derived.Deadline(); !ok {
		t.Fatal("materializeContext() deadline missing, want derived deadline")
	}

	parent, parentCancel := context.WithTimeout(context.Background(), time.Minute)
	defer parentCancel()
	preserved, stop, added := materializeContext(parent, time.Second)
	defer stop()
	if added {
		t.Fatal("materializeContext() added got true want false when parent already has deadline")
	}
	parentDeadline, _ := parent.Deadline()
	preservedDeadline, _ := preserved.Deadline()
	if !preservedDeadline.Equal(parentDeadline) {
		t.Fatalf("materializeContext() deadline got %v want %v", preservedDeadline, parentDeadline)
	}
}
