// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"cosmosVideoApi/controllers"
	beego "github.com/stonemeta/beego/server/web"
)

func init() {

	beego.Router("/:message/:videoID",&controllers.VideoController{},"get:GetVideo")
	beego.Router("/senderVoucher",&controllers.VideoController{}, "post:SendVoucher")
	beego.Router("/settle", &controllers.VideoController{}, "post:Settlement")

}
