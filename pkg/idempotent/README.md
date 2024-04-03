# Idempotent


api idempotent tool based on redis lua script.


## Usage


```bash
go get -u github.com/superproj/onex/pkg/idempotent
```

```go
import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
  "github.com/superproj/onex/pkg/idempotent"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	i := idempotent.New(
		idempotent.WithRedis(client),
	)

	token := i.Token(context.Background())
	fmt.Println(i.Check(context.Background(), token))
	// true

	// second check will fail
	fmt.Println(i.Check(context.Background(), token))
	// false
}
```


## Options


- `WithRedis` - redis client, default 127.0.0.1:6379
- `WithPrefix` - cache key prefix, default idempotent
- `WithExpire` - key expire time, default 60 minute
