package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-modular-monolith/internal/shared/worker"
	"go-modular-monolith/internal/shared/worker/mocks"
)

func TestTaskPayload(t *testing.T) {
	payload := worker.TaskPayload{
		"id":    "123",
		"name":  "test",
		"count": 5,
	}

	assert.Equal(t, "123", payload["id"])
	assert.Equal(t, "test", payload["name"])
	assert.Equal(t, 5, payload["count"])
}

func TestTaskPayloadEmpty(t *testing.T) {
	payload := worker.TaskPayload{}
	assert.Empty(t, payload)
}

func TestTaskDefinition(t *testing.T) {
	handler := func(ctx context.Context, payload worker.TaskPayload) error {
		return nil
	}

	def := worker.TaskDefinition{
		TaskName: "test_task",
		Handler:  handler,
	}

	assert.Equal(t, "test_task", def.TaskName)
	assert.NotNil(t, def.Handler)
}

func TestTaskDefinitionExecution(t *testing.T) {
	handler := func(ctx context.Context, payload worker.TaskPayload) error {
		return nil
	}

	def := worker.TaskDefinition{
		TaskName: "test_task",
		Handler:  handler,
	}

	err := def.Handler(context.Background(), worker.TaskPayload{"key": "value"})
	assert.NoError(t, err)
}

func TestCronJobDefinition(t *testing.T) {
	cronJob := worker.CronJobDefinition{
		JobID:          "job-1",
		TaskName:       "task-1",
		CronExpression: worker.Daily(10, 30),
		Payload: map[string]interface{}{
			"data": "test",
		},
	}

	assert.Equal(t, "job-1", cronJob.JobID)
	assert.Equal(t, "task-1", cronJob.TaskName)
	assert.Equal(t, 30, cronJob.CronExpression.Minute)
	assert.Equal(t, 10, cronJob.CronExpression.Hour)
}

func TestCronExpressionEveryMinute(t *testing.T) {
	expr := worker.EveryMinute()
	assert.Equal(t, -1, expr.Minute)
	assert.Equal(t, -1, expr.Hour)
	assert.Equal(t, -1, expr.Day)
	assert.Equal(t, -1, expr.Month)
	assert.Equal(t, -1, expr.Weekday)
}

func TestCronExpressionEveryHour(t *testing.T) {
	expr := worker.EveryHour()
	assert.Equal(t, 0, expr.Minute)
	assert.Equal(t, -1, expr.Hour)
	assert.Equal(t, -1, expr.Day)
	assert.Equal(t, -1, expr.Month)
	assert.Equal(t, -1, expr.Weekday)
}

func TestCronExpressionDaily(t *testing.T) {
	expr := worker.Daily(14, 30)
	assert.Equal(t, 30, expr.Minute)
	assert.Equal(t, 14, expr.Hour)
	assert.Equal(t, -1, expr.Day)
	assert.Equal(t, -1, expr.Month)
	assert.Equal(t, -1, expr.Weekday)
}

func TestCronExpressionWeekly(t *testing.T) {
	expr := worker.Weekly(3, 9, 0)
	assert.Equal(t, 0, expr.Minute)
	assert.Equal(t, 9, expr.Hour)
	assert.Equal(t, -1, expr.Day)
	assert.Equal(t, -1, expr.Month)
	assert.Equal(t, 3, expr.Weekday)
}

func TestCronExpressionMonthly(t *testing.T) {
	expr := worker.Monthly(15, 12, 0)
	assert.Equal(t, 0, expr.Minute)
	assert.Equal(t, 12, expr.Hour)
	assert.Equal(t, 15, expr.Day)
	assert.Equal(t, -1, expr.Month)
	assert.Equal(t, -1, expr.Weekday)
}

func TestPriorityOption(t *testing.T) {
	opt := worker.NewPriorityOption(10)
	assert.Equal(t, 10, opt.Priority)
}

func TestMaxRetriesOption(t *testing.T) {
	opt := worker.NewMaxRetriesOption(3)
	assert.Equal(t, 3, opt.MaxRetries)
}

func TestTimeoutOption(t *testing.T) {
	timeout := 5 * time.Second
	opt := worker.NewTimeoutOption(timeout)
	assert.Equal(t, timeout, opt.Timeout)
}

func TestQueueOption(t *testing.T) {
	opt := worker.NewQueueOption("priority")
	assert.Equal(t, "priority", opt.Queue)
}

func TestMockClientEnqueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	ctx := context.Background()

	mockClient.EXPECT().
		Enqueue(ctx, "task1", gomock.Any()).
		Return(nil).
		Times(1)

	err := mockClient.Enqueue(ctx, "task1", worker.TaskPayload{"key": "value"})
	require.NoError(t, err)
}

func TestMockClientEnqueueMultiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	ctx := context.Background()

	gomock.InOrder(
		mockClient.EXPECT().Enqueue(ctx, "task1", gomock.Any()).Return(nil),
		mockClient.EXPECT().Enqueue(ctx, "task2", gomock.Any()).Return(nil),
		mockClient.EXPECT().Enqueue(ctx, "task3", gomock.Any()).Return(nil),
	)

	mockClient.Enqueue(ctx, "task1", worker.TaskPayload{})
	mockClient.Enqueue(ctx, "task2", worker.TaskPayload{})
	mockClient.Enqueue(ctx, "task3", worker.TaskPayload{})
}

func TestMockClientEnqueueDelayed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	ctx := context.Background()

	mockClient.EXPECT().
		EnqueueDelayed(ctx, "delayed_task", gomock.Any(), 5*time.Second).
		Return(nil).
		Times(1)

	err := mockClient.EnqueueDelayed(ctx, "delayed_task", worker.TaskPayload{}, 5*time.Second)
	require.NoError(t, err)
}

func TestMockClientClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	mockClient.EXPECT().
		Close().
		Return(nil).
		Times(1)

	err := mockClient.Close()
	require.NoError(t, err)
}

func TestMockServerRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := mocks.NewMockServer(ctrl)

	handler := func(ctx context.Context, payload worker.TaskPayload) error {
		return nil
	}

	mockServer.EXPECT().
		RegisterHandler("task1", gomock.Any()).
		Return(nil).
		Times(1)

	err := mockServer.RegisterHandler("task1", handler)
	require.NoError(t, err)
}

func TestMockServerMultipleHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := mocks.NewMockServer(ctrl)

	handler1 := func(ctx context.Context, payload worker.TaskPayload) error { return nil }
	handler2 := func(ctx context.Context, payload worker.TaskPayload) error { return nil }

	gomock.InOrder(
		mockServer.EXPECT().RegisterHandler("task1", gomock.Any()).Return(nil),
		mockServer.EXPECT().RegisterHandler("task2", gomock.Any()).Return(nil),
	)

	mockServer.RegisterHandler("task1", handler1)
	mockServer.RegisterHandler("task2", handler2)
}

func TestMockServerStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := mocks.NewMockServer(ctrl)

	gomock.InOrder(
		mockServer.EXPECT().Start(context.Background()).Return(nil),
		mockServer.EXPECT().Stop(context.Background()).Return(nil),
	)

	err := mockServer.Start(context.Background())
	require.NoError(t, err)

	err = mockServer.Stop(context.Background())
	require.NoError(t, err)
}

func TestMockSchedulerAddJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockScheduler := mocks.NewMockScheduler(ctrl)

	mockScheduler.EXPECT().
		AddJob("job1", "task1", gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	err := mockScheduler.AddJob("job1", "task1", worker.Daily(10, 0), worker.TaskPayload{"data": "test"})
	require.NoError(t, err)
}

func TestMockSchedulerRemoveJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockScheduler := mocks.NewMockScheduler(ctrl)

	mockScheduler.EXPECT().
		RemoveJob("job1").
		Return(nil).
		Times(1)

	err := mockScheduler.RemoveJob("job1")
	require.NoError(t, err)
}

func TestMockSchedulerEnableDisable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockScheduler := mocks.NewMockScheduler(ctrl)

	gomock.InOrder(
		mockScheduler.EXPECT().EnableJob("job1").Return(nil),
		mockScheduler.EXPECT().DisableJob("job1").Return(nil),
	)

	err := mockScheduler.EnableJob("job1")
	require.NoError(t, err)

	err = mockScheduler.DisableJob("job1")
	require.NoError(t, err)
}

func TestMockSchedulerStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockScheduler := mocks.NewMockScheduler(ctrl)

	gomock.InOrder(
		mockScheduler.EXPECT().Start(context.Background()).Return(nil),
		mockScheduler.EXPECT().Stop().Return(nil),
	)

	err := mockScheduler.Start(context.Background())
	require.NoError(t, err)

	err = mockScheduler.Stop()
	require.NoError(t, err)
}

func TestClientInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	var _ worker.Client = mockClient
}

func TestServerInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := mocks.NewMockServer(ctrl)
	var _ worker.Server = mockServer
}

func TestSchedulerInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockScheduler := mocks.NewMockScheduler(ctrl)
	var _ worker.Scheduler = mockScheduler
}
