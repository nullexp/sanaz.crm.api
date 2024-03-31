//nolint:all
package memory

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryClient_Publish(t *testing.T) {
	ctx := context.Background()
	memoryClient := NewMemoryClient()

	sampleSubject := "sample.subject"
	sampleContent := []byte("sample-content")
	errCh := memoryClient.Publish(ctx, sampleSubject, sampleContent)
	select {
	case err := <-errCh:
		assert.Error(t, err)
	}
}

func TestMemoryClient_Subscribe(t *testing.T) {
	ctx := context.Background()
	memoryClient := NewMemoryClient()

	sampleSubject := "sample.subject"
	sampleContent := []byte("sample-content")

	// subscribe to sample.subject
	dataCh, err := memoryClient.Subscribe(ctx, sampleSubject)
	assert.NoError(t, err)

	go func() {
		for {
			select {
			case msg := <-dataCh:
				assert.Equal(t, sampleContent, msg)
			}
		}
	}()

	// publish 5 message on sample.subject
	for i := 0; i < 5; i++ {
		errCh := memoryClient.Publish(ctx, sampleSubject, sampleContent)
		select {
		case err = <-errCh:
			assert.NoError(t, err)
		case <-time.After(time.Second * 3):
			assert.FailNow(t, "publish timeout")
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func TestMemoryClient_Request(t *testing.T) {
	ctx := context.Background()
	memoryClient := NewMemoryClient()

	sampleSubject := "sample.subject"
	sampleContent := []byte("sample-content")

	replyCh := memoryClient.Request(ctx, sampleSubject, sampleContent)

	select {
	case replyMsg := <-replyCh:
		assert.Error(t, replyMsg.Error)
	}
}

func TestMemoryClient_SubscribeToRespond(t *testing.T) {
	ctx := context.Background()
	memoryClient := NewMemoryClient()

	sampleSubject := "sample.subject"
	sampleContent := []byte("sample-content")
	sampleAnswer := []byte("sample-answer")

	subscribeCh, err := memoryClient.SubscribeToRespond(ctx, sampleSubject)
	assert.NoError(t, err)

	go func() {
		for {
			select {
			case msg := <-subscribeCh:
				assert.Equal(t, msg.GetData(), sampleContent)

				// response to received data
				answerCh := msg.Answer(ctx, sampleAnswer)
				err = <-answerCh
				assert.NoError(t, err)
			}
		}
	}()

	for i := 0; i < 5; i++ {
		replyCh := memoryClient.Request(context.Background(), sampleSubject, sampleContent)
		select {
		case reply := <-replyCh:
			assert.NoError(t, reply.Error)
			assert.Equal(t, sampleAnswer, reply.Value)
			log.Println(reply.Value)
		case <-time.After(time.Second * 3):
			assert.FailNow(t, "request timeout")
		}
	}
}
