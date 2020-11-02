/* For license and copyright information please see LEGAL file in repository */

package main

import (
	"time"

	"./datastore"
	"./libgo/achaemenid"
	as "./libgo/achaemenid-services"
	er "./libgo/error"
	"./libgo/ganjine"
	gs "./libgo/ganjine-services"
	lang "./libgo/language"
	"./libgo/letsencrypt"
	"./libgo/log"
	ps "./services"
)

var (
	server  achaemenid.Server
	cluster ganjine.Cluster
)

func init() {
	var err *er.Error

	server.Manifest = achaemenid.Manifest{
		SocietyID:  1,
		DomainName: "sabz.city",
		Email:      "ict@sabz.city",
		Icon:       "",

		Organization: map[lang.Language]string{
			lang.EnglishLanguage: "Persia Society",
			lang.PersianLanguage: "جامعه پارس",
		},
		Name: map[lang.Language]string{
			lang.EnglishLanguage: "SabzCity",
			lang.PersianLanguage: "شهرسبز",
		},
		Description: map[lang.Language]string{
			lang.EnglishLanguage: "SabzCity Platform",
			lang.PersianLanguage: "پلتفرم شهرسبز",
		},
		TermsOfService: map[lang.Language]string{
			lang.EnglishLanguage: "https://sabz.city/terms?hl=en",
			lang.PersianLanguage: "https://sabz.city/terms?hl=fa",
		},
		Licence: map[lang.Language]string{
			lang.EnglishLanguage: "https://sabz.city/licence?hl=en",
			lang.PersianLanguage: "https://sabz.city/licence?hl=fa",
		},
		TAGS: []string{
			"Society", "Innovative", "Government", "Life", "Life Style",
			"جامعه", "ابتکاری", "حکومت", "زندگی", "سبک زندگی",
		},

		RequestedPermission: []uint32{},

		TechnicalInfo: achaemenid.TechnicalInfo{
			ShutdownDelay: 2 * time.Second, // 2 * time.Minute,

			MaxOpenConnection:     5000000,
			ConnectionIdleTimeout: 24 * time.Hour,

			GuestMaxConnections: 2000000,

			CPUCores: 1,                        // one core
			CPUSpeed: 1 * 1000 * 1000,          // 1 GHz
			RAM:      4 * 1024 * 1024 * 1024,   // 4 GB
			GPU:      0,                        // 0 Ghz
			Network:  100 * 1024 * 1024,        // 100 MB/S
			Storage:  100 * 1024 * 1024 * 1024, // 100 GB

			DistributeOutOfSociety: false,
			DataCentersClass:       5,
			MaxNodeNumber:          30,
			NodeFailureTimeOut:     6 * time.Hour,
		},
	}

	// Initialize server
	server.Init()

	// Register stream app layer protocols. Dev can remove below and register only needed protocols handlers.
	server.StreamProtocols.Init()

	// Register default Achaemenid services
	as.Init(&server)
	// Register platform defined custom service in ./services/ folder
	ps.Init(&server)

	if log.DevMode {
		devInit()
	} else {
		productionInit()
	}

	// Connect to other nodes or become first node!
	err = server.Nodes.Init(&server)
	if err != nil {
		log.Fatal(err)
	}

	cluster.Manifest = ganjine.Manifest{
		DataCentersClass: 0,
		TotalZones:       3,

		TransactionTimeOut: 500,
		NodeFailureTimeOut: 60,

		CachePercent: 20,
	}

	// Ganjine initialize
	err = cluster.Init(&server)
	if err != nil {
		log.Fatal(err)
	}
	// Register default Ganjine services
	gs.Init(&server, &cluster)
	// Initialize datastore
	datastore.Init(&server, &cluster)

	// Register some other services for Achaemenid
	server.Connections.GetConnByID = getConnectionsByID
	server.Connections.GetConnByUserIDThingID = getConnectionsByUserIDThingID
	server.Connections.SaveConn = saveConnection
}

func main() {
	server.Start()
}

func productionInit() {
	var err error

	// register networks.
	err = server.Networks.Init(&server)
	if err != nil {
		log.Fatal(err)
	}

	// Check LetsEncrypt Certificate
	err = letsencrypt.CheckByAchaemenid(&server)
	if err != nil {
		log.Fatal(err)
	}

	// Register some selectable networks. Check to add or delete networks.
	// log.Info("try to register TCP on port 80 to listen for HTTP protocol")
	// server.StreamProtocols.SetProtocolHandler(80, achaemenid.HTTPToHTTPSHandler)
	// err = achaemenid.MakeTCPNetwork(server, 80)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	log.Info("try to register TCP/TLS on port 443 to listen for HTTPs protocol")
	server.StreamProtocols.SetProtocolHandler(443, achaemenid.HTTPIncomeRequestHandler)
	err = achaemenid.MakeTCPTLSNetwork(&server, 443)
	if err != nil {
		log.Fatal(err)
	}

	// log.Info("try to register UDP on port 53 to listen for DNS protocol")
	// server.StreamProtocols.SetProtocolHandler(53, achaemenid.DNSIncomeRequestHandler)
	// err = achaemenid.MakeUDPNetwork(server, 53)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func devInit() {
	var err error

	log.Info("try to register TCP on port 8080 to listen for HTTP protocol in dev phase")
	server.StreamProtocols.SetProtocolHandler(8080, achaemenid.HTTPIncomeRequestHandler)
	err = achaemenid.MakeTCPNetwork(&server, 8080)
	if err != nil {
		log.Warn(err)
	}

	log.Info("try to register TCP/TLS on port 443 to listen for HTTPs protocol")
	server.StreamProtocols.SetProtocolHandler(443, achaemenid.HTTPIncomeRequestHandler)
	err = achaemenid.MakeTCPTLSNetwork(&server, 443)
	if err != nil {
		log.Fatal(err)
	}

	go server.Assets.ReLoadFromStorageByCLI()
}
