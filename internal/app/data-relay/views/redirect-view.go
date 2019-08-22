package views

import (
	"bitbucket.org/project-alkemi/data-relay/internal/app/data-relay/util"
	"github.com/labstack/echo/v4"
	"net/http"
)

func RedirectView(c echo.Context) error {

	cc := c.(*util.EndpointContext)

	return cc.Context.Redirect(http.StatusFound, "/dashboard")
}