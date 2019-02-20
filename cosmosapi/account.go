package cosmosapi

import (
	"context"
)

func (c *Client) GetDatabaseAccount(ctx context.Context) (DatabaseAccount, error) {
	dbacc := DatabaseAccount{}
	_, err := c.get(ctx, "", &dbacc, nil)
	return dbacc, err
}

func (c *Client) getDatabaseAccountCustomURL(ctx context.Context, url string) (DatabaseAccount, error) {
	dbacc := DatabaseAccount{}
	_, err := c.rawMethod(ctx, "GET", url, &dbacc, nil, nil)
	return dbacc, err
}

type DatabaseAccount struct {
	Resource
	Addresses                string                  `json:"addresses"`
	Media                    string                  `json:"media"`
	QueryEngineConfiguration string                  `json:"queryEngineConfiguration"`
	ReadPolicy               ReadPolicy              `json:"readPolicy"`
	ReadableLocations        []DatabaseLocation      `json:"readableLocations"`
	WritableLocations        []DatabaseLocation      `json:"writableLocations"`
	SystemReplicationPolicy  SystemReplicationPolicy `json:"systemReplicationPolicy"`
	UserConsistencyPolicy    UserConsistencyPolicy   `json:"userConsistencyPolicy"`
	UserReplicationPolicy    UserReplicationPolicy   `json:"userReplicationPolicy"`
}

type UserReplicationPolicy struct {
	AsyncReplication  bool  `json:"asyncReplication"`
	MaxReplicasetSize int64 `json:"maxReplicasetSize"`
	MinReplicaSetSize int64 `json:"minReplicaSetSize"`
}

type DatabaseLocation struct {
	DatabaseAccountEndpoint string `json:"databaseAccountEndpoint"`
	Name                    string `json:"name"`
}

type UserConsistencyPolicy struct {
	DefaultConsistencyLevel string `json:"defaultConsistencyLevel"`
}

type SystemReplicationPolicy struct {
	MaxReplicasetSize int64 `json:"maxReplicasetSize"`
	MinReplicaSetSize int64 `json:"minReplicaSetSize"`
}

type ReadPolicy struct {
	PrimaryReadCoefficient   int64 `json:"primaryReadCoefficient"`
	SecondaryReadCoefficient int64 `json:"secondaryReadCoefficient"`
}
