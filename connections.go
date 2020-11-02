/* For license and copyright information please see LEGAL file in repository */

package main

import (
	"./datastore"
	"./libgo/achaemenid"
	er "./libgo/error"
)

// GetConnectionsByID returns available connection by given data
func getConnectionsByID(connID [32]byte) (conn *achaemenid.Connection) {
	var err *er.Error
	var uac = datastore.UserAppsConnection{
		ID: connID,
	}
	err = uac.GetLastByID()
	if err != nil {
		return
	}
	conn = &achaemenid.Connection{
		/* Connection data */
		Server: &server,
		ID:     uac.ID,
		State:  achaemenid.StateLoaded,
		Weight: uac.Weight,

		/* Peer data */
		// Peer Location
		SocietyID: uac.SocietyID,
		RouterID:  uac.RouterID,
		GPAddr:    uac.GPAddr,
		IPAddr:    uac.IPAddr,
		ThingID:   uac.ThingID,
		// Peer Identifiers
		UserID:           uac.UserID,
		UserType:         uac.UserType,
		DelegateUserID:   uac.DelegateUserID,
		DelegateUserType: uac.DelegateUserType,

		/* Security data */
		PeerPublicKey: uac.PeerPublicKey,
		// Cipher        crypto.Cipher
		AccessControl: uac.AccessControl,

		/* Metrics data */
		LastUsage:             uac.LastUsage,
		PacketPayloadSize:     uac.PacketPayloadSize,
		MaxBandwidth:          uac.MaxBandwidth,
		ServiceCallCount:      uac.ServiceCallCount,
		BytesSent:             uac.BytesSent,
		PacketsSent:           uac.PacketsSent,
		BytesReceived:         uac.BytesReceived,
		PacketsReceived:       uac.PacketsReceived,
		FailedPacketsReceived: uac.FailedPacketsReceived,
		FailedServiceCall:     uac.FailedServiceCall,
	}
	conn.StreamPool.Init()
	return
}

// getConnectionsByUserIDThingID returns available connection by given data
func getConnectionsByUserIDThingID(userID, thingID [32]byte) (conn *achaemenid.Connection) {
	var err error
	var uac = datastore.UserAppsConnection{
		UserID:  userID,
		ThingID: thingID,
	}
	err = uac.GetLastByUserIDThingID()
	if err != nil {
		return
	}
	conn = &achaemenid.Connection{
		/* Connection data */
		Server: &server,
		ID:     uac.ID,
		State:  achaemenid.StateLoaded,
		Weight: uac.Weight,

		/* Peer data */
		// Peer Location
		SocietyID: uac.SocietyID,
		RouterID:  uac.RouterID,
		GPAddr:    uac.GPAddr,
		IPAddr:    uac.IPAddr,
		ThingID:   uac.ThingID,
		// Peer Identifiers
		UserID:           uac.UserID,
		UserType:         uac.UserType,
		DelegateUserID:   uac.DelegateUserID,
		DelegateUserType: uac.DelegateUserType,

		/* Security data */
		PeerPublicKey: uac.PeerPublicKey,
		// Cipher        crypto.Cipher
		AccessControl: uac.AccessControl,

		/* Metrics data */
		LastUsage:             uac.LastUsage,
		PacketPayloadSize:     uac.PacketPayloadSize,
		MaxBandwidth:          uac.MaxBandwidth,
		ServiceCallCount:      uac.ServiceCallCount,
		BytesSent:             uac.BytesSent,
		PacketsSent:           uac.PacketsSent,
		BytesReceived:         uac.BytesReceived,
		PacketsReceived:       uac.PacketsReceived,
		FailedPacketsReceived: uac.FailedPacketsReceived,
		FailedServiceCall:     uac.FailedServiceCall,
	}
	conn.StreamPool.Init()
	return
}

// saveConnection get a connection and save it its data by platform rules
func saveConnection(conn *achaemenid.Connection) {
	var uac = datastore.UserAppsConnection{
		/* Unique data */
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: conn.ID,
		Status:           datastore.UserAppsConnectionUpdate,
		// Description:      conn.Description,

		/* Connection data */
		ID:     conn.ID,
		Weight: conn.Weight,

		/* Peer data */
		// Peer Location
		SocietyID: conn.SocietyID,
		RouterID:  conn.RouterID,
		GPAddr:    conn.GPAddr,
		IPAddr:    conn.IPAddr,
		ThingID:   conn.ThingID,
		// Peer Identifiers
		UserID:           conn.UserID,
		UserType:         conn.UserType,
		DelegateUserID:   conn.DelegateUserID,
		DelegateUserType: conn.DelegateUserType,

		/* Security data */
		PeerPublicKey: conn.PeerPublicKey,
		AccessControl: conn.AccessControl,

		// Metrics data
		LastUsage:             conn.LastUsage,
		PacketPayloadSize:     conn.PacketPayloadSize,
		MaxBandwidth:          conn.MaxBandwidth,
		ServiceCallCount:      conn.ServiceCallCount,
		BytesSent:             conn.BytesSent,
		PacketsSent:           conn.PacketsSent,
		BytesReceived:         conn.BytesReceived,
		PacketsReceived:       conn.PacketsReceived,
		FailedPacketsReceived: conn.FailedPacketsReceived,
		FailedServiceCall:     conn.FailedServiceCall,
	}
	var err *er.Error
	err = uac.Set()
	if err != nil {
		// TODO::: Handle error due to called by go keyword!!
	}
	uac.IndexID()
	if conn.State == achaemenid.StateNew {
		conn.State = achaemenid.StateLoaded
		if conn.UserType == achaemenid.UserTypeGuest {
			uac.IndexIDforUserTypeDaily()
		} else {
			uac.IndexIDforUserID()
		}
	}
	return
}
