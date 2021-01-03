/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"../libgo/ganjine"
)

func init() {
	ganjine.Cluster.DataStructures.RegisterDataStructure(&financialTransactionStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&organizationAuthenticationStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&personAuthenticationStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&personNumberStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&personPublicKeyStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&productAuctionStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&productPriceStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&productStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&quiddityStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&userAppConnectionStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&userNameStructure)
	ganjine.Cluster.DataStructures.RegisterDataStructure(&userPictureStructure)
	// ganjine.Cluster.DataStructures.RegisterDataStructure(&)
}
