package snowflake

import (
	"github.com/disgoorg/snowflake/v2"
	"strconv"
	"time"
)

func GenerateNew() string {
	epoch, err := time.Parse(time.RFC3339Nano, "2019-08-28T07:17:29.593Z")

	if err != nil {
		panic(err)
	}

	id := snowflake.ID((time.Now().UnixMilli() - epoch.Unix()) << 22)

	return strconv.FormatUint(uint64(id), 10)
}
