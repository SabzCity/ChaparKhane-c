
/* For license and copyright information please see LEGAL file in repository */
// Auto-generated, edits will be overwritten

/*
Service Details:
	- Status : ServiceStatePreAlpha  >> https://en.wikipedia.org/wiki/Software_release_life_cycle
	- IssueDate : 16/06/2020 18:33:07 +0430
	- ExpireDate : 01/01/1970 03:30:00 +0330
	- ExpireInFavorOf :  ""
	- ExpireInFavorOfID : 0
	- TAGS : "Authentication"

Usage :
    // RegisterNewPersonReq is the request structure of RegisterNewPerson()
    const RegisterNewPersonReq = {
        "PhoneNumber":  0,   // uint64
        "PhoneOTP":     0,   // uint64
        "PasswordHash": [],  // [32]byte
        "CaptchaID":    [],  // [16]byte
    }
    // RegisterNewPersonRes is the response structure of RegisterNewPerson()
    const RegisterNewPersonRes = {
        "PersonID": [], // [16]byte
    }
    RegisterNewPerson(RegisterNewPersonReq)
        .then(res => {
            // Handle response
            console.log(res)
        })
        .catch(err => {
            // Handle error situation here
            console.log(err)
        })

Also you can use "async function (){ try{await}catch{} }" instead func().then().catch()
*/

// RegisterNewPerson register a new user in SabzCity platform.
async function RegisterNewPerson(req) {
    // TODO::: First validate req before send to apis server!

    // TODO::: Check QUIC protocol availability!

    const request = new Request('/apis?956555232', {
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
        // TODO::: Toast to GUI about no network connectivity!
        throw err
    }
}
