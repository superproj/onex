# Id


id generator.


## Code


generate short str by unique uint64.


### Usage


```bash
go get -u github.com/superproj/onex/pkg/id
```

```go
import (
	"fmt"
  "github.com/superproj/onex/pkg/id"
)

func main() {
	fmt.Println(id.NewCode(1))
	fmt.Println(id.NewCode(2))
	fmt.Println(id.NewCode(3))
	fmt.Println(id.NewCode(4))

	fmt.Println(id.NewCode(
		1,
		id.WithCodeChars([]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}),
		id.WithCodeN1(9),
		id.WithCodeN2(3),
		id.WithCodeL(5),
		id.WithCodeSalt(99999),
	))
}
```


### Options


- `WithCodeChars` - code set, each char will generate from this set, default ['2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y']
- `WithCodeL` - code length
- `WithCodeN1` - n1 & chars.length coprime
- `WithCodeN2` - n2 & code.length coprime
- `WithCodeSalt` - code salt, same options and same uint64 id will generate same code, u can set different salt to generate new code


## Snowflake Id


generate snowflake id based on [sonyflake](https://github.com/sony/sonyflake).


### Usage


```go
import (
	"context"
	"fmt"

  "github.com/superproj/onex/pkg/id"
)

func main() {
	sf := id.NewSonyflake(
		id.WithSonyflakeMachineId(1),
	)
	if sf.Error != nil {
		fmt.Println(sf.Error)
		return
	}
	fmt.Println(sf.Id(context.Background()))
}
```


### Options


- `WithSonyflakeMachineId` - machine id
- `WithSonyflakeStartTime` - start time, do not modify after setting once, otherwise, u may get duplicate ids
