/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"encoding/base64"
	"fmt"
	"testing"

	"../libgo/pehrest"
)

func TestUserAppConnection_GetLastByUserIDThingID(t *testing.T) {
	var err error
	var uac = UserAppConnection{
		UserID:  [32]byte{128},
		ThingID: [32]byte{255, 11, 43, 107, 15, 207, 188, 186, 64, 98, 28, 242, 146, 170, 95, 239, 65, 121, 200, 243, 16, 4, 188, 239, 98, 83, 222, 41, 185, 128, 185, 194},
	}

	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: uac.hashUserIDThingIDForID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	fmt.Println(base64.RawURLEncoding.EncodeToString(indexReq.IndexKey[:]))
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = pehrest.HashGetValues(indexReq)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(*indexRes)
}
