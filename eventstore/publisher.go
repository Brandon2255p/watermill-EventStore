package eventstore

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	goes "github.com/jetbasrawi/go.geteventstore"
	"github.com/pkg/errors"
)

// PublisherConfig test
type PublisherConfig struct {
	Username string
	Password string
	URL      string
}

// Publisher inserts the Messages as events into EventStore
type Publisher struct {
	Config PublisherConfig

	publishWg *sync.WaitGroup
	closeCh   chan struct{}
	closed    bool
	client    *goes.Client
	logger    watermill.LoggerAdapter
}

// NewPublisher creates a publisher of EventStore
func NewPublisher(config PublisherConfig, logger watermill.LoggerAdapter) (*Publisher, error) {
	client, err := goes.NewClient(nil, config.URL)
	if err != nil {
		log.Fatal(err)
	}
	client.SetBasicAuth(config.Username, config.Password)
	publisher := &Publisher{
		client: client,
		Config: config,
		logger: logger,
	}
	return publisher, nil
}

// Publish inserts the messages as rows into the MessagesTable.
// Order is guaranteed for messages within each aggregate ID.
func (p *Publisher) Publish(topic string, messages ...*message.Message) (err error) {
	for _, message := range messages {
		aggregateID, hasID := message.Metadata["aggregate_id"]
		if !hasID {
			return errors.Errorf("Message is missing aggregate_id in metadata for message %s", message.UUID)
		}
		metadata, err := json.Marshal(message.Metadata)
		if err != nil {
			return errors.Wrapf(err, "could not marshal metadata into JSON for message %s", message.UUID)
		}
		newEvent := goes.NewEvent(goes.NewUUID(), "FooEvent", message.Payload, metadata)
		if err != nil {
			return errors.Wrap(err, "could not insert message as row")
		}

		writer := p.client.NewStreamWriter(aggregateID)
		p.logger.Info("Stream client created", nil)
		// Write the event to the stream, here we pass nil as the expectedVersion as we
		// are not wanting to flag concurrency errors
		err = writer.Append(nil, newEvent)
		if err != nil {
			return errors.Wrap(err, "Unable to commit event")
		}
		p.logger.Info("Event emitted", nil)
	}

	return nil
}

// Close closes the publisher, which means that all the Publish calls called before are finished
// and no more Publish calls are accepted.
// Close is blocking until all the ongoing Publish calls have returned.
func (p *Publisher) Close() error {
	if p.closed {
		return nil
	}

	p.closed = true

	close(p.closeCh)
	p.publishWg.Wait()

	return nil
}
