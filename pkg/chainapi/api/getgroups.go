package api

import (
	"net/http"
	"sort"

	"encoding/base64"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/echo/v4"
	chain "github.com/rumsystem/quorum/internal/pkg/chainsdk/core"
)

type groupInfo struct {
	GroupId        string `json:"group_id" validate:"required,uuid4"`
	GroupName      string `json:"group_name" validate:"required"`
	OwnerPubKey    string `json:"owner_pubkey" validate:"required"`
	UserPubkey     string `json:"user_pubkey" validate:"required"`
	UserEthaddr    string `json:"user_eth_addr" validate:"required"`
	ConsensusType  string `json:"consensus_type" validate:"required"`
	EncryptionType string `json:"encryption_type" validate:"required"`
	CipherKey      string `json:"cipher_key" validate:"required"`
	AppKey         string `json:"app_key" validate:"required"`
	Epoch          int64  `json:"epoch" validate:"required"`
	LastUpdated    int64  `json:"last_updated" validate:"required"`
	GroupStatus    string `json:"group_status" validate:"required"`
}

type GroupInfoList struct {
	GroupInfos []*groupInfo `json:"groups"`
}

func (s *GroupInfoList) Len() int { return len(s.GroupInfos) }
func (s *GroupInfoList) Swap(i, j int) {
	s.GroupInfos[i], s.GroupInfos[j] = s.GroupInfos[j], s.GroupInfos[i]
}

func (s *GroupInfoList) Less(i, j int) bool {
	return s.GroupInfos[i].GroupName < s.GroupInfos[j].GroupName
}

// @Tags Groups
// @Summary GetGroups
// @Description Get all joined groups
// @Produce json
// @Success 200 {object} GroupInfoList
// @Router /api/v1/groups [get]
func (h *Handler) GetGroups(c echo.Context) (err error) {
	var groups []*groupInfo
	groupmgr := chain.GetGroupMgr()
	for _, value := range groupmgr.Groups {
		group := &groupInfo{}

		group.OwnerPubKey = value.Item.OwnerPubKey
		group.GroupId = value.Item.GroupId
		group.GroupName = value.Item.GroupName
		group.OwnerPubKey = value.Item.OwnerPubKey
		group.UserPubkey = value.Item.UserSignPubkey
		group.ConsensusType = value.Item.ConsenseType.String()
		group.EncryptionType = value.Item.EncryptType.String()
		group.CipherKey = value.Item.CipherKey
		group.AppKey = value.Item.AppKey
		group.LastUpdated = value.Item.LastUpdate

		//get chainInfo (lastUpdate, currEpoch)
		//TBD
		group.Epoch = -1 //value.Item.Epoch

		b, err := base64.RawURLEncoding.DecodeString(group.UserPubkey)
		if err != nil {
			//try libp2pkey
		} else {
			ethpubkey, err := ethcrypto.DecompressPubkey(b)
			//ethpubkey, err := ethcrypto.UnmarshalPubkey(b)
			if err == nil {
				ethaddr := ethcrypto.PubkeyToAddress(*ethpubkey)
				group.UserEthaddr = ethaddr.Hex()
			}
		}

		switch value.GetSyncerStatus() {
		case chain.CONSENSUS_SYNC:
			group.GroupStatus = "CONSENSUS_SYNCING"
		case chain.SYNCING_FORWARD:
			group.GroupStatus = "BLOCK_SYNCING"
		case chain.SYNC_FAILED:
			group.GroupStatus = "SYNC_FAILED"
		case chain.IDLE:
			group.GroupStatus = "IDLE"
		}

		groups = append(groups, group)
	}

	ret := GroupInfoList{groups}
	sort.Sort(&ret)
	return c.JSON(http.StatusOK, &ret)
}
