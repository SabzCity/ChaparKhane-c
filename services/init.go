/* For license and copyright information please see LEGAL file in ChaparKhane repository */
// Auto-generated, edits will be overwritten

package services

import (
	"crypto/rand"

	"../libgo/achaemenid"
	"../libgo/asanak.com"
	"../libgo/captcha"
	"../libgo/log"
)

var (
	phraseCaptchas = captcha.NewDefaultPhraseCaptchas()

	smsProvider       asanak.Asanak
	smsOTPSecurityKey = make([]byte, 32)
)

// Server store address location to server use by other part of app!
var server *achaemenid.Server

// Init use to register all available services to given achaemenid.
func Init(s *achaemenid.Server) {
	var err error

	server = s

	var asanakJSON = s.Assets.Secret.GetFile("asanak.com.json")
	if asanakJSON == nil {
		log.Fatal("Can't find 'asanak.com.json' file in 'secret' folder in top of repository")
	}
	smsProvider.Init(asanakJSON.Data)

	_, err = rand.Read(smsOTPSecurityKey)
	// Note that err == nil only if we read len(SecurityKey) bytes.
	if err != nil {
		log.Fatal(err)
	}

	// Authentication
	s.Services.RegisterService(&registerNewPersonService)
	s.Services.RegisterService(&changePersonPasswordService)
	s.Services.RegisterService(&blockPersonService)
	s.Services.RegisterService(&recoverPersonAccountService)
	s.Services.RegisterService(&revokePersonPublicKeyService)
	s.Services.RegisterService(&unblockPersonService)
	// s.Services.RegisterService(&)

	// PersonNumber
	s.Services.RegisterService(&registerPersonNumberService)
	s.Services.RegisterService(&getPersonNumberStatusService)
	s.Services.RegisterService(&getPersonNumberService)
	// s.Services.RegisterService(&)

	// UserAppsConnection
	s.Services.RegisterService(&authenticateAppConnectionService)
	s.Services.RegisterService(&getUserAppConnectionsIDService)
	s.Services.RegisterService(&getUserAppConnectionService)
	s.Services.RegisterService(&getUserAppGivenDelegateConnectionsIDService)
	s.Services.RegisterService(&getUserAppGottenDelegateConnectionsIDService)
	s.Services.RegisterService(&registerUserAppConnectionService)
	s.Services.RegisterService(&updateUserAppConnectionService)
	// s.Services.RegisterService(&)

	// PersonPublicKey
	s.Services.RegisterService(&approvePersonPublicKeyService)
	s.Services.RegisterService(&authenticatePersonPublicKeyService)

	//
	// s.Services.RegisterService(&blockOrgService)

	// ForeignDetail
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)

	//
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)

	//
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)

	// OrganizationStaff
	// s.Services.RegisterService(&approveOrgPositionByPersonService)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)

	// Product
	// s.Services.RegisterService(&approveProductAuctionByWarehouseService)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService(&)

	// Common Services
	s.Services.RegisterService(&getNewPhraseCaptchaService)
	s.Services.RegisterService(&getPhraseCaptchaAudioService)
	s.Services.RegisterService(&solvePhraseCaptchaService)
	s.Services.RegisterService(&sendOtpService)
}
