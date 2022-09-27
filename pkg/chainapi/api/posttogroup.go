package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	rumerrors "github.com/rumsystem/quorum/internal/pkg/errors"
	"github.com/rumsystem/quorum/internal/pkg/utils"
	"github.com/rumsystem/quorum/pkg/chainapi/handlers"
	quorumpb "github.com/rumsystem/rumchaindata/pkg/pb"
)

// @Tags Groups
// @Summary PostToGroup
// @Description Post object to a group
// @Accept json
// @Produce json
// @Param data body quorumpb.Activity true "Activity object"
// @Success 200 {object} handlers.TrxResult
// @Router /api/v1/group/content [post]
func (h *Handler) PostToGroup(c echo.Context) (err error) {
	cc := c.(*utils.CustomContext)
	paramspb := new(quorumpb.Activity)
	if err := cc.BindAndValidate(paramspb); err != nil {
		return err
	}

	//var sudo bool
	sudo, err := strconv.ParseBool(c.Param("sudo"))
	if err != nil {
		return rumerrors.NewBadRequestError(err)
	}

	/*
		if c.QueryParams().Get("sudo") == "" {
			sudo = false
		} else {

		}
	*/

	res, err := handlers.PostToGroup(paramspb, sudo)
	if err != nil {
		return rumerrors.NewBadRequestError(err)
	}

	return c.JSON(http.StatusOK, res)
}
