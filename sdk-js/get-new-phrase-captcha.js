
/* For license and copyright information please see LEGAL file in repository */
// Auto-generated, edits will be overwritten

/*
Service Details :
	- Status : ServiceStatePreAlpha  >> https://en.wikipedia.org/wiki/Software_release_life_cycle
	- IssueDate : 24/06/2020 20:14:12 +0430
	- ExpireDate : 01/01/1970 03:30:00 +0330
	- ExpireInFavorOf :  ""
	- ExpireInFavorOfID : 0
	- TAGS : "Authentication"

Usage :
	// GetNewPhraseCaptchaReq is the request structure of GetNewPhraseCaptcha()
	const GetNewPhraseCaptchaReq = {
        "Language": 0, // uint8
		"ImageFormat": 0, // uint8
	}
	// GetNewPhraseCaptchaRes is the response structure of GetNewPhraseCaptcha()
	const GetNewPhraseCaptchaRes = {
    	"CaptchaID": [], // [16]byte
        "Image": [], // []byte
	}
	GetNewPhraseCaptcha(GetNewPhraseCaptchaReq)
		.then(res => {
			// Handle response
			console.log(res)
		})
		.catch(err => {
			// Handle error situation here
			console.log(err)
		})

Also you can use "async function (){ try{await func()}catch (err){} }" instead "func(req).then(res).catch(err)"
*/

// GetNewPhraseCaptcha returns new phrase captcha challenge that expire in 2 minute
async function GetNewPhraseCaptcha(req) {
    // TODO::: First validate req before send to apis server!

    // TODO::: Check Quic protocol availability!

    const request = new Request('/apis?2701809032', {
        method: "POST",
        // compress: true,
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(req)
    })

    try {
        let res = await fetch(request)

        switch (res.status) {
            case 200:
                const contentType = res.headers.get('content-type')
                switch (contentType) {
                    case 'application/json':
                        try {
                            return await res.json()
                        } catch (err) {
                            throw err
                        }
                    // case 'text/plain':
                    //     try {
                    //         return await res.text()
                    //     } catch (err) {
                    //         throw err
                    //     }
                    default:
                        throw new TypeError("Oops, we haven't got valid data type in response!")
                }
            case 201:
                return null
            case 400:
            case 500:
            default:
                // Almost not reachable code!
                throw res.text()
        }
    } catch (err) {
        // TODO::: new more check here for error type!
        // TODO::: Toast to GUI about no network connectivity!
        throw err
    }
}
