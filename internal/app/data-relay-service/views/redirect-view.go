package views

import (
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/util"
	"github.com/labstack/echo/v4"
	"net/http"
)

func RedirectView(c echo.Context) error {

	cc := c.(*util.EndpointContext)

	return cc.Context.Redirect(http.StatusFound, "/dashboard")
}
