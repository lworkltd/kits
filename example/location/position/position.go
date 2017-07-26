package position

import (
	"context"
	"github.com/lworkltd/kits/example/location/model"
	"github.com/lworkltd/kits/service/restful/code"
)

func GetCitizenPosition(ctx context.Context, id string) (model.Location, code.Error) {
	return model.GetRedisSession().GetCitizenLocation(id)
}
