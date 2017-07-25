package position

import (
	"context"
	"github.com/lvhuat/kits/example/location/model"
	"github.com/lvhuat/kits/service/restful/code"
)

func GetCitizenPosition(ctx context.Context, id string) (model.Location, code.Error) {
	return model.GetRedisSession().GetCitizenLocation(id)
}
