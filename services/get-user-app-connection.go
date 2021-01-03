/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	gp "../libgo/GP"
	ip "../libgo/IP"
	"../libgo/achaemenid"
	"../libgo/authorization"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getUserAppConnectionService = achaemenid.Service{
	ID:                2106700127,
	IssueDate:         1603802112,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get User App Connection",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"UserAppConnection",
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
	WriteTime        etime.Time
	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	Status           datastore.UserAppConnectionStatus
	Description      string // User custom text to identify connection easily.

	/* Connection data */
	ID     [32]byte `json:",string"`
	Weight achaemenid.Weight

	/* Peer data */
	// Peer Location
	GPAddr  gp.Addr  `json:",string"`
	IPAddr  ip.Addr  `json:",string"`
	ThingID [32]byte `json:",string"`
	// Peer Identifiers
	UserID           [32]byte `json:",string"`
	UserType         authorization.UserType
	DelegateUserID   [32]byte `json:",string"`
	DelegateUserType authorization.UserType

	/* Security data */
	PeerPublicKey [32]byte `json:",string"`
	AccessControl authorization.AccessControl

	// Metrics data
	LastUsage             etime.Time // Last use of this connection
	PacketPayloadSize     uint16     // Always must respect max frame size, so usually packets can't be more than 8192Byte!
	MaxBandwidth          uint64     // Peer must respect this, otherwise connection will terminate and GP go to black list!
	ServiceCallCount      uint64     // Count successful or unsuccessful request.
	BytesSent             uint64     // Counts the bytes of payload data sent.
	PacketsSent           uint64     // Counts packets sent.
	BytesReceived         uint64     // Counts the bytes of payload data Receive.
	PacketsReceived       uint64     // Counts packets Receive.
	FailedPacketsReceived uint64     // Counts failed packets receive for firewalling server from some attack types!
	FailedServiceCall     uint64     // Counts failed service call e.g. data validation failed, ...
}

