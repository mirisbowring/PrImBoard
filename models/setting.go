package models

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

// Settings stores all user related configurations
type Settings struct {
	IPFSNodes []*IPFSNode `json:"ipfsNodes,omitempty" bson:"ipfsNodes,omitempty"`
}

// IPFSNode stores all relevant information about an additional node for uploads
type IPFSNode struct {
	Title       string `json:"title" bson:"title"`
	Username    string `json:"username" bson:"username"`
	Password    string `json:"password,omitempty" bson:"password,omitempty"`
	Address     string `json:"address,omitempty" bson:"address,omitempty"`
	IPFSAPIPort int    `json:"ipfsApiPort,omitempty" bson:"ipfsApiPort,omitempty"`
	IPFSAPIURL  string `json:"ipfsApiUrl,omitempty" bson:"ipfsApiUrl,omitempty"`
	IPFSGateway string `json:"ipfsGateway,omitempty" bson:"ipfsGateway,omitempty"`
}

// SettingsProject is a bson representation of the settings object
var SettingsProject = bson.M{
	"ipfsNodes": IPFSNodeProject,
}

// IPFSNodeProject is a bson representation of the ipfs-node setting object
var IPFSNodeProject = bson.M{
	"title":       1,
	"username":    1,
	"password":    1,
	"address":     1,
	"ipfsApiPort": 1,
	"ipfsApiUrl":  1,
	"ipfsGateway": 1,
}

// AddIPFSNode checks if the passed node struct is valid and adds it to the user
// func (u *User) AddIPFSNode(ipfs IPFSNode) error {
// 	// check if ipfs-node setting is valid
// 	if err := ipfs.Valid(); err != nil {
// 		return err
// 	}
// 	// check if setting exists
// 	if u.Settings == nil {
// 		u.Settings = new(Settings)
// 	}
// 	// append node
// 	u.Settings.IPFSNodes = append(u.Settings.IPFSNodes, &ipfs)
// 	return nil
// }

// Valid checks whether the ipfs-node settings is valid or not
func (s *IPFSNode) Valid() error {
	if s.Username == "" {
		return errors.New("The username should not be empty")
	} else if s.Password == "" {
		return errors.New("The password should not be empty")
	} else if s.Address == "" {
		return errors.New("The address should not be empty")
	} else if s.IPFSAPIPort == 0 {
		return errors.New("The IPFS API Port should not be empty")
	} else if s.IPFSAPIURL == "" {
		return errors.New("The IPFS API URL should not be empty")
	} else if s.IPFSGateway == "" {
		return errors.New("The IPFS Gateway should not be empty")
	}
	return nil
}
