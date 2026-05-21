package mongo

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultPageSize    = 20
	schemaSampleSize   = 100
	connectTimeout     = 5 * time.Second
	disconnectTimeout  = 3 * time.Second
	operationTimeout   = 5 * time.Second
	aggregateTimeout   = 10 * time.Second
)

type Client struct {
	MongoClient *mongo.Client
}

type DBInfo struct {
	Name        string
	Collections []string
}

type QueryOptions struct {
	Filter string
	Sort   string
	Limit  int64
	Skip   int64
}

type SchemaField struct {
	Path          string
	Types         map[string]int
	TotalDocCount int
}

type IndexInfo struct {
	Name      string
	Key       string
	Unique    bool
	Sparse    bool
	ExtraInfo string
}

func Connect(uri string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Client{MongoClient: client}, nil
}

func (c *Client) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), disconnectTimeout)
	defer cancel()
	return c.MongoClient.Disconnect(ctx)
}

func (c *Client) ListDatabasesAndCollections(ctx context.Context) ([]DBInfo, error) {
	dbNames, err := c.MongoClient.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}

	sort.Strings(dbNames)

	var infos []DBInfo
	for _, name := range dbNames {
		collections, err := c.MongoClient.Database(name).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			continue
		}
		sort.Strings(collections)
		infos = append(infos, DBInfo{Name: name, Collections: collections})
	}

	return infos, nil
}

func (c *Client) FetchDocuments(ctx context.Context, dbName, collName string, opts QueryOptions) ([]bson.M, int64, error) {
	coll := c.MongoClient.Database(dbName).Collection(collName)

	filter, err := parseFilter(opts.Filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}

	findOpts := options.Find()
	if opts.Limit > 0 {
		findOpts.SetLimit(opts.Limit)
	} else {
		findOpts.SetLimit(defaultPageSize)
	}
	if opts.Skip > 0 {
		findOpts.SetSkip(opts.Skip)
	}

	sortDoc, err := parseSort(opts.Sort)
	if err != nil {
		return nil, 0, err
	}
	findOpts.SetSort(sortDoc)

	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("decode documents: %w", err)
	}

	if docs == nil {
		docs = []bson.M{}
	}

	return docs, total, nil
}

func (c *Client) UpdateDocument(ctx context.Context, dbName, collName string, id interface{}, doc bson.M) error {
	coll := c.MongoClient.Database(dbName).Collection(collName)
	_, err := coll.ReplaceOne(ctx, bson.M{"_id": id}, doc)
	return err
}

func (c *Client) DeleteDocument(ctx context.Context, dbName, collName string, id interface{}) error {
	coll := c.MongoClient.Database(dbName).Collection(collName)
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (c *Client) InsertDocument(ctx context.Context, dbName, collName string, doc bson.M) (interface{}, error) {
	coll := c.MongoClient.Database(dbName).Collection(collName)
	res, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return res.InsertedID, nil
}

func (c *Client) AnalyzeSchema(ctx context.Context, dbName, collName string) ([]SchemaField, error) {
	coll := c.MongoClient.Database(dbName).Collection(collName)

	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetLimit(schemaSampleSize))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	return AnalyzeSchemaDocs(docs), nil
}

func AnalyzeSchemaDocs(docs []bson.M) []SchemaField {
	totalDocs := len(docs)
	if totalDocs == 0 {
		return nil
	}

	fieldMap := make(map[string]*SchemaField)

	var traverse func(path string, val interface{})
	traverse = func(path string, val interface{}) {
		if val == nil {
			return
		}

		typeName := fmt.Sprintf("%T", val)
		switch v := val.(type) {
		case primitive.ObjectID:
			typeName = "ObjectID"
		case primitive.DateTime:
			typeName = "Date"
		case primitive.A:
			typeName = "Array"
			if len(v) > 0 {
				traverse(path+"[]", v[0])
			}
		case []interface{}:
			typeName = "Array"
			if len(v) > 0 {
				traverse(path+"[]", v[0])
			}
		case primitive.M:
			typeName = "Object"
			for k, inner := range v {
				subPath := k
				if path != "" {
					subPath = path + "." + k
				}
				traverse(subPath, inner)
			}
		case map[string]interface{}:
			typeName = "Object"
			for k, inner := range v {
				subPath := k
				if path != "" {
					subPath = path + "." + k
				}
				traverse(subPath, inner)
			}
		case string:
			typeName = "String"
		case int, int32, int64:
			typeName = "Number"
		case float32, float64:
			typeName = "Double"
		case bool:
			typeName = "Boolean"
		}

		if path == "" {
			return
		}

		sf, exists := fieldMap[path]
		if !exists {
			sf = &SchemaField{
				Path:          path,
				Types:         make(map[string]int),
				TotalDocCount: totalDocs,
			}
			fieldMap[path] = sf
		}
		sf.Types[typeName]++
	}

	for _, doc := range docs {
		traverse("", doc)
	}

	var fields []SchemaField
	for _, f := range fieldMap {
		fields = append(fields, *f)
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})

	return fields
}

