package primboard

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Invite represents the database entry for the tokens
type Invite struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Token string             `json:"token,omitempty" bson:"token,omitempty"`
	Until int64              `json:"until,omitempty" bson:"until,omitempty"`
	Used  bool               `json:"used,omitempty" bson:"used,omitempty"`
}

var inviteColName = "invite"

// Init inits a token and saves it into database.
// Default validity is 3 days
func (i *Invite) Init(db *mongo.Database, validity int) (*mongo.InsertOneResult, error) {
	col, ctx := GetColCtx(inviteColName, db, 30)
	if err := i.generateToken(); err != nil {
		return nil, err
	}
	token := i.Token
	// try to find token.
	err := i.FindToken(db)
	if err == nil {
		// if no error occured -> token was found -> generate new
		CloseContext()
		i.Init(db, validity)
	} else if err != mongo.ErrNoDocuments {
		// if no ErrNoDocuments error occured -> return
		CloseContext()
		return nil, err
	}
	// Token could not be found -> continue
	i.Token = token //prevent, that token becomes resetted after noDocErr
	i.Until = time.Now().Add(time.Hour * time.Duration(validity*24)).Unix()
	i.Used = false

	result, err := col.InsertOne(ctx, i)
	CloseContext()
	return result, err
}

// Invalidate verifies that the token is in the database and is valid.
// If the token is valid, it sets it's used state to used.
func (i *Invite) Invalidate(db *mongo.Database) error {
	col, ctx := GetColCtx(inviteColName, db, 30)
	// find token
	if err := i.FindToken(db); err != nil {
		CloseContext()
		return err
	}
	// verify
	if i.Used {
		// token has been used already
		CloseContext()
		return errors.New("token has been used already")
	} else if i.Until < time.Now().Unix() {
		// token has been expired
		CloseContext()
		return errors.New("token has been expired")
	}
	// invalidate
	i.Used = true
	filter := bson.M{"token": i.Token}
	update := bson.M{"$set": i}
	// options to return the updated document
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// execute update
	err := col.FindOneAndUpdate(ctx, filter, update, &options).Decode(&i)
	CloseContext()
	if err != nil {
		return err
	}
	return nil
}

// FindToken selects an Invite with the given token
func (i *Invite) FindToken(db *mongo.Database) error {
	col, ctx := GetColCtx(inviteColName, db, 30)
	filter := bson.M{"token": i.Token}
	err := col.FindOne(ctx, filter).Decode(&i)
	CloseContext()
	return err
}

// FindID selects an Invite with the given ID
func (i *Invite) FindID(db *mongo.Database) error {
	col, ctx := GetColCtx(inviteColName, db, 30)
	filter := bson.M{"_id": i.ID}
	err := col.FindOne(ctx, filter).Decode(&i)
	CloseContext()
	return err
}

// Revalidate sets the given token usage to false and adjusts the until date if
// needed
func (i *Invite) Revalidate(db *mongo.Database, validity int) error {
	col, ctx := GetColCtx(inviteColName, db, 30)
	// find token
	if err := i.FindToken(db); err != nil {
		CloseContext()
		return err
	}
	i.Used = false
	// adjust date if expired
	if i.Until < time.Now().Unix() {
		i.Until = time.Now().Add(time.Hour * time.Duration(validity*24)).Unix()
	}

	filter := bson.M{"token": i.Token}
	update := bson.M{"$set": i}
	// set update options
	after := options.After
	upsert := true
	options := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// execute update
	err := col.FindOneAndUpdate(ctx, filter, update, &options).Decode(&i)
	CloseContext()
	if err != nil {
		return err
	}
	return nil
}

// GenerateToken creates a crypto/rand based unique token
func (i *Invite) generateToken() error {
	token, err := generateRandomStringURLSafe(32)
	if err != nil {
		return err
	}
	i.Token = token
	return nil
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

// GenerateRandomStringURLSafe returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomStringURLSafe(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}
