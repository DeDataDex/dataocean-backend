package routers

import (
	beego "github.com/stonemeta/beego/server/web"
	"github.com/stonemeta/beego/server/web/context/param"
)

func init() {

    beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"] = append(beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"] = append(beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"],
        beego.ControllerComments{
            Method: "GetVideoFromAr",
            Router: "/getVideoFromAr",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"] = append(beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"],
        beego.ControllerComments{
            Method: "GetIP",
            Router: "/ip",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"] = append(beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"],
        beego.ControllerComments{
            Method: "SendVideoToAr",
            Router: "/sendVideoToAr",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"] = append(beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"],
        beego.ControllerComments{
            Method: "SendVoucher",
            Router: "/sendVoucher",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"] = append(beego.GlobalControllerRouter["cosmosVideoApi/controllers:VideoController"],
        beego.ControllerComments{
            Method: "Settlement",
            Router: "/settle",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
