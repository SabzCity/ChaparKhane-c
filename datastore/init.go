/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"../libgo/achaemenid"
	"../libgo/ganjine"
)

var (
	// Cluster store cluster data to use by services!
	cluster *ganjine.Cluster
	// Server store address location to server use by other part of app!
	server *achaemenid.Server
)

// Init must call in main file before use any methods!
func Init(s *achaemenid.Server, c *ganjine.Cluster) {
	server = s
	cluster = c
}
