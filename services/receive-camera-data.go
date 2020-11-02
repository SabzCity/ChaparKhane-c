/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	lang "../libgo/language"
	"../libgo/srpc"
)

var receiveCameraDataService = achaemenid.Service{
	ID:                2887752942,
	CRUD:              authorization.CRUDNone,
	IssueDate:         1604307348,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Receive Camera Data",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "Suggest to send more than 30fps. It is rule to send data each 10sec, means send more than 300frame in each request",
	},
	TAGS: []string{
		"",
	},

	SRPCHandler: ReceiveCameraDataSRPC,
}

// ReceiveCameraDataSRPC is sRPC handler of ReceiveCameraData service.
func ReceiveCameraDataSRPC(st *achaemenid.Stream) {
	var req = &receiveCameraDataReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = receiveCameraData(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}
	st.OutcomePayload = make([]byte, 4)
}

// read more:
// https://www.svs-vistek.com/en/knowledgebase/svs-about-machine-vision.php?p=polarized_en-893
// http://www.nikondigital.org/articles/rgb_digital_camera_color.htm
type receiveCameraDataReq struct {
	// IR data??

	RedFramesPolarized0   []light
	RedFramesPolarized45  []light
	RedFramesPolarized90  []light
	RedFramesPolarized135 []light

	GreenFramesPolarized0   []light
	GreenFramesPolarized45  []light
	GreenFramesPolarized90  []light
	GreenFramesPolarized135 []light

	BlueFramesPolarized0   []light
	BlueFramesPolarized45  []light
	BlueFramesPolarized90  []light
	BlueFramesPolarized135 []light

	YellowFramesPolarized0   []light
	YellowFramesPolarized45  []light
	YellowFramesPolarized90  []light
	YellowFramesPolarized135 []light

	MonochromeFramesPolarized0   []light
	MonochromeFramesPolarized45  []light
	MonochromeFramesPolarized90  []light
	MonochromeFramesPolarized135 []light

	// UV data??
}

type light struct {
	intensity  []uint8
	wavelength []uint8
	phase      []uint8
	distance   []uint8 // depth of objects
}

func receiveCameraData(st *achaemenid.Stream, req *receiveCameraDataReq) (err *er.Error) {
	// TODO::: Authenticate & Authorizing request first by service policy.

	err = st.Authorize()
	if err != nil {
		return
	}

	// TODO::: Proccess req data to mine any usefull data e.g. person location, ...
	// ai.proccessIncomeData(req)

	// TODO::: Check org policy to save or not save raw data??
	// TODO::: Proccess req data to make visual video and save it to datastore || just store data and visualize when needed??

	return
}

/*
	Request Encoders & Decoders
*/

func (req *receiveCameraDataReq) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *receiveCameraDataReq) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (req *receiveCameraDataReq) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (req *receiveCameraDataReq) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *receiveCameraDataReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}
