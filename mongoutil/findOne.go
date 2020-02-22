package mongoutil

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// FindOne wraps the FindOne and Decode method.
func (c *Client) FindOne(ctx context.Context, dbName, collectionName string, filter bson.D, output interface{}) error {

	collection, err := c.GetCollection(ctx, dbName, collectionName)
	if err != nil {
		return errors.Wrapf(err, "failed to get db %v collection %v", dbName, collectionName)
	}

	if filter == nil {
		filter = bson.D{}
	}
	resp := collection.FindOne(ctx, filter)
	if resp == nil {
		return errors.New("nil response")
	}

	if resp.Err() != nil {
		return resp.Err()
	}

	err = resp.Decode(output)
	return err
}
