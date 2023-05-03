package snowflake

import (
	"github.com/disgoorg/snowflake/v2"
	"strconv"
	"time"
)

func GenerateNew() string {
	id := snowflake.New(time.Now())

	return strconv.FormatUint(uint64(id), 10)
}
