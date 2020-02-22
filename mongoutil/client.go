// Package mongoutil contains helpers to interact with mongoDB.
package mongoutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/pastequo/libs.golang.utils/logutil"
)

type Client struct {
	mongoClient *mongo.Client

	initDone chan bool
	stopInit chan bool

	stopOnce  sync.Once
	stopError error
}

// ErrClientNotConnected is the returned error when the cache is not ready (yet?).
var ErrClientNotConnected = errors.New("client was not connected")

// NewClient creates a client to a mongodb.
func NewClient(mongoURL, mongoUser, mongoPwd, mongoAuthSource string) *Client {

	logger := logutil.GetDefaultLogger()

	ret := &Client{}

	ret.initDone = make(chan bool, 1)
	ret.stopInit = make(chan bool, 1)

	waitingTime := 3 * time.Second
	maxTime := 5 * time.Minute

	go func() {
loop:
		for {
			err := ret.connect(mongoURL, mongoUser, mongoPwd, mongoAuthSource)
			if err == nil {
				logger.Info("mongodb connection successfull")
				break
			}

			logger.WithError(err).Errorf("failed to connect to mongodb, retrying in %v", waitingTime)

			select {
			case <-ret.stopInit:
				logger.Info("interrupting initialization")
				break loop
			case <-time.After(waitingTime):
			}

			waitingTime *= 2
			if waitingTime > maxTime {
				waitingTime = maxTime
			}
		}

		logger.Debug("notifying initialization is done")
		close(ret.initDone)
	}()

	return ret
}

// GetCollection returns the mongo.Collection with the given name, inside livinglab_config db.
func (c *Client) GetCollection(ctx context.Context, dbName, collectionName string) (*mongo.Collection, error) {

	select {
	case <-c.initDone:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if c.mongoClient == nil {
		return nil, ErrClientNotConnected
	}

	db := c.mongoClient.Database(dbName)
	if db == nil {
		return nil, fmt.Errorf("failed to get %v database", dbName)
	}

	collection := db.Collection(collectionName)
	if collection == nil {
		return nil, fmt.Errorf("nil collection %v", collectionName)
	}

	return collection, nil
}

// Close closes the connection to the mongodb.
func (c *Client) Close(ctx context.Context) error {

	c.stopOnce.Do(func() {
		logger := logutil.GetLogger(ctx)
		logger.Info("close is called")

		// Notify init to stop trying to connect, and wait until it's done.
		close(c.stopInit)
		select {
		case <-c.initDone:
			logger.Debug("init is done")
		case <-ctx.Done():
			logger.WithError(ctx.Err()).Warn("init is not finished before this timeout")
			c.stopError = ctx.Err()
			return
		}

		if c.mongoClient == nil {
			c.stopError = nil
			return
		}

		logger.Debug("closing connection")
		c.stopError = c.mongoClient.Disconnect(ctx)
	})

	return c.stopError
}

func (c *Client) connect(mongoURL, mongoUser, mongoPwd, mongoAuthSource string) error {

	if mongoURL == "" {
		return errors.New("mongoURL can't be empty")
	}

	mongoFullURL := fmt.Sprintf("mongodb://%s", mongoURL)

	cred := options.Credential{
		AuthSource: mongoAuthSource,
		Username:   mongoUser,
		Password:   mongoPwd,
	}

	authOption := (&options.ClientOptions{}).SetAuth(cred)
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoFullURL), authOption)
	if err != nil {
		return err
	}

	if mongoClient == nil {
		return errors.New("nil mongoDB client")
	}

	c.mongoClient = mongoClient

	return nil
}
