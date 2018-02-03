package impl

import (
	"ocopea/mysqldsb/models"
	"ocopea/mysqldsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
)

const DSB_NAME = "mysql-k8s-dsb"

func DsbInfoResponse() middleware.Responder {

	return dsb_web.NewGetDSBInfoOK().WithPayload(
		&models.DsbInfo{
			Name: DSB_NAME,
			Description: "mysql DSB for kubernetes",
			Type:  "datasoruce",
			Plans: []*models.DsbPlan{
				{
					ID: "standard",
					Name: "standard",
					Description: "Standard mysql container",
					DsbSettings: nil,
					CopyProtocols: []*models.DsbSupportedCopyProtocol{
						{
							CopyProtocol: "ShpanRest",
							CopyProtocolVersion: "1.0",
						},
					},
					Protocols: []*models.DsbSupportedProtocol{
						{
							Protocol: "mysql",
						},
					},
				},
			},
		})

}