func (c *Client) RunAggregation(ctx context.Context, dbName, collName string, pipelineJSON string) ([]bson.M, error) {
	coll := c.MongoClient.Database(dbName).Collection(collName)

	pipeline, err := parsePipeline(pipelineJSON)
	if err != nil {
		return nil, err
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	if docs == nil {
		docs = []bson.M{}
	}

	return docs, nil
}

func (c *Client) ListIndexes(ctx context.Context, dbName, collName string) ([]IndexInfo, error) {
	coll := c.MongoClient.Database(dbName).Collection(collName)

	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rawIndexes []bson.M
	if err = cursor.All(ctx, &rawIndexes); err != nil {
		return nil, err
	}

	var indexes []IndexInfo
	for _, idx := range rawIndexes {
		name, _ := idx["name"].(string)
		unique, _ := idx["unique"].(bool)
		sparse, _ := idx["sparse"].(bool)

		keyStr := formatIndexKey(idx["key"])

		extra := ""
		if v, ok := idx["v"]; ok {
			extra = fmt.Sprintf("v: %v", v)
		}

		indexes = append(indexes, IndexInfo{
			Name:      name,
			Key:       keyStr,
			Unique:    unique,
			Sparse:    sparse,
			ExtraInfo: extra,
		})
	}

	return indexes, nil
}

func parseFilter(raw string) (bson.M, error) {
	filter := bson.M{}
	if strings.TrimSpace(raw) == "" {
		return filter, nil
	}

	if err := bson.UnmarshalExtJSON([]byte(raw), true, &filter); err != nil {
		var wrapped interface{}
		if err2 := bson.UnmarshalExtJSON([]byte(fmt.Sprintf("{%s}", raw)), true, &wrapped); err2 == nil {
			if m, ok := wrapped.(bson.M); ok {
				return m, nil
			}
		}
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	return filter, nil
}

func parseSort(raw string) (interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return bson.D{{Key: "_id", Value: -1}}, nil
	}

	var sortDoc bson.D
	if err := bson.UnmarshalExtJSON([]byte(raw), true, &sortDoc); err != nil {
		if err2 := bson.UnmarshalExtJSON([]byte(fmt.Sprintf("{%s}", raw)), true, &sortDoc); err2 != nil {
			return nil, fmt.Errorf("invalid sort (expected e.g. {\"field\": 1}): %w", err)
		}
	}

	return sortDoc, nil
}

func parsePipeline(raw string) ([]bson.D, error) {
	var pipeline []bson.D
	if err := bson.UnmarshalExtJSON([]byte(raw), true, &pipeline); err == nil {
		return pipeline, nil
	}

	var stage bson.D
	if err := bson.UnmarshalExtJSON([]byte(raw), true, &stage); err == nil {
		return []bson.D{stage}, nil
	}

	wrapped := fmt.Sprintf("[%s]", strings.TrimSpace(raw))
	var wrapped_pipeline []bson.D
	if err := bson.UnmarshalExtJSON([]byte(wrapped), true, &wrapped_pipeline); err != nil {
		return nil, fmt.Errorf("invalid aggregation pipeline: %w", err)
	}

	return wrapped_pipeline, nil
}

func formatIndexKey(raw interface{}) string {
	v, ok := raw.(bson.M)
	if !ok {
		return ""
	}
	var parts []string
	for k, val := range v {
		parts = append(parts, fmt.Sprintf("%s: %v", k, val))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
