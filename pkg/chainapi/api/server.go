package api

import (
	"fmt"
	"os"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rumsystem/ip-cert/pkg/zerossl"
	localcrypto "github.com/rumsystem/keystore/pkg/crypto"
	"github.com/rumsystem/quorum/internal/pkg/cli"
	"github.com/rumsystem/quorum/internal/pkg/conn/p2p"
	rummiddleware "github.com/rumsystem/quorum/internal/pkg/middleware"
	"github.com/rumsystem/quorum/internal/pkg/options"
	"github.com/rumsystem/quorum/internal/pkg/utils"
	appapi "github.com/rumsystem/quorum/pkg/chainapi/appapi"
)

var quitch chan os.Signal

//StartAPIServer : Start local web server
func StartAPIServer(config cli.Config, signalch chan os.Signal, h *Handler, apph *appapi.Handler, node *p2p.Node, nodeopt *options.NodeOptions, ks localcrypto.Keystore, ethaddr string, isbootstrapnode bool) {
	quitch = signalch
	e := utils.NewEcho(config.IsDebug)
	customJWTConfig := appapi.CustomJWTConfig(nodeopt.JWTKey)
	e.Use(middleware.JWTWithConfig(customJWTConfig))
	e.Use(rummiddleware.OpaWithConfig(rummiddleware.OpaConfig{
		Skipper:   rummiddleware.LocalhostSkipper,
		Policy:    policyStr,
		Query:     "x = data.quorum.restapi.authz.allow", // FIXME: hardcode
		InputFunc: opaInputFunc,
	}))
	r := e.Group("/api")
	a := e.Group("/app/api")
	r.GET("/quit", quitapp)
	if isbootstrapnode == false {
		r.POST("/v1/group", h.CreateGroup())
		r.POST("/v1/group/join", h.JoinGroup())
		r.POST("/v1/group/leave", h.LeaveGroup)
		r.POST("/v1/group/clear", h.ClearGroupData)
		r.POST("/v1/group/content", h.PostToGroup)
		r.POST("/v1/group/profile", h.UpdateProfile)
		r.POST("/v1/network/peers", h.AddPeers)
		r.POST("/v1/network/relay", h.AddRelayServers)
		r.POST("/v1/group/chainconfig", h.MgrChainConfig)
		r.POST("/v1/group/producer", h.GroupProducer)
		r.POST("/v1/group/user", h.GroupUser)
		r.POST("/v1/group/announce", h.Announce)
		//r.POST("/v1/group/schema", h.Schema)
		r.POST("/v1/group/:group_id/startsync", h.StartSync)
		r.POST("/v1/group/appconfig", h.MgrAppConfig)
		r.GET("/v1/node", h.GetNodeInfo)
		r.GET("/v1/network", h.GetNetwork(&node.Host, node.Info, nodeopt, ethaddr))
		r.GET("/v1/network/stats", h.GetNetworkStatsSummary)
		r.GET("/v1/network/peers/ping", h.PingPeers(node))
		r.POST("/v1/psping", h.PSPingPeer(node))
		r.POST("/v1/ping", h.P2PPingPeer(node))
		r.GET("/v1/block/:group_id/:block_id", h.GetBlockById)
		r.GET("/v1/trx/:group_id/:trx_id", h.GetTrx)
		r.POST("/v1/trx/ack", h.PubQueueAck)
		r.GET("/v1/groups", h.GetGroups)
		r.GET("/v1/group/:group_id/trx/allowlist", h.GetChainTrxAllowList)
		r.GET("/v1/group/:group_id/trx/denylist", h.GetChainTrxDenyList)
		r.GET("/v1/group/:group_id/trx/auth/:trx_type", h.GetChainTrxAuthMode)
		r.GET("/v1/group/:group_id/producers", h.GetGroupProducers)
		r.GET("/v1/group/:group_id/announced/users", h.GetAnnouncedGroupUsers)
		r.GET("/v1/group/:group_id/announced/user/:sign_pubkey", h.GetAnnouncedGroupUser)
		r.GET("/v1/group/:group_id/announced/producers", h.GetAnnouncedGroupProducer)
		//r.GET("/v1/group/:group_id/app/schema", h.GetGroupAppSchema)
		r.GET("/v1/group/:group_id/appconfig/keylist", h.GetAppConfigKey)
		r.GET("/v1/group/:group_id/appconfig/:key", h.GetAppConfigItem)
		r.GET("/v1/group/:group_id/seed", h.GetGroupSeedHandler)
		r.GET("/v1/group/:group_id/pubqueue", h.GetPubQueue)
		a.POST("/v1/group/:group_id/content", apph.ContentByPeers)
		a.POST("/v1/token/refresh", apph.RefreshToken)
		a.POST("/v1/token/create", apph.CreateToken)

		r.POST("/v1/tools/pubkeytoaddr", h.PubkeyToEthaddr)

		r.POST("/v1/preview/relay/req", h.RequestRelay)
		r.GET("/v1/preview/relay", h.ListRelay)
		r.GET("/v1/preview/relay/:req_id/approve", h.ApproveRelay)
		r.DELETE("/v1/preview/relay/:relay_id", h.RemoveRelay)

		r.POST("/v1/keystore/signtx", h.SignTx)

		//for nodesdk
		r.POST("/v1/node/trx/:group_id", h.SendTrx)
		r.POST("/v1/node/groupctn/:group_id", h.GetContentNSdk)
		r.POST("/v1/node/getchaindata/:group_id", h.GetDataNSdk)

	} else {
		r.GET("/v1/node", h.GetBootstrapNodeInfo)
	}

	// get public ip
	pubIps := utils.GetPublicIPs(config.APIIPAddresses)
	if len(pubIps) >= 1 { // issue cert and start https server
		// NOTE: choose the first public ip, and ignore others
		privKeyPath, certPath, err := zerossl.IssueIPCert(config.CertDir, pubIps[0], config.ZeroAccessKey)
		if err != nil {
			e.Logger.Fatal(err)
		}
		e.Logger.Fatal(e.StartTLS(config.APIListenAddresses, certPath, privKeyPath))
	} else { // start http server
		e.Logger.Fatal(e.Start(config.APIListenAddresses))
	}
}

func quitapp(c echo.Context) (err error) {
	fmt.Println("/api/quit has been called, send Signal SIGTERM...")
	quitch <- syscall.SIGTERM
	return nil
}
