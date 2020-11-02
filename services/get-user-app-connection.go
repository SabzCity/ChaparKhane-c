/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getUserAppConnectionService = achaemenid.Service{
	ID:                2106700127,
	URI:               "", // API services can set like "/apis?2106700127" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDRead,
	IssueDate:         1603802112,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get User App Connection",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"UserAppsConnection",
	},

	SRPCHandler: GetUserAppConnectionSRPC,
	HTTPHandler: GetUserAppConnectionHTTP,
}

// GetUserAppConnectionSRPC is sRPC handler of GetUserAppConnection service.
func GetUserAppConnectionSRPC(st *achaemenid.Stream) {
	var req = &getUserAppConnectionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getUserAppConnectionRes
	res, st.Err = getUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetUserAppConnectionHTTP is HTTP handler of GetUserAppConnection service.
func GetUserAppConnectionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getUserAppConnectionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getUserAppConnectionRes
	res, st.Err = getUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getUserAppConnectionReq struct {
	ID [32]byte `json:",string"`
}

type getUserAppConnectionRes struct {
	WriteTime        int64
	AppInstanceID    [32]byte `json:",string"` // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte `json:",string"` // Store to remember which user connection set||chanaged this record!
	Status           datastore.UserAppsConnectionStatus
	Description      string // User custom text to identify connection easily.

	/* Connection data */
	ID     [32]byte `json:",string"`
	Weight achaemenid.Weight

	/* Peer data */
	// Peer Location
	SocietyID uint32
	RouterID  uint32
	GPAddr    [14]byte `json:",string"`
	IPAddr    [16]byte `json:",string"`
	ThingID   [32]byte `json:",string"`
	// Peer Identifiers
	UserID           [32]byte `json:",string"`
	UserType         achaemenid.UserType
	DelegateUserID   [32]byte `json:",string"`
	DelegateUserType achaemenid.UserType

	/* Security data */
	PeerPublicKey [32]byte `json:",string"`
	AccessControl authorization.AccessControl

	// Metrics data
	LastUsage             int64  // Last use of this connection
	PacketPayloadSize     uint16 // Always must respect max frame size, so usually packets can't be more than 8192Byte!
	MaxBandwidth          uint64 // Peer must respect this, otherwise connection will terminate and GP go to black list!
	ServiceCallCount      uint64 // Count successful or unsuccessful request.
	BytesSent             uint64 // Counts the bytes of payload data sent.
	PacketsSent           uint64 // Counts packets sent.
	BytesReceived         uint64 // Counts the bytes of payload data Receive.
	PacketsReceived       uint64 // Counts packets Receive.
	FailedPacketsReceived uint64 // Counts failed packets receive for firewalling server from some attack types!
	FailedServiceCall     uint64 // Counts failed service call e.g. data validation failed, ...
}

func getUserAppConnection(st *achaemenid.Stream, req *getUserAppConnectionReq) (res *getUserAppConnectionRes, err *er.Error) {
	if st.Connection.UserType == achaemenid.UserTypeGuest {
		err = authorization.ErrAuthorizationUserNotAllow
		return
	}

	err = st.Authorize()
	if err != nil {
		return
	}

	var uac = datastore.UserAppsConnection{
		ID: req.ID,
	}
	err = uac.GetLastByID()
	if err != nil {
		return
	}

	if st.Connection.UserID != uac.UserID {
		err = authorization.ErrAuthorizationUserNotAllow
		return
	}

	res = &getUserAppConnectionRes{
		WriteTime:        uac.WriteTime,
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: uac.ID,
		Status:           datastore.UserAppsConnectionUpdate,
		Description:      uac.Description,

		/* Connection data */
		ID:     uac.ID,
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
		AccessControl: uac.AccessControl,

		// Metrics data
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

	return
}

/*
	Request Encoders & Decoders
*/

func (req *getUserAppConnectionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	return
}

func (req *getUserAppConnectionReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
}

func (req *getUserAppConnectionReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getUserAppConnectionReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getUserAppConnectionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getUserAppConnectionReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
			if err != nil {
				return
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getUserAppConnectionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getUserAppConnectionReq) jsonLen() (ln int) {
	ln = 52
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getUserAppConnectionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.WriteTime = syllab.GetInt64(buf, 0)
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	res.Status = datastore.UserAppsConnectionStatus(syllab.GetUInt8(buf, 72))
	res.Description = syllab.UnsafeGetString(buf, 73)
	copy(res.ID[:], buf[81:])
	res.Weight = achaemenid.Weight(syllab.GetUInt8(buf, 113))
	res.SocietyID = syllab.GetUInt32(buf, 114)
	res.RouterID = syllab.GetUInt32(buf, 118)
	copy(res.GPAddr[:], buf[122:])
	copy(res.IPAddr[:], buf[136:])
	copy(res.ThingID[:], buf[152:])
	copy(res.UserID[:], buf[184:])
	res.UserType = achaemenid.UserType(syllab.GetUInt8(buf, 216))
	copy(res.DelegateUserID[:], buf[217:])
	res.DelegateUserType = achaemenid.UserType(syllab.GetUInt8(buf, 249))
	copy(res.PeerPublicKey[:], buf[250:])
	res.AccessControl.SyllabDecoder(buf, 282)

	res.LastUsage = syllab.GetInt64(buf, 282+res.AccessControl.SyllabStackLen())
	res.PacketPayloadSize = syllab.GetUInt16(buf, 290+res.AccessControl.SyllabStackLen())
	res.MaxBandwidth = syllab.GetUInt64(buf, 292+res.AccessControl.SyllabStackLen())
	res.ServiceCallCount = syllab.GetUInt64(buf, 300+res.AccessControl.SyllabStackLen())
	res.BytesSent = syllab.GetUInt64(buf, 308+res.AccessControl.SyllabStackLen())
	res.PacketsSent = syllab.GetUInt64(buf, 316+res.AccessControl.SyllabStackLen())
	res.BytesReceived = syllab.GetUInt64(buf, 324+res.AccessControl.SyllabStackLen())
	res.PacketsReceived = syllab.GetUInt64(buf, 332+res.AccessControl.SyllabStackLen())
	res.FailedPacketsReceived = syllab.GetUInt64(buf, 340+res.AccessControl.SyllabStackLen())
	res.FailedServiceCall = syllab.GetUInt64(buf, 348+res.AccessControl.SyllabStackLen())
	return
}

func (res *getUserAppConnectionRes) syllabEncoder(buf []byte) {
	// buf = make([]byte, res.syllabLen()+offset)
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	// var i, ln uint32 // len of strings, slices, maps, ...

	syllab.SetInt64(buf, 0, res.WriteTime)
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	syllab.SetUInt8(buf, 72, uint8(res.Status))
	hsi = syllab.SetString(buf, res.Description, 73, hsi)
	copy(buf[81:], res.ID[:])
	syllab.SetUInt8(buf, 113, uint8(res.Weight))
	syllab.SetUInt32(buf, 114, res.SocietyID)
	syllab.SetUInt32(buf, 118, res.RouterID)
	copy(buf[122:], res.GPAddr[:])
	copy(buf[136:], res.IPAddr[:])
	copy(buf[152:], res.ThingID[:])
	copy(buf[184:], res.UserID[:])
	syllab.SetUInt8(buf, 216, uint8(res.UserType))
	copy(buf[217:], res.DelegateUserID[:])
	syllab.SetUInt8(buf, 249, uint8(res.DelegateUserType))
	copy(buf[250:], res.PeerPublicKey[:])
	res.AccessControl.SyllabEncoder(buf, 282, hsi)

	syllab.SetInt64(buf, 282+res.AccessControl.SyllabStackLen(), res.LastUsage)
	syllab.SetUInt16(buf, 290+res.AccessControl.SyllabStackLen(), res.PacketPayloadSize)
	syllab.SetUInt64(buf, 292+res.AccessControl.SyllabStackLen(), res.MaxBandwidth)
	syllab.SetUInt64(buf, 300+res.AccessControl.SyllabStackLen(), res.ServiceCallCount)
	syllab.SetUInt64(buf, 308+res.AccessControl.SyllabStackLen(), res.BytesSent)
	syllab.SetUInt64(buf, 316+res.AccessControl.SyllabStackLen(), res.PacketsSent)
	syllab.SetUInt64(buf, 324+res.AccessControl.SyllabStackLen(), res.BytesReceived)
	syllab.SetUInt64(buf, 332+res.AccessControl.SyllabStackLen(), res.PacketsReceived)
	syllab.SetUInt64(buf, 340+res.AccessControl.SyllabStackLen(), res.FailedPacketsReceived)
	syllab.SetUInt64(buf, 348+res.AccessControl.SyllabStackLen(), res.FailedServiceCall)
	return
}

func (res *getUserAppConnectionRes) syllabStackLen() (ln uint32) {
	return 356 + res.AccessControl.SyllabStackLen()
}

func (res *getUserAppConnectionRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.Description))
	ln += res.AccessControl.SyllabHeapLen()
	return
}

func (res *getUserAppConnectionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getUserAppConnectionRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	var keyName string
	for len(decoder.Buf) > 2 {
		keyName = decoder.DecodeKey()
		switch keyName {
		case "WriteTime":
			decoder.SetFounded()
			var num int64
			num, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
			res.WriteTime = int64(num)
		case "AppInstanceID":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
			if err != nil {
				return
			}
		case "UserConnectionID":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
			if err != nil {
				return
			}
		case "Status":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			res.Status = datastore.UserAppsConnectionStatus(num)
		case "Description":
			decoder.SetFounded()
			res.Description = decoder.DecodeString()
		case "ID":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
			if err != nil {
				return
			}
		case "Weight":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			res.Weight = achaemenid.Weight(num)
		case "SocietyID":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.SocietyID = uint32(num)
		case "RouterID":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.RouterID = uint32(num)
		case "GPAddr":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.GPAddr[:])
			if err != nil {
				return
			}
		case "IPAddr":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.IPAddr[:])
			if err != nil {
				return
			}
		case "ThingID":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.ThingID[:])
			if err != nil {
				return
			}
		case "UserID":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.UserID[:])
			if err != nil {
				return
			}
		case "UserType":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			res.UserType = achaemenid.UserType(num)
		case "DelegateUserID":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.DelegateUserID[:])
			if err != nil {
				return
			}
		case "DelegateUserType":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			res.DelegateUserType = achaemenid.UserType(num)
		case "PeerPublicKey":
			decoder.SetFounded()
			err = decoder.DecodeByteArrayAsBase64(res.PeerPublicKey[:])
			if err != nil {
				return
			}
		case "AccessControl":
			decoder.SetFounded()
			decoder.Buf, err = res.AccessControl.JSONDecoder(decoder.Buf)
			if err != nil {
				return
			}
		case "LastUsage":
			decoder.SetFounded()
			var num int64
			num, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
			res.LastUsage = int64(num)
		case "PacketPayloadSize":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.PacketPayloadSize = uint16(num)
		case "MaxBandwidth":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.MaxBandwidth = uint64(num)
		case "ServiceCallCount":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.ServiceCallCount = uint64(num)
		case "BytesSent":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.BytesSent = uint64(num)
		case "PacketsSent":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.PacketsSent = uint64(num)
		case "BytesReceived":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.BytesReceived = uint64(num)
		case "PacketsReceived":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.PacketsReceived = uint64(num)
		case "FailedPacketsReceived":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.FailedPacketsReceived = uint64(num)
		case "FailedServiceCall":
			decoder.SetFounded()
			res.FailedServiceCall, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *getUserAppConnectionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"WriteTime":`)
	encoder.EncodeInt64(int64(res.WriteTime))

	encoder.EncodeString(`,"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","Status":`)
	encoder.EncodeUInt64(uint64(res.Status))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(res.Description)

	encoder.EncodeString(`","ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`","Weight":`)
	encoder.EncodeUInt64(uint64(res.Weight))

	encoder.EncodeString(`,"SocietyID":`)
	encoder.EncodeUInt64(uint64(res.SocietyID))

	encoder.EncodeString(`,"RouterID":`)
	encoder.EncodeUInt64(uint64(res.RouterID))

	encoder.EncodeString(`,"GPAddr":"`)
	encoder.EncodeByteSliceAsBase64(res.GPAddr[:])

	encoder.EncodeString(`","IPAddr":"`)
	encoder.EncodeByteSliceAsBase64(res.IPAddr[:])

	encoder.EncodeString(`","ThingID":"`)
	encoder.EncodeByteSliceAsBase64(res.ThingID[:])

	encoder.EncodeString(`","UserID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserID[:])

	encoder.EncodeString(`","UserType":`)
	encoder.EncodeUInt64(uint64(res.UserType))

	encoder.EncodeString(`,"DelegateUserID":"`)
	encoder.EncodeByteSliceAsBase64(res.DelegateUserID[:])

	encoder.EncodeString(`","DelegateUserType":`)
	encoder.EncodeUInt64(uint64(res.DelegateUserType))

	encoder.EncodeString(`,"PeerPublicKey":"`)
	encoder.EncodeByteSliceAsBase64(res.PeerPublicKey[:])

	encoder.EncodeString(`","AccessControl":`)
	encoder.Buf = res.AccessControl.JSONEncoder(encoder.Buf)

	encoder.EncodeString(`,"LastUsage":`)
	encoder.EncodeInt64(int64(res.LastUsage))

	encoder.EncodeString(`,"PacketPayloadSize":`)
	encoder.EncodeUInt64(uint64(res.PacketPayloadSize))

	encoder.EncodeString(`,"MaxBandwidth":`)
	encoder.EncodeUInt64(uint64(res.MaxBandwidth))

	encoder.EncodeString(`,"ServiceCallCount":`)
	encoder.EncodeUInt64(uint64(res.ServiceCallCount))

	encoder.EncodeString(`,"BytesSent":`)
	encoder.EncodeUInt64(uint64(res.BytesSent))

	encoder.EncodeString(`,"PacketsSent":`)
	encoder.EncodeUInt64(uint64(res.PacketsSent))

	encoder.EncodeString(`,"BytesReceived":`)
	encoder.EncodeUInt64(uint64(res.BytesReceived))

	encoder.EncodeString(`,"PacketsReceived":`)
	encoder.EncodeUInt64(uint64(res.PacketsReceived))

	encoder.EncodeString(`,"FailedPacketsReceived":`)
	encoder.EncodeUInt64(uint64(res.FailedPacketsReceived))

	encoder.EncodeString(`,"FailedServiceCall":`)
	encoder.EncodeUInt64(uint64(res.FailedServiceCall))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getUserAppConnectionRes) jsonLen() (ln int) {
	ln += len(res.Description)
	ln += res.AccessControl.JSONLen()
	ln += 1127
	return
}