func getUserAppConnection(st *achaemenid.Stream, req *getUserAppConnectionReq) (res *getUserAppConnectionRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var uac = datastore.UserAppConnection{
		ID: req.ID,
	}
	err = uac.GetLastByID()
	if err != nil {
		return
	}

	if st.Connection.UserID != uac.UserID && st.Connection.UserID != uac.DelegateUserID {
		err = authorization.ErrUserNotAllow
		return
	}

	res = &getUserAppConnectionRes{
		WriteTime:        uac.WriteTime,
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: uac.ID,
		Status:           datastore.UserAppConnectionUpdate,
		Description:      uac.Description,

		/* Connection data */
		ID:     uac.ID, // TODO::: Due to HTTP use ConnectionID to authenticate connections enable it now!!??
		Weight: uac.Weight,

		/* Peer data */
		// Peer Location
		GPAddr:  uac.GPAddr,
		IPAddr:  uac.IPAddr,
		ThingID: uac.ThingID,
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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
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

	res.WriteTime = etime.Time(syllab.GetInt64(buf, 0))
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	res.Status = datastore.UserAppConnectionStatus(syllab.GetUInt8(buf, 72))
	res.Description = syllab.UnsafeGetString(buf, 73)
	copy(res.ID[:], buf[81:])
	res.Weight = achaemenid.Weight(syllab.GetUInt8(buf, 113))

	copy(res.GPAddr[:], buf[114:])
	copy(res.IPAddr[:], buf[128:])
	copy(res.ThingID[:], buf[144:])
	copy(res.UserID[:], buf[176:])
	res.UserType = authorization.UserType(syllab.GetUInt8(buf, 208))
	copy(res.DelegateUserID[:], buf[209:])
	res.DelegateUserType = authorization.UserType(syllab.GetUInt8(buf, 241))

	copy(res.PeerPublicKey[:], buf[242:])
	res.AccessControl.SyllabDecoder(buf, 274)

	res.LastUsage = etime.Time(syllab.GetInt64(buf, 274+res.AccessControl.SyllabStackLen()))
	res.PacketPayloadSize = syllab.GetUInt16(buf, 282+res.AccessControl.SyllabStackLen())
	res.MaxBandwidth = syllab.GetUInt64(buf, 284+res.AccessControl.SyllabStackLen())
	res.ServiceCallCount = syllab.GetUInt64(buf, 292+res.AccessControl.SyllabStackLen())
	res.BytesSent = syllab.GetUInt64(buf, 300+res.AccessControl.SyllabStackLen())
	res.PacketsSent = syllab.GetUInt64(buf, 308+res.AccessControl.SyllabStackLen())
	res.BytesReceived = syllab.GetUInt64(buf, 316+res.AccessControl.SyllabStackLen())
	res.PacketsReceived = syllab.GetUInt64(buf, 324+res.AccessControl.SyllabStackLen())
	res.FailedPacketsReceived = syllab.GetUInt64(buf, 332+res.AccessControl.SyllabStackLen())
	res.FailedServiceCall = syllab.GetUInt64(buf, 340+res.AccessControl.SyllabStackLen())
	return
}

func (res *getUserAppConnectionRes) syllabEncoder(buf []byte) {
	// buf = make([]byte, res.syllabLen()+offset)
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	// var i, ln uint32 // len of strings, slices, maps, ...

	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	syllab.SetUInt8(buf, 72, uint8(res.Status))
	hsi = syllab.SetString(buf, res.Description, 73, hsi)
	copy(buf[81:], res.ID[:])
	syllab.SetUInt8(buf, 113, uint8(res.Weight))

	copy(buf[114:], res.GPAddr[:])
	copy(buf[128:], res.IPAddr[:])
	copy(buf[144:], res.ThingID[:])
	copy(buf[176:], res.UserID[:])
	syllab.SetUInt8(buf, 208, uint8(res.UserType))
	copy(buf[209:], res.DelegateUserID[:])
	syllab.SetUInt8(buf, 241, uint8(res.DelegateUserType))

	copy(buf[242:], res.PeerPublicKey[:])
	res.AccessControl.SyllabEncoder(buf, 274, hsi)

	syllab.SetInt64(buf, 274+res.AccessControl.SyllabStackLen(), int64(res.LastUsage))
	syllab.SetUInt16(buf, 282+res.AccessControl.SyllabStackLen(), res.PacketPayloadSize)
	syllab.SetUInt64(buf, 284+res.AccessControl.SyllabStackLen(), res.MaxBandwidth)
	syllab.SetUInt64(buf, 292+res.AccessControl.SyllabStackLen(), res.ServiceCallCount)
	syllab.SetUInt64(buf, 300+res.AccessControl.SyllabStackLen(), res.BytesSent)
	syllab.SetUInt64(buf, 308+res.AccessControl.SyllabStackLen(), res.PacketsSent)
	syllab.SetUInt64(buf, 316+res.AccessControl.SyllabStackLen(), res.BytesReceived)
	syllab.SetUInt64(buf, 324+res.AccessControl.SyllabStackLen(), res.PacketsReceived)
	syllab.SetUInt64(buf, 332+res.AccessControl.SyllabStackLen(), res.FailedPacketsReceived)
	syllab.SetUInt64(buf, 340+res.AccessControl.SyllabStackLen(), res.FailedServiceCall)
	return
}

func (res *getUserAppConnectionRes) syllabStackLen() (ln uint32) {
	return 348 + res.AccessControl.SyllabStackLen()
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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "WriteTime":
			var num int64
			num, err = decoder.DecodeInt64()
			res.WriteTime = etime.Time(num)
		case "AppInstanceID":
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
		case "UserConnectionID":
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
		case "Status":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Status = datastore.UserAppConnectionStatus(num)
		case "Description":
			res.Description, err = decoder.DecodeString()
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
		case "Weight":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Weight = achaemenid.Weight(num)
		case "GPAddr":
			err = decoder.DecodeByteArrayAsBase64(res.GPAddr[:])
		case "IPAddr":
			err = decoder.DecodeByteArrayAsBase64(res.IPAddr[:])
		case "ThingID":
			err = decoder.DecodeByteArrayAsBase64(res.ThingID[:])
		case "UserID":
			err = decoder.DecodeByteArrayAsBase64(res.UserID[:])
		case "UserType":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.UserType = authorization.UserType(num)
		case "DelegateUserID":
			err = decoder.DecodeByteArrayAsBase64(res.DelegateUserID[:])
		case "DelegateUserType":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.DelegateUserType = authorization.UserType(num)
		case "PeerPublicKey":
			err = decoder.DecodeByteArrayAsBase64(res.PeerPublicKey[:])

		case "AccessControl":
			err = res.AccessControl.JSONDecoder(decoder)

		case "LastUsage":
			var num int64
			num, err = decoder.DecodeInt64()
			res.LastUsage = etime.Time(num)
		case "PacketPayloadSize":
			res.PacketPayloadSize, err = decoder.DecodeUInt16()
		case "MaxBandwidth":
			res.MaxBandwidth, err = decoder.DecodeUInt64()
		case "ServiceCallCount":
			res.ServiceCallCount, err = decoder.DecodeUInt64()
		case "BytesSent":
			res.BytesSent, err = decoder.DecodeUInt64()
		case "PacketsSent":
			res.PacketsSent, err = decoder.DecodeUInt64()
		case "BytesReceived":
			res.BytesReceived, err = decoder.DecodeUInt64()
		case "PacketsReceived":
			res.PacketsReceived, err = decoder.DecodeUInt64()
		case "FailedPacketsReceived":
			res.FailedPacketsReceived, err = decoder.DecodeUInt64()
		case "FailedServiceCall":
			res.FailedServiceCall, err = decoder.DecodeUInt64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
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
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(res.Description)

	encoder.EncodeString(`","ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`","Weight":`)
	encoder.EncodeUInt8(uint8(res.Weight))

	encoder.EncodeString(`,"GPAddr":"`)
	encoder.EncodeByteSliceAsBase64(res.GPAddr[:])

	encoder.EncodeString(`","IPAddr":"`)
	encoder.EncodeByteSliceAsBase64(res.IPAddr[:])

	encoder.EncodeString(`","ThingID":"`)
	encoder.EncodeByteSliceAsBase64(res.ThingID[:])

	encoder.EncodeString(`","UserID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserID[:])

	encoder.EncodeString(`","UserType":`)
	encoder.EncodeUInt8(uint8(res.UserType))

	encoder.EncodeString(`,"DelegateUserID":"`)
	encoder.EncodeByteSliceAsBase64(res.DelegateUserID[:])

	encoder.EncodeString(`","DelegateUserType":`)
	encoder.EncodeUInt8(uint8(res.DelegateUserType))

	encoder.EncodeString(`,"PeerPublicKey":"`)
	encoder.EncodeByteSliceAsBase64(res.PeerPublicKey[:])

	encoder.EncodeString(`","AccessControl":`)
	res.AccessControl.JSONEncoder(encoder)

	encoder.EncodeString(`,"LastUsage":`)
	encoder.EncodeInt64(int64(res.LastUsage))

	encoder.EncodeString(`,"PacketPayloadSize":`)
	encoder.EncodeUInt64(uint64(res.PacketPayloadSize))

	encoder.EncodeString(`,"MaxBandwidth":`)
	encoder.EncodeUInt64(res.MaxBandwidth)

	encoder.EncodeString(`,"ServiceCallCount":`)
	encoder.EncodeUInt64(res.ServiceCallCount)

	encoder.EncodeString(`,"BytesSent":`)
	encoder.EncodeUInt64(res.BytesSent)

	encoder.EncodeString(`,"PacketsSent":`)
	encoder.EncodeUInt64(res.PacketsSent)

	encoder.EncodeString(`,"BytesReceived":`)
	encoder.EncodeUInt64(res.BytesReceived)

	encoder.EncodeString(`,"PacketsReceived":`)
	encoder.EncodeUInt64(res.PacketsReceived)

	encoder.EncodeString(`,"FailedPacketsReceived":`)
	encoder.EncodeUInt64(res.FailedPacketsReceived)

	encoder.EncodeString(`,"FailedServiceCall":`)
	encoder.EncodeUInt64(res.FailedServiceCall)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getUserAppConnectionRes) jsonLen() (ln int) {
	ln = len(res.Description)
	ln += res.AccessControl.JSONLen()
	ln += 1127
	return
}
