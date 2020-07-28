/* For license and copyright information please see LEGAL file in ChaparKhane repository */
// Auto-generated, edits will be overwritten

package services

import (
	"../libgo/achaemenid"
	"../libgo/asanak.com"
	"../libgo/captcha"
	"../libgo/log"
)

var phraseCaptchas = captcha.NewDefaultPhraseCaptchas()

var smsProvider = &asanak.Asanak{}

// Init use to register all available services to given achaemenid.
func Init(s *achaemenid.Server) {
	var asanakJSON = s.Assets.Secret.GetFile("asanak.com.json")
	if asanakJSON == nil {
		log.Fatal("Can't find 'asanak.com.json' file in 'secret' folder in top of repository")
	}
	smsProvider.Init(asanakJSON.Data)

	// Authentication
	s.Services.RegisterService(&registerNewPersonService)
	s.Services.RegisterService(&getNewPhraseCaptchaService)
	s.Services.RegisterService(&getPhraseCaptchaAudioService)
	s.Services.RegisterService(&solvePhraseCaptchaService)
	s.Services.RegisterService(&sendOtpService)
	s.Services.RegisterService(&authenticatePersonPublicKeyService)
	s.Services.RegisterService(&approvePersonPublicKeyService)
	s.Services.RegisterService(&changePersonPasswordService)
	s.Services.RegisterService(&blockPersonService)
	s.Services.RegisterService(&recoverPersonAccountService)
	s.Services.RegisterService(&revokePersonPublicKeyService)
	s.Services.RegisterService(&unblockPersonService)
	// s.Services.RegisterService(&)
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService(&blockOrgService)

	//
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()

	// ForeignDetail
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()

	//
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()

	// OrganizationStaff
	// s.Services.RegisterService(&approveOrgPositionByPersonService)
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()

	// Product
	// s.Services.RegisterService(&approveProductAuctionByWarehouseService)
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()
	// s.Services.RegisterService()

	// Common Services
}
