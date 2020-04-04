package eventstore_test

import (
	"os"
	"testing"

	"github.com/Brandon2255p/watermill-eventstore/eventstore"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/require"
)

var (
	logger = watermill.NewStdLogger(true, true)
)

func newPubSub(t *testing.T) *eventstore.Publisher {
	addr := os.Getenv("WATERMILL_TEST_EVENTSTORE_HTTP_URL")
	if addr == "" {
		addr = "localhost"
	}
	conf := &eventstore.PublisherConfig{
		URL:      "http://localhost:2113",
		Username: "admin",
		Password: "changeit",
	}
	pubsub, err := eventstore.NewPublisher(*conf, logger)
	require.NoError(t, err)

	return pubsub
}

func TestES(t *testing.T) {
	newPubSub(t)
}

func TestPublishNoId(t *testing.T) {
	pubSub := newPubSub(t)
	payload := []byte("hello")
	msg := message.NewMessage("test", payload)
	err := pubSub.Publish("test", msg)
	require.EqualError(t, err, "Message is missing aggregate_id in metadata for message test")
}

func TestPublish(t *testing.T) {
	pubSub := newPubSub(t)
	payload := []byte("hello")
	msg := message.NewMessage("test", payload)
	msg.Metadata.Set("aggregate_id", "testid-1234")
	err := pubSub.Publish("test", msg)
	require.NoError(t, err)
}
